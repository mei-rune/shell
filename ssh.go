package shell

import (
	"io"
	"os"
	"strings"

	"github.com/runner-mei/errors"
	"golang.org/x/crypto/ssh"
)

// SupportedCiphers xx
var SupportedCiphers = GetSupportedCiphers()
var SupportedKeyExchanges = GetKeyExchanges()

// GetSupportedCiphers xx
func GetSupportedCiphers() []string {
	config := &ssh.ClientConfig{}
	config.SetDefaults()

	for _, cipher := range []string{
		"aes128-cbc",
		"aes128-ctr",
		"aes192-ctr",
		"aes256-ctr",
		"aes128-gcm@openssh.com",
		"chacha20-poly1305@openssh.com",
		"arcfour256",
		"arcfour128",
		"arcfour",
		"3des-cbc",
	} {
		found := false
		for _, defaultCipher := range config.Ciphers {
			if cipher == defaultCipher {
				found = true
				break
			}
		}

		if !found {
			config.Ciphers = append(config.Ciphers, cipher)
		}
	}

	return config.Ciphers
}

func GetKeyExchanges() []string {
	config := &ssh.ClientConfig{}
	config.SetDefaults()

	for _, keyAlg := range []string{
		"diffie-hellman-group1-sha1",
		"diffie-hellman-group14-sha1",
		"ecdh-sha2-nistp256",
		"ecdh-sha2-nistp384",
		"ecdh-sha2-nistp521",
		"curve25519-sha256@libssh.org",
		"diffie-hellman-group-exchange-sha1",
		"diffie-hellman-group-exchange-sha256",
	} {
		found := false
		for _, defaultKeyAlg := range config.KeyExchanges {
			if keyAlg == defaultKeyAlg {
				found = true
				break
			}
		}

		if !found {
			config.KeyExchanges = append(config.KeyExchanges, keyAlg)
		}
	}

	return config.KeyExchanges
}

func init() {
	value := os.Getenv("ssh_key_exchanges")
	// if value == "" {
	// 	value = "diffie-hellman-group-exchange-sha256,diffie-hellman-group-exchange-sha1,diffie-hellman-group1-sha1,diffie-hellman-group14-sha1,ecdh-sha2-nistp256,ecdh-sha2-nistp384,ecdh-sha2-nistp521,curve25519-sha256@libssh.org"
	// }
	if value != "" {
		SupportedKeyExchanges = GetKeyExchanges()
		ss := strings.Split(value, ",")
		var newKeyExchanges []string
		for _, s := range ss {
			found := false
			for _, key := range SupportedKeyExchanges {
				if s == key {
					found = true
					break
				}
			}
			if found {
				newKeyExchanges = append(newKeyExchanges, s)
			}
		}
		for _, s := range SupportedKeyExchanges {
			found := false
			for _, key := range newKeyExchanges {
				if s == key {
					found = true
					break
				}
			}
			if !found {
				newKeyExchanges = append(newKeyExchanges, s)
			}
		}

		SupportedKeyExchanges = newKeyExchanges
	}
}

func ConnectSSH(host, user, password, privateKey string, sWriter, cWriter io.Writer) (Conn, error) {
	conn, err := DialSSH(host, user, password, privateKey)
	if err != nil {
		return nil, err
	}

	// Create a session
	session, err := conn.NewSession()
	if err != nil {
		conn.Close()

		return nil, errors.Wrap(err, "unable to create session")
	}
	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,      // disable echoing
		ssh.TTY_OP_ISPEED: 115200, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 115200, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := session.RequestPty("xterm", 800, 1600, modes); err != nil {
		conn.Close()
		session.Close()
		return nil, errors.Wrap(err, "request for pseudo terminal failed")
	}
	stdin, e := session.StdinPipe()
	if nil != e {
		conn.Close()
		session.Close()
		return nil, e
	}

	p := MakePipe(2048)
	if sWriter != nil {
		session.Stdout = MultWriters(p, sWriter)
	} else {
		session.Stdout = p
	}
	session.Stderr = session.Stdout

	// Start remote shell
	if err := session.Shell(); err != nil {
		conn.Close()
		session.Close()
		return nil, errors.Wrap(err, "failed to start shell")
	}

	go func() {
		err := session.Wait()
		if err != nil {
			p.CloseWithError(err)
		}
	}()

	if cWriter != nil {
		cWriter = MultWriters(stdin, cWriter)
	} else {
		cWriter = stdin
	}
	return &ConnWrapper{
		session:         conn,
		w:               cWriter,
		r:               p,
		readByte:        p,
		drainto:         p,
		setReadDeadline: p,
		// setWriteDeadline: p,
	}, nil
}

// DailSSH 连接到 ssh 服务
func DialSSH(host, user, password, privateKey string) (*ssh.Client, error) {
	var buffer strings.Builder
	interactiveCount := 0
	config := &ssh.ClientConfig{
		Config:          ssh.Config{Ciphers: SupportedCiphers, KeyExchanges: SupportedKeyExchanges},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            user,
		Auth: []ssh.AuthMethod{
			// ClientAuthPassword wraps a ClientPassword implementation
			// in a type that implements ClientAuth.
			ssh.Password(password),
			ssh.PasswordCallback(func() (string, error) {
				return password, nil
			}),

			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
				interactiveCount++
				if interactiveCount > 20 {
					return nil, errors.New("interactive count is too much")
				}
				if len(questions) == 0 {
					return []string{}, nil
				}
				for _, question := range questions {
					buffer.WriteString(question)
					switch strings.ToLower(strings.TrimSpace(question)) {
					case "password:", "password as":
						answers = append(answers, password)
						buffer.WriteString("******\r\n")
					default:
						answers = append(answers, "yes")
						buffer.WriteString("yes\r\n")
					}
				}
				return answers, nil
			}),
		},
	}

	if privateKey != "" {
		if password == "" {
			signer, err := ssh.ParsePrivateKey([]byte(privateKey))
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse private key")
			}

			config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		} else {
			signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(password))
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse private key")
			}
			config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		}
	}

	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		if buffer.Len() != 0 {
			if password == "" && privateKey == "" {
				return nil, errors.WrapWithSuffix(err, "可能是因为密码为空?\r\n"+buffer.String())
			}
			return nil, errors.WrapWithSuffix(err, buffer.String())
		}

		if password == "" && privateKey == "" {
			return nil, errors.WrapWithSuffix(err, "可能是因为密码为空?")
		}
		return nil, err
	}
	return conn, nil
}
