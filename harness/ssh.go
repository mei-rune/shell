package harness

import (
	"context"
	"io"
	"net"
	"os"
	"time"

	"github.com/mei-rune/shell"
	"github.com/runner-mei/errors"
)

const (
	DefaultReadTimeout  = 60 * time.Second
	DefaultWriteTimeout = 10 * time.Second
)

func JoinHostPort(addr, port string) string {
	if port == "" || port == "0" {
		return addr
	}
	return net.JoinHostPort(addr, port)
}

type SSHParam struct {
	Address             string `json:"address,omitempty" xml:"address,omitempty" form:"address,omitempty" query:"ssh.address,omitempty"`
	Port                string `json:"port,omitempty" xml:"port,omitempty" form:"port,omitempty" query:"ssh.port,omitempty"`
	UsernameQuest       string `json:"user_quest,omitempty" xml:"user_quest,omitempty" form:"user_quest,omitempty" query:"ssh.user_quest"`
	Username            string `json:"username,omitempty" xml:"username,omitempty" form:"username,omitempty" query:"ssh.user_name"`
	PasswordQuest       string `json:"password_quest,omitempty" xml:"password_quest,omitempty" form:"password_quest,omitempty" query:"ssh.password_quest"`
	Password            string `json:"password,omitempty" xml:"password,omitempty" form:"password,omitempty" query:"ssh.user_password,omitempty"`
	PrivateKey          string `json:"private_key,omitempty" xml:"private_key,omitempty" form:"private_key,omitempty" query:"ssh.private_key,omitempty"`
	Prompt              string `json:"prompt,omitempty" xml:"prompt,omitempty" form:"prompt,omitempty" query:"ssh.prompt,omitempty"`
	EnableCommand       string `json:"enable_command,omitempty" xml:"enable_command,omitempty" form:"enable_command,omitempty" query:"ssh.enable_command,omitempty"`
	EnablePasswordQuest string `json:"enable_password_quest,omitempty" xml:"enable_password_quest,omitempty" form:"enable_password_quest,omitempty" query:"ssh.enable_password_quest"`
	EnablePassword      string `json:"enable_password,omitempty" xml:"enable_password,omitempty" form:"enable_password,omitempty" query:"ssh.enable_password,omitempty"`
	EnablePrompt        string `json:"enable_prompt,omitempty" xml:"enable_prompt,omitempty" form:"enable_prompt,omitempty" query:"ssh.enable_prompt,omitempty"`
	UseExternalSSH      bool   `json:"use_external_ssh,omitempty" xml:"use_external_ssh,omitempty" form:"use_external_ssh,omitempty" query:"ssh.use_external_ssh,omitempty"`
	UseCRLF             bool   `json:"use_crlf,omitempty" xml:"use_crlf,omitempty" form:"use_crlf,omitempty" query:"ssh.use_crlf,omitempty"`

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (param *SSHParam) Host() string {
	if param.Port == "" || param.Port == "0" {
		return JoinHostPort(param.Address, "22")
	}
	return JoinHostPort(param.Address, param.Port)
}

var dumpSSH = false

func DailSSH(ctx context.Context, params *SSHParam, args ...Option) (shell.Conn, []byte, error) {
	var opts options
	for _, o := range args {
		o.apply(&opts)
	}

	if opts.questions == nil {
		opts.questions = noQuestions
	}

	if dumpSSH {
		sw := shell.WriteFunc(func(p []byte) (int, error) {
			io.WriteString(os.Stdout, "s:")
			io.WriteString(os.Stdout, shell.ToHexStringIfNeed(p))
			io.WriteString(os.Stdout, "\r\n")
			return len(p), nil
		})

		if opts.sWriter == nil {
			opts.sWriter = sw
		} else {
			opts.sWriter = io.MultiWriter(opts.sWriter, sw)
		}

		cw := shell.WriteFunc(func(p []byte) (int, error) {
			io.WriteString(os.Stdout, "c:")
			io.WriteString(os.Stdout, shell.ToHexStringIfNeed(p))
			io.WriteString(os.Stdout, "\r\n")
			return len(p), nil
		})

		if opts.cWriter == nil {
			opts.cWriter = cw
		} else {
			opts.cWriter = io.MultiWriter(opts.cWriter, cw)
		}
	}

	if params.ReadTimeout <= 0 {
		params.ReadTimeout = DefaultReadTimeout
	}
	if params.WriteTimeout <= 0 {
		params.WriteTimeout = DefaultWriteTimeout
	}

	if params.UseExternalSSH {
		c, err := shell.ConnectPlink(params.Host(), params.Username, params.Password, params.PrivateKey, opts.sWriter, opts.cWriter)
		if err != nil {
			return nil, nil, err
		}

		if params.UseCRLF {
			c.UseCRLF()
		}
		c.SetReadDeadline(params.ReadTimeout)
		c.SetWriteDeadline(params.WriteTimeout)

		if opts.skipLogin {
			return c, nil, nil
		}

		c1 := c.SetTeeReader(opts.inWriter)
		c2 := c.SetTeeWriter(opts.outWriter)

		defer func() {
			c1()
			c2()
		}()

		return sshLoginWithExternSSH(ctx, c, params, &opts)
	}

	c, err := shell.ConnectSSH(params.Host(), params.Username, params.Password, params.PrivateKey, opts.sWriter, opts.cWriter)
	if err != nil {
		return nil, nil, err
	}

	if params.UseCRLF {
		c.UseCRLF()
	}
	c.SetReadDeadline(params.ReadTimeout)
	c.SetWriteDeadline(params.WriteTimeout)

	if opts.skipLogin {
		return c, nil, nil
	}
	c1 := c.SetTeeReader(opts.inWriter)
	c2 := c.SetTeeWriter(opts.outWriter)

	defer func() {
		c1()
		c2()
	}()
	return sshLogin(ctx, c, params, &opts)
}

func sshLogin(ctx context.Context, c shell.Conn, params *SSHParam, opts *options) (shell.Conn, []byte, error) {
	var prompts [][]byte
	if params.Prompt != "" {
		prompts = [][]byte{[]byte(params.Prompt)}
	}

	if opts.skipEnable && opts.skipPrompt {
		return c, nil, nil
	}

	prompt, err := shell.ReadPrompt(ctx, c, prompts, opts.questions...)
	if err != nil {
		c.Close()
		return nil, nil, err
	}

	if opts.skipEnable {
		return c, prompt, nil
	}
	return sshEnableLogin(ctx, c, params, prompt, opts)
}

func sshLoginWithExternSSH(ctx context.Context, c shell.Conn, params *SSHParam, opts *options) (shell.Conn, []byte, error) {
	var prompts [][]byte
	if params.Prompt != "" {
		prompts = [][]byte{[]byte(params.Prompt)}
	}

	var userPrompts [][]byte
	if params.UsernameQuest != "" {
		userPrompts = [][]byte{[]byte(params.UsernameQuest)}
	}
	var passwordPrompts [][]byte
	if params.PasswordQuest != "" {
		passwordPrompts = [][]byte{[]byte(params.PasswordQuest)}
	}

	prompt, err := shell.UserLogin(ctx, c, userPrompts, []byte(params.Username), passwordPrompts, []byte(params.Password), prompts, opts.questions...)
	if err != nil {
		c.Close()
		return nil, nil, err
	}

	if opts.skipPrompt {
		c.Close()
		return nil, nil, errors.New("便用 UseExternalSSH 时不支持 skipPrompt 选项")
	}

	if opts.skipEnable {
		return c, prompt, nil
	}
	return sshEnableLogin(ctx, c, params, prompt, opts)
}

func sshEnableLogin(ctx context.Context, c shell.Conn, params *SSHParam, prompt []byte, opts *options) (shell.Conn, []byte, error) {
	var enablePasswordPrompts [][]byte
	if params.EnablePasswordQuest != "" {
		enablePasswordPrompts = [][]byte{[]byte(params.EnablePasswordQuest)}
	}
	var enablePrompts [][]byte
	if params.EnablePrompt != "" {
		enablePrompts = [][]byte{[]byte(params.EnablePrompt)}
	}

	if params.EnablePassword != "" {
		var err error

		prompt, err = shell.WithEnable(ctx, c, []byte(params.EnableCommand), enablePasswordPrompts, []byte(params.EnablePassword), enablePrompts)
		if err != nil {
			c.Close()
			return nil, nil, err
		}
	}

	return c, prompt, nil
}
