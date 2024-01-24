package harness

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/mei-rune/shell"
	"github.com/runner-mei/errors"
)

type DoFunc func(conn *Shell, idx int) (bool, error)

type ExecuteResult struct {
	LineNumber int    `json:"line_number"`
	LineText   string `json:"line_text,omitempty"`
	Command    string `json:"command,omitempty"`

	Incomming string `json:"incoming,omitempty"`
	Outgoing  string `json:"outgoing,omitempty"`

	SubResults []ExecuteResult `json:"sub_results,omitempty"`
}

type Shell struct {
	SSHParams    *SSHParam
	TelnetParams *TelnetParam
	SerialParams *SerialParam
	Variables    map[string]string

	IsSSHConn   bool
	Conn        shell.Conn
	Prompt      []byte
	promptStack [][]byte
	FailStrings [][]byte

	userCRLF  bool
	questions []shell.Matcher

	teeWriter io.Writer
	teeReader io.Writer

	opts []Option
}

func (s *Shell) WithOptions(opts ...Option) {
	s.opts = append(s.opts, opts...)
}

func (s *Shell) Close() error {
	if s.Conn == nil {
		return nil
	}
	return s.Conn.Close()
}

func (s *Shell) addFail(msg string) {
	if msg != "" {
		s.FailStrings = append(s.FailStrings, []byte(msg))
	}
}

func (s *Shell) OnFunc(question interface{}, answer shell.DoFunc) {
	s.questions = append(s.questions, shell.Match(question, answer))
}

func (s *Shell) AddQuestions(questions ...shell.Matcher) {
	s.questions = append(s.questions, questions...)
}

func (s *Shell) On(question interface{}, answer DoFunc) {
	cb := shell.DoFunc(func(conn shell.Conn, bs []byte, idx int) (bool, error) {
		if s.Conn == nil {
			s.Conn = conn // 可能是正在连接中
		}
		return answer(s, idx)
	})

	s.questions = append(s.questions, shell.Match(question, cb))
}

func (s *Shell) OnFail(question string) {
	s.questions = append(s.questions, shell.Match(question, func(conn shell.Conn, bs []byte, idx int) (bool, error) {
		return false, errors.New("收到错误消息: " + question)
	}))
}

func (s *Shell) pushPrompt() {
	s.promptStack = append(s.promptStack, s.Prompt)
}

func (s *Shell) popPrompt() error {
	if len(s.promptStack) == 0 {
		return errors.New("current isnot view mode")
	}
	s.Prompt = s.promptStack[len(s.promptStack)-1]
	s.promptStack = s.promptStack[:len(s.promptStack)-1]
	return nil
}

func (s *Shell) SetPrompt(prompt []byte) {
	s.Prompt = prompt
}

func (s *Shell) SetTeeWriter(w io.Writer) context.CancelFunc {
	if s.Conn != nil {
		return s.Conn.SetTeeWriter(w)
	}

	s.teeWriter = w
	return func() {
		s.teeWriter = nil
	}
}

func (s *Shell) SetTeeReader(w io.Writer) context.CancelFunc {
	if s.Conn != nil {
		return s.Conn.SetTeeReader(w)
	}
	s.teeReader = w
	return func() {
		s.teeReader = nil
	}
}

func (s *Shell) Connect(ctx context.Context, target string, opts ...Option) error {
	if target == "" || target == "auto" {
		if s.SSHParams != nil {
			target = "ssh"
		} else if s.TelnetParams != nil {
			target = "telnet"
		} else if s.SerialParams != nil {
			target = "serial"
		} else {
			return errors.New("没有 ssh 和 telnet 参数")
		}
	}

	if len(s.questions) > 0 {
		opts = append(opts, Questions(s.questions))
	}
	if len(s.opts) > 0 {
		opts = append(opts, s.opts...)
	}
	if s.teeWriter != nil {
		opts = append(opts, Outgoing(s.teeWriter))
	}
	if s.teeReader != nil {
		opts = append(opts, Incomming(s.teeReader))
	}

	switch target {
	case "ssh":
		if s.SSHParams == nil {
			return errors.New("没有 ssh 参数")
		}
		return s.connectSSH(ctx, opts...)
	case "telnet":
		if s.TelnetParams == nil {
			return errors.New("没有 telnet 参数")
		}
		return s.connectTelnet(ctx, opts...)
	case "serial":
		if s.SerialParams == nil {
			return errors.New("没有 serial 参数")
		}
		return s.connectSerial(ctx, opts...)

	default:
		return errors.New("不支持 '" + target + "' 参数")
	}
}

func (s *Shell) connectSerial(ctx context.Context, opts ...Option) error {
	if s.userCRLF {
		s.SerialParams.UseCRLF = true
	}

	conn, prompt, err := DailSerial(ctx, s.SerialParams, opts...)
	if err != nil {
		return err
	}
	s.IsSSHConn = false
	s.Conn = conn
	s.Prompt = prompt
	return nil
}

func (s *Shell) connectTelnet(ctx context.Context, opts ...Option) error {
	if s.userCRLF {
		s.TelnetParams.UseCRLF = true
	}

	conn, prompt, err := DailTelnet(ctx, s.TelnetParams, opts...)
	if err != nil {
		return err
	}
	s.IsSSHConn = false
	s.Conn = conn
	s.Prompt = prompt
	return nil
}

func (s *Shell) connectSSH(ctx context.Context, opts ...Option) error {
	if s.userCRLF {
		s.SSHParams.UseCRLF = true
	}
	conn, prompt, err := DailSSH(ctx, s.SSHParams, opts...)
	if err != nil {
		return err
	}
	s.IsSSHConn = true
	s.Conn = conn
	s.Prompt = prompt
	return nil
}

func (s *Shell) Login(ctx context.Context, args ...Option) error {
	if s.Conn == nil {
		return errors.New("无连接")
	}
	var opts options
	opts.questions = s.questions
	for _, o := range args {
		o.apply(&opts)
	}

	if opts.inWriter != nil {
		return errors.New("Login 不支持 Incoming 选项，请用 SetTeeReader() 替换")
	}
	if opts.outWriter != nil {
		return errors.New("Login 不支持 Outgoing 选项，请用 SetTeeWriter() 替换")
	}

	var prompt []byte
	var err error
	if s.IsSSHConn {
		if s.SSHParams.UseExternalSSH {
			_, prompt, err = sshLoginWithExternSSH(ctx, s.Conn, s.SSHParams, &opts)
		} else {
			_, prompt, err = sshLogin(ctx, s.Conn, s.SSHParams, &opts)
		}
	} else {
		_, prompt, err = telnetLogin(ctx, s.Conn, s.TelnetParams, &opts)
	}

	if err != nil {
		return err
	}

	s.SetPrompt(prompt)
	return nil
}

func (s *Shell) Enable(ctx context.Context, args ...Option) error {
	if s.Conn == nil {
		return errors.New("无连接")
	}
	var opts options
	opts.questions = s.questions
	for _, o := range args {
		o.apply(&opts)
	}

	if opts.inWriter != nil {
		return errors.New("Enable 不支持 Incoming 选项，请用 SetTeeReader() 替换")
	}
	if opts.outWriter != nil {
		return errors.New("Enable 不支持 Outgoing 选项，请用 SetTeeWriter() 替换")
	}

	var prompt []byte
	var err error
	if s.IsSSHConn {
		_, prompt, err = sshEnableLogin(ctx, s.Conn, s.SSHParams, s.Prompt, &opts)
	} else {
		_, prompt, err = telnetEnableLogin(ctx, s.Conn, s.TelnetParams, s.Prompt, &opts)
	}

	if err != nil {
		return err
	}

	s.SetPrompt(prompt)
	return nil
}

func (s *Shell) ReadPrompt(ctx context.Context, expected [][]byte) error {
	prompt, err := shell.ReadPrompt(ctx, s.Conn, expected, s.questions...)
	if err != nil {
		return err
	}
	s.SetPrompt(prompt)
	return nil
}

func (s *Shell) WithView(ctx context.Context, cmd []byte, newPrompts [][]byte) error {
	newPrompt, err := shell.WithView(ctx, s.Conn, cmd, newPrompts)
	if err != nil {
		return err
	}

	s.pushPrompt()
	s.SetPrompt(newPrompt)
	return nil
}

func (s *Shell) ExitView(ctx context.Context, cmd []byte) error {
	if err := s.popPrompt(); err != nil {
		return err
	}
	return s.exec(ctx, cmd)
}

func (s *Shell) Write(ctx context.Context, bs []byte) error {
	if s.Conn == nil {
		return errors.New("无连接")
	}

	return shell.WriteFull(s.Conn, bs)
}

func (s *Shell) Sendln(ctx context.Context, bs []byte) error {
	if s.Conn == nil {
		return errors.New("无连接")
	}

	return s.Conn.Sendln(bs)
}

func (s *Shell) RunScript(ctx context.Context, subScript *Script) ([]ExecuteResult, error) {
	return subScript.Run(ctx, s)
}

func (s *Shell) Exec(ctx context.Context, command string) error {
	return s.exec(ctx, []byte(command))
}

func (s *Shell) exec(ctx context.Context, command []byte) error {
	if s.Conn == nil {
		return errors.New("无连接")
	}

	_, err := s.Conn.DrainOff(0)
	if err != nil {
		if err == io.EOF {
			return errors.New("执行命令之前清空缓存失败: 连接断开")
		}
		return errors.Wrap(err, "执行命令之前清空缓存失败")
	}

	err = s.Conn.Sendln(command)
	if err != nil {
		return errors.Wrap(err, "发送命令失败")
	}

	err = shell.Expect(ctx, s.Conn, shell.Match(s.Prompt, shell.ReturnOK))
	if err != nil {
		return errors.Wrap(err, "执行命令时读提示符失败")
	}

	return nil
}

func Exec(ctx context.Context, s *Shell, command string) (*ExecuteResult, error) {
	var in strings.Builder
	var out strings.Builder

	c1 := s.Conn.SetTeeReader(&in)
	c2 := s.Conn.SetTeeWriter(&out)

	err := s.Exec(ctx, command)

	c1()
	c2()

	result := &ExecuteResult{
		LineText:  command,
		Command:   command,
		Incomming: in.String(),
		Outgoing:  out.String(),
	}

	if err == nil {
		bs := []byte(result.Incomming)
		for _, msg := range s.FailStrings {
			if bytes.Contains(bs, msg) {
				err = errors.New(result.Incomming)
				break
			}
		}
	}

	return result, err
}

type options struct {
	sWriter, cWriter    io.Writer
	inWriter, outWriter io.Writer

	skipPrompt bool
	skipLogin  bool
	skipEnable bool
	questions  []shell.Matcher

	// UserQuest           string
	// PasswordQuest       string
	// Prompt              string
	// EnablePasswordQuest string
	// EnablePrompt        string
}

type Option interface {
	apply(*options)
}

type optionFunc func(o *options)

func (f optionFunc) apply(o *options) {
	f(o)
}

func ServerWriter(w io.Writer) Option {
	return optionFunc(func(o *options) {
		if o.sWriter != nil {
			o.sWriter = shell.MultWriters(o.sWriter, w)
		} else {
			o.sWriter = w
		}
	})
}

func ClientWriter(w io.Writer) Option {
	return optionFunc(func(o *options) {
		if o.cWriter != nil {
			o.cWriter = shell.MultWriters(o.cWriter, w)
		} else {
			o.cWriter = w
		}
	})
}

func Incomming(w io.Writer) Option {
	return optionFunc(func(o *options) {
		if o.inWriter != nil {
			o.inWriter = shell.MultWriters(o.inWriter, w)
		} else {
			o.inWriter = w
		}
	})
}

func Outgoing(w io.Writer) Option {
	return optionFunc(func(o *options) {
		if o.outWriter != nil {
			o.outWriter = shell.MultWriters(o.outWriter, w)
		} else {
			o.outWriter = w
		}
	})
}

func SkipLogin(skip bool) Option {
	return optionFunc(func(o *options) {
		o.skipLogin = skip
	})
}

func SkipEnable(skip bool) Option {
	return optionFunc(func(o *options) {
		o.skipEnable = skip
	})
}

func SkipPrompt(skip bool) Option {
	return optionFunc(func(o *options) {
		o.skipPrompt = skip
	})
}

func Question(question interface{}, answer shell.DoFunc) Option {
	return optionFunc(func(o *options) {
		o.questions = append(o.questions, shell.Match(question, answer))
	})
}

func Questions(matchs []shell.Matcher) Option {
	return optionFunc(func(o *options) {
		if len(matchs) > 0 {
			o.questions = append(o.questions, matchs...)
		}
	})
}

var noQuestions = []shell.Matcher{}
