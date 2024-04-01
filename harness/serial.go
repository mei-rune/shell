package harness

import (
	"context"
	"io"
	"net"
	"os"
	"time"

	"github.com/mei-rune/shell"
	"github.com/tarm/serial"
)

type SerialParam struct {
	Port                string `json:"port,omitempty" xml:"port,omitempty" form:"port,omitempty" query:"serial.port,omitempty"`
	BaudRate            int    `json:"baud_rate,omitempty" xml:"baud_rate,omitempty" form:"baud_rate,omitempty" query:"serial.baud_rate,omitempty"`
	UsernameQuest       string `json:"user_quest,omitempty" xml:"user_quest,omitempty" form:"user_quest,omitempty" query:"serial.user_quest"`
	Username            string `json:"username,omitempty" xml:"username,omitempty" form:"username,omitempty" query:"serial.user_name"`
	PasswordQuest       string `json:"password_quest,omitempty" xml:"password_quest,omitempty" form:"password_quest,omitempty" query:"serial.password_quest"`
	Password            string `json:"password,omitempty" xml:"password,omitempty" form:"password,omitempty" query:"serial.user_password,omitempty"`
	Prompt              string `json:"prompt,omitempty" xml:"prompt,omitempty" form:"prompt,omitempty" query:"serial.prompt,omitempty"`
	EnableCommand       string `json:"enable_command,omitempty" xml:"enable_command,omitempty" form:"enable_command,omitempty" query:"serial.enable_command,omitempty"`
	EnablePasswordQuest string `json:"enable_password_quest,omitempty" xml:"enable_password_quest,omitempty" form:"enable_password_quest,omitempty" query:"serial.enable_password_quest"`
	EnablePassword      string `json:"enable_password,omitempty" xml:"enable_password,omitempty" form:"enable_password,omitempty" query:"serial.enable_password,omitempty"`
	EnablePrompt        string `json:"enable_prompt,omitempty" xml:"enable_prompt,omitempty" form:"enable_prompt,omitempty" query:"serial.enable_prompt,omitempty"`
	UseCRLF             bool   `json:"use_crlf,omitempty" xml:"use_crlf,omitempty" form:"use_crlf,omitempty" query:"serial.use_crlf,omitempty"`
}

func DailSerial(ctx context.Context, params *SerialParam, args ...Option) (shell.Conn, []byte, error) {
	var opts options
	for _, o := range args {
		o.apply(&opts)
	}
	if opts.questions == nil {
		opts.questions = noQuestions
	}

	cfg := &serial.Config{Name: params.Port, Baud: params.BaudRate, ReadTimeout: 5 * time.Second}
	serialConn, err := serial.OpenPort(cfg)
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

	c := shell.TelnetWrap(shell.NewTelnet(wrapSerial(serialConn)), opts.sWriter, opts.cWriter)
	if params.UseCRLF {
		c.UseCRLF()
	}
	c.SetReadDeadline(30 * time.Second)
	c.SetWriteDeadline(10 * time.Second)

	if opts.skipLogin {
		return c, nil, nil
	}
	c1 := c.SetTeeReader(opts.inWriter)
	c2 := c.SetTeeWriter(opts.outWriter)

	defer func() {
		c1()
		c2()
	}()

	_, err = c.Write([]byte("\n"))
	if err != nil {
		return nil, nil, err
	}

	return telnetLogin(ctx, c, &TelnetParam{
		UsernameQuest:       params.UsernameQuest,
		Username:            params.Username,
		PasswordQuest:       params.PasswordQuest,
		Password:            params.Password,
		Prompt:              params.Prompt,
		EnableCommand:       params.EnableCommand,
		EnablePasswordQuest: params.EnablePasswordQuest,
		EnablePassword:      params.EnablePassword,
		EnablePrompt:        params.EnablePrompt,
		UseCRLF:             params.UseCRLF,
	}, &opts)
}

func wrapSerial(port *serial.Port) net.Conn {
	return WrapPort{port}
}

type WrapPort struct {
	*serial.Port
}

func (wp WrapPort) LocalAddr() net.Addr {
	return nil
}

func (wp WrapPort) RemoteAddr() net.Addr {
	return nil
}

func (wp WrapPort) SetDeadline(time.Time) error {
	return nil
}

func (wp WrapPort) SetReadDeadline(time.Time) error {
	return nil
}

func (wp WrapPort) SetWriteDeadline(time.Time) error {
	return nil
}
