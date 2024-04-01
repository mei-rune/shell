package harness

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mei-rune/shell"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	eunicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func parsePassword(script *Script, line int, rawText string, copyed []byte) error {
	script.Cmds = append(script.Cmds,
		Command{
			LineNumber: line,
			LineText:   rawText,
			Run: func(ctx context.Context, script *Script, conn *Shell) error {
				if conn.IsSSHConn {
					if conn.SSHParams == nil {
						return errors.New("ssh 参数不存在")
					}

					return conn.Conn.SendPassword([]byte(conn.SSHParams.Password))
				}

				if conn.TelnetParams == nil {
					return errors.New("telnet 参数不存在")
				}

				return conn.Conn.SendPassword([]byte(conn.TelnetParams.Password))
			}})
	return nil
}

func parseEnablePassword(script *Script, line int, rawText string, copyed []byte) error {
	script.Cmds = append(script.Cmds,
		Command{
			LineNumber: line,
			LineText:   rawText,
			Run: func(ctx context.Context, script *Script, conn *Shell) error {
				if conn.IsSSHConn {
					if conn.SSHParams == nil {
						return errors.New("ssh 参数不存在")
					}

					return conn.Conn.SendPassword([]byte(conn.SSHParams.EnablePassword))
				}

				if conn.TelnetParams == nil {
					return errors.New("telnet 参数不存在")
				}

				return conn.Conn.SendPassword([]byte(conn.TelnetParams.EnablePassword))
			}})
	return nil
}

var Placeholders = []struct {
	Placeholder []byte
	Replace     func(*Shell, []byte) ([]byte, error)
}{
	{
		Placeholder: []byte("<<username>>"),
		Replace: func(conn *Shell, sendbuf []byte) ([]byte, error) {
			if conn.IsSSHConn {
				if conn.SSHParams == nil {
					return nil, errors.New("ssh 参数不存在")
				}

				return bytes.Replace(sendbuf,
					[]byte("<<username>>"),
					[]byte(conn.SSHParams.Username), -1), nil
			} else if conn.TelnetParams == nil {
				return nil, errors.New("telnet 参数不存在")
			} else {
				return bytes.Replace(sendbuf,
					[]byte("<<username>>"),
					[]byte(conn.TelnetParams.Username), -1), nil
			}
		},
	},
	{
		Placeholder: []byte("<<password>>"),
		Replace: func(conn *Shell, sendbuf []byte) ([]byte, error) {
			if conn.IsSSHConn {
				if conn.SSHParams == nil {
					return nil, errors.New("ssh 参数不存在")
				}

				return bytes.Replace(sendbuf,
					[]byte("<<password>>"),
					[]byte(conn.SSHParams.Password), -1), nil
			} else if conn.TelnetParams == nil {
				return nil, errors.New("telnet 参数不存在")
			} else {
				return bytes.Replace(sendbuf,
					[]byte("<<password>>"),
					[]byte(conn.TelnetParams.Password), -1), nil
			}
		},
	},

	{
		Placeholder: []byte("<<enable>>"),
		Replace: func(conn *Shell, sendbuf []byte) ([]byte, error) {
			if conn.IsSSHConn {
				if conn.SSHParams == nil {
					return nil, errors.New("ssh 参数不存在")
				}
				if len(conn.SSHParams.EnableCommand) == 0 {
					return bytes.Replace(sendbuf,
						[]byte("<<enable>>"),
						[]byte("enable"), -1), nil
				}

				return bytes.Replace(sendbuf,
					[]byte("<<enable>>"),
					[]byte(conn.SSHParams.EnableCommand), -1), nil
			} else if conn.TelnetParams == nil {
				return nil, errors.New("telnet 参数不存在")
			} else {
				if len(conn.TelnetParams.EnableCommand) == 0 {
					return bytes.Replace(sendbuf,
						[]byte("<<enable>>"),
						[]byte("enable"), -1), nil
				}
				return bytes.Replace(sendbuf,
					[]byte("<<enable>>"),
					[]byte(conn.TelnetParams.EnableCommand), -1), nil
			}
		},
	},
	{
		Placeholder: []byte("<<enable_password>>"),
		Replace: func(conn *Shell, sendbuf []byte) ([]byte, error) {
			if conn.IsSSHConn {
				if conn.SSHParams == nil {
					return nil, errors.New("ssh 参数不存在")
				}

				return bytes.Replace(sendbuf,
					[]byte("<<enable_password>>"),
					[]byte(conn.SSHParams.EnablePassword), -1), nil
			} else if conn.TelnetParams == nil {
				return nil, errors.New("telnet 参数不存在")
			} else {
				return bytes.Replace(sendbuf,
					[]byte("<<enable_password>>"),
					[]byte(conn.TelnetParams.EnablePassword), -1), nil
			}
		},
	},
}

func RegisterPlaceholder(name string) {
	tagName := "<<" + name + ">>"

	Placeholders = append(Placeholders, struct {
		Placeholder []byte
		Replace     func(*Shell, []byte) ([]byte, error)
	}{
		Placeholder: []byte(tagName),
		Replace: func(conn *Shell, sendbuf []byte) ([]byte, error) {
			if len(conn.Variables) == 0 {
				return sendbuf, nil
			}

			value, ok := conn.Variables[name]
			if !ok {
				return nil, errors.New("参数 '" + name + "' 不存在")
			}
			return bytes.Replace(sendbuf,
				[]byte(tagName),
				[]byte(value), -1), nil
		},
	})
}

var SubParsers = map[string]func(*Script, int, int, string, []byte, *Script) error{
	"@trigger": func(script *Script, start, end int, rawText string, fields []byte, subScript *Script) error {
		if len(fields) == 0 {
			return errors.New("缺少匹配参数")
		}

		types, charsets, ss, err := split([]rune(string(fields)))
		if err != nil {
			return errors.New("参数语法不正确: " + err.Error())
		}

		var words [][]byte
		var alreadyMore bool
		for idx := range types {
			switch types[idx] {
			case 0:
				if charsets[idx] == "" {
					words = append(words, []byte(string(ss[idx])))
				} else {
					bs, err := toBytes(ss[idx], charsets[idx])
					if err != nil {
						return errors.New("参数语法不正确: " + err.Error())
					}
					words = append(words, bs)
				}
			case 1:
				opt := string(ss[idx])
				switch strings.ToLower(opt) {
				case "alreadymore", "more":
					alreadyMore = true
				default:
					return errors.New("参数语法不正确, 选项 '" + opt + "' 是未知的")
				}
			}
		}

		if len(words) == 0 {
			return errors.New("参数语法不正确, 匹配字符没有")
		}

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: start,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					conn.On(words, DoFunc(func(conn *Shell, idx int) (bool, error) {
						results, err := conn.RunScript(ctx, subScript)
						if alreadyMore {
							if err == nil {
								return true, nil
							}
							return true, &scriptResult{
								results: results,
								err:     err,
							}
						}

						if err == nil {
							return false, nil
						}
						return false, &scriptResult{
							results: results,
							err:     err,
						}
					}))
					return nil
				}})
		return nil

	},
}

var Parsers = map[string]func(*Script, int, string, []byte) error{
	"@connect": func(script *Script, line int, rawText string, copyed []byte) error {
		var target string
		copyed = bytes.TrimSpace(copyed)
		a := bytes.Fields(copyed)

		var opts []Option
		for idx := range a {
			if bytes.EqualFold(a[idx], []byte("auto")) {
				target = "auto"
			} else if bytes.EqualFold(a[idx], []byte("ssh")) {
				target = "ssh"
			} else if bytes.EqualFold(a[idx], []byte("telnet")) {
				target = "telnet"
			} else if bytes.EqualFold(a[idx], []byte("skiplogin")) {
				opts = append(opts, SkipLogin(true))
			} else if bytes.EqualFold(a[idx], []byte("skipprompt")) {
				opts = append(opts, SkipPrompt(true))
			} else if bytes.EqualFold(a[idx], []byte("skipenable")) {
				opts = append(opts, SkipEnable(true))
				// } else if bytes.HasPrefix(a[idx], []byte("out:")) {
			} else {
				return errors.New("'" + string(a[idx]) + "' 是未知选项")
			}
		}

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					if conn.Conn != nil {
						return nil
					}
					return conn.Connect(ctx, target, opts...)
				}})
		return nil
	},
	"@login": func(script *Script, line int, rawText string, copyed []byte) error {
		copyed = bytes.TrimSpace(copyed)
		a := bytes.Fields(copyed)

		var opts []Option
		for idx := range a {
			if bytes.EqualFold(a[idx], []byte("skipenable")) {
				opts = append(opts, SkipEnable(true))
				// } else if bytes.HasPrefix(a[idx], []byte("out:")) {
			} else {
				return errors.New("'" + string(a[idx]) + "' 是未知选项")
			}
		}

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					return conn.Login(ctx, opts...)
				}})
		return nil
	},
	"@enable": func(script *Script, line int, rawText string, copyed []byte) error {
		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					return conn.Enable(ctx)
				}})
		return nil
	},
	"@write": func(script *Script, line int, rawText string, copyed []byte) error {
		copyed = bytes.TrimSpace(copyed)
		if bytes.Equal(copyed, []byte("<<password>>")) {
			return parsePassword(script, line, rawText, copyed)
		}

		if bytes.Equal(copyed, []byte("<<enable_password>>")) {
			return parseEnablePassword(script, line, rawText, copyed)
		}

		copyed = escapeBytes(copyed)

		var run CommandFunc = func(ctx context.Context, script *Script, conn *Shell) error {
			return conn.Write(ctx, copyed)
		}

		for _, a := range Placeholders {
			if bytes.Contains(copyed, a.Placeholder) {
				run = func(ctx context.Context, script *Script, conn *Shell) error {
					sendbuf := copyed

					var err error
					for _, a := range Placeholders {
						sendbuf, err = a.Replace(conn, sendbuf)
						if err != nil {
							return err
						}
					}
					return conn.Write(ctx, sendbuf) // shell.WriteFull(conn.Conn, sendbuf)
				}
				break
			}
		}

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run:        run,
			})
		return nil
	},
	"@send": func(script *Script, line int, rawText string, copyed []byte) error {
		copyed = bytes.TrimSpace(copyed)
		if bytes.Equal(copyed, []byte("<<password>>")) {
			return parsePassword(script, line, rawText, copyed)
		}

		if bytes.Equal(copyed, []byte("<<enable_password>>")) {
			return parseEnablePassword(script, line, rawText, copyed)
		}

		copyed = escapeBytes(copyed)

		var run CommandFunc = func(ctx context.Context, script *Script, conn *Shell) error {
			return conn.Sendln(ctx, copyed)
		}

		for _, a := range Placeholders {
			if bytes.Contains(copyed, a.Placeholder) {
				run = func(ctx context.Context, script *Script, conn *Shell) error {
					sendbuf := copyed

					var err error
					for _, a := range Placeholders {
						sendbuf, err = a.Replace(conn, sendbuf)
						if err != nil {
							return err
						}
					}
					return conn.Sendln(ctx, sendbuf)
				}
				break
			}
		}

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run:        run,
			})
		return nil
	},
	"@echo": func(script *Script, line int, rawText string, copyed []byte) error {
		copyed = escapeBytes(copyed)

		excepted := bytes.Split(copyed, []byte("$$$$$$$$"))
		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					return shell.Expect(ctx, conn.Conn, shell.Match(excepted, shell.ReturnOK))
				}})
		return nil
	},
	"@sleep": func(script *Script, line int, rawText string, copyed []byte) error {
		timeout := 1 * time.Second
		if len(copyed) > 0 {
			to, err := time.ParseDuration(string(copyed))
			if err != nil {
				return errors.New("timeout is invalid - " + string(copyed))
			}
			timeout = to
		}
		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					time.Sleep(timeout)
					return nil
				}})
		return nil
	},
	"@prompt": func(script *Script, line int, rawText string, copyed []byte) error {
		copyed = escapeBytes(copyed)

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					return conn.ReadPrompt(ctx, nil)
				}})
		return nil
	},
	"@exec": func(script *Script, line int, rawText string, copyed []byte) error {
		copyed = bytes.TrimSpace(copyed)
		if len(copyed) == 0 {
			return errors.New("命令不能为空")
		}
		command := string(copyed)

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Command:    command,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					return conn.Exec(ctx, command)
				}})
		return nil
	},
	"@password": func(script *Script, line int, rawText string, copyed []byte) error {
		if bytes.Equal(copyed, []byte("<<password>>")) {
			return parsePassword(script, line, rawText, copyed)
		}

		if bytes.Equal(copyed, []byte("<<enable_password>>")) {
			return parseEnablePassword(script, line, rawText, copyed)
		}

		if shell.IsEmptyPassword(copyed) {
			copyed = []byte{}
		}

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					return conn.Conn.SendPassword(copyed)
				}})
		return nil
	},

	"@drain": func(script *Script, line int, rawText string, copyed []byte) error {
		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					_, err := conn.Conn.DrainOff(0)
					return err
				}})
		return nil
	},

	"@@use_crlf": func(script *Script, line int, rawText string, copyed []byte) error {
		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					if conn.Conn != nil {
						conn.Conn.UseCRLF()
					}
					conn.userCRLF = true
					return nil
				}})
		return nil
	},
	"@@read_timeout": func(script *Script, line int, rawText string, copyed []byte) error {
		timeout := 10 * time.Second
		if len(copyed) > 0 {
			to, err := time.ParseDuration(string(copyed))
			if err != nil {
				return errors.New("read_timeout is invalid - " + string(copyed))
			}
			timeout = to
		}

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					conn.Conn.SetReadDeadline(timeout)
					return nil
				}})
		return nil
	},
	"@@fail": func(script *Script, line int, rawText string, copyed []byte) error {
		copyed = bytes.TrimSpace(copyed)
		if len(copyed) == 0 {
			return errors.New("@@fail 指令不正确, 错误不能为空")
		}
		failMsg := string(copyed)

		script.Cmds = append(script.Cmds,
			Command{
				LineNumber: line,
				LineText:   rawText,
				Run: func(ctx context.Context, script *Script, conn *Shell) error {
					conn.addFail(failMsg)
					//s.failed = append(s.failed, bytes.TrimSpace(copyed))
					return nil
				}})
		return nil
	},
}

var defaultParse = func(script *Script, line int, copyed []byte) error {
	return errors.New("unknown command error")
}

func toBytes(ss string, charset string) ([]byte, error) {
	switch strings.ToLower(charset) {
	case "", "utf8", "utf-8":
		return []byte(ss), nil
	case "gbk", "gb2312", "gb18030":
		bs, _, err := transform.Bytes(simplifiedchinese.GB18030.NewEncoder(), []byte(ss))
		return bs, err
	case "hz-gb2312":
		bs, _, err := transform.Bytes(simplifiedchinese.HZGB2312.NewEncoder(), []byte(ss))
		return bs, err
	case "big5":
		bs, _, err := transform.Bytes(traditionalchinese.Big5.NewEncoder(), []byte(ss))
		return bs, err
	case "utf16", "utf-16":
		bs, _, err := transform.Bytes(eunicode.UTF16(eunicode.BigEndian, eunicode.IgnoreBOM).NewEncoder(), []byte(ss))
		return bs, err
	case "utf16-be", "utf-16-be":
		bs, _, err := transform.Bytes(eunicode.UTF16(eunicode.BigEndian, eunicode.IgnoreBOM).NewEncoder(), []byte(ss))
		return bs, err
	case "utf16-le", "utf-16-le":
		bs, _, err := transform.Bytes(eunicode.UTF16(eunicode.LittleEndian, eunicode.IgnoreBOM).NewEncoder(), []byte(ss))
		return bs, err
	default:
		return nil, errors.New("charset '" + charset + "' is unknown")
	}
}
