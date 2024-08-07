package harness

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/mei-rune/shell"
	"github.com/runner-mei/errors"
)

type TelnetParam struct {
	Address             string `json:"address,omitempty" xml:"address,omitempty" form:"address,omitempty" query:"telnet.address,omitempty"`
	Port                string `json:"port,omitempty" xml:"port,omitempty" form:"port,omitempty" query:"telnet.port,omitempty"`
	UsernameQuest       string `json:"user_quest,omitempty" xml:"user_quest,omitempty" form:"user_quest,omitempty" query:"telnet.user_quest"`
	Username            string `json:"username,omitempty" xml:"username,omitempty" form:"username,omitempty" query:"telnet.user_name"`
	PasswordQuest       string `json:"password_quest,omitempty" xml:"password_quest,omitempty" form:"password_quest,omitempty" query:"telnet.password_quest"`
	Password            string `json:"password,omitempty" xml:"password,omitempty" form:"password,omitempty" query:"telnet.user_password,omitempty"`
	Prompt              string `json:"prompt,omitempty" xml:"prompt,omitempty" form:"prompt,omitempty" query:"telnet.prompt,omitempty"`
	EnableCommand       string `json:"enable_command,omitempty" xml:"enable_command,omitempty" form:"enable_command,omitempty" query:"telnet.enable_command,omitempty"`
	EnablePasswordQuest string `json:"enable_password_quest,omitempty" xml:"enable_password_quest,omitempty" form:"enable_password_quest,omitempty" query:"telnet.enable_password_quest"`
	EnablePassword      string `json:"enable_password,omitempty" xml:"enable_password,omitempty" form:"enable_password,omitempty" query:"telnet.enable_password,omitempty"`
	EnablePrompt        string `json:"enable_prompt,omitempty" xml:"enable_prompt,omitempty" form:"enable_prompt,omitempty" query:"telnet.enable_prompt,omitempty"`
	UseCRLF             bool   `json:"use_crlf,omitempty" xml:"use_crlf,omitempty" form:"use_crlf,omitempty" query:"telnet.use_crlf,omitempty"`

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (param *TelnetParam) Host() string {
	if param.Port == "" || param.Port == "0" {
		return JoinHostPort(param.Address, "23")
	}
	return JoinHostPort(param.Address, param.Port)
}

var dumpTelnet = false

func DailTelnet(ctx context.Context, params *TelnetParam, args ...Option) (shell.Conn, []byte, error) {
	var opts options
	for _, o := range args {
		o.apply(&opts)
	}
	if opts.questions == nil {
		opts.questions = noQuestions
	}

	telnetConn, err := shell.DialTelnetTimeout("tcp", params.Host(), 30*time.Second)
	if err != nil {
		return nil, nil, err
	}

	if dumpTelnet {
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

	c := shell.TelnetWrap(telnetConn, opts.sWriter, opts.cWriter)
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

	return telnetLogin(ctx, c, params, &opts)
}

func telnetLogin(ctx context.Context, c shell.Conn, params *TelnetParam, opts *options) (shell.Conn, []byte, error) {
	var prompts [][]byte
	if params.Prompt != "" {
		prompts = [][]byte{[]byte(params.Prompt)}
	}

	var err error
	var prompt []byte
	if shell.IsNonePassword([]byte(params.Password)) && shell.IsNoneUsername([]byte(params.Username)) {
		if !opts.skipPrompt {
			prompt, err = shell.ReadPrompt(ctx, c, prompts)
			if err != nil {
				c.Close()
				return nil, nil, err
			}
		}
	} else {
		var userPrompts [][]byte
		if params.UsernameQuest != "" {
			userPrompts = [][]byte{[]byte(params.UsernameQuest)}
		}
		var passwordPrompts [][]byte
		if params.PasswordQuest != "" {
			passwordPrompts = [][]byte{[]byte(params.PasswordQuest)}
		}

		prompt, err = shell.UserLogin(ctx, c, userPrompts, []byte(params.Username), passwordPrompts, []byte(params.Password), prompts, opts.questions...)
		if err != nil {
			c.Close()
			return nil, nil, err
		}

		if opts.skipPrompt {
			c.Close()
			return nil, nil, errors.New("便用 Telnet 时不支持 skipPrompt 选项")
		}
	}

	if opts.skipEnable {
		return c, prompt, nil
	}

	return telnetEnableLogin(ctx, c, params, prompt, opts)
}

func telnetEnableLogin(ctx context.Context, c shell.Conn, params *TelnetParam, prompt []byte, opts *options) (shell.Conn, []byte, error) {
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
