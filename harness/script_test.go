package harness

import (
	"context"
	"strings"
	"testing"

	// "github.com/mei-rune/shell/sim/sshd"
	// "github.com/mei-rune/shell/sim/telnetd"
	"tech.hengwei.com.cn/go/private/sim/sshd"
	"tech.hengwei.com.cn/go/private/sim/telnetd"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func TestScriptSimpleSSH(t *testing.T) {
	for _, test := range []struct {
		skipNotExternal bool
		skipExternal    bool
		testname        string
		scriptText      string
		expected        []ExecuteResult
	}{
		{
			skipNotExternal: false,
			skipExternal:    true,
			testname:        "test raw",
			scriptText: `
			@connect skipprompt skipenable
			@prompt
			@send <<enable>>
			@echo password:
			@send <<enable_password>>
			@prompt
			@exec echo abcd
	  		`,
			expected: []ExecuteResult{
				{LineNumber: 2, LineText: "@connect skipprompt skipenable"},
				{LineNumber: 3, LineText: "@prompt", Incomming: "ABC>"},
				{LineNumber: 4, LineText: "@send <<enable>>", Outgoing: "enable\r\n"},
				{LineNumber: 5, LineText: "@echo password:", Incomming: "enable\r\npassword:"},
				{LineNumber: 6, LineText: "@send <<enable_password>>", Outgoing: "********\r\n"},
				{LineNumber: 7, LineText: "@prompt", Incomming: "\r\nenable OK\r\nabc#"},
				{LineNumber: 8, LineText: "@exec echo abcd", Command: "echo abcd", Incomming: "echo abcd\r\nprint abcd\r\nabc#", Outgoing: "echo abcd\r\n"},
			},
		},

		{
			skipNotExternal: true,
			skipExternal:    false,
			testname:        "test raw",
			scriptText: `
			@connect skipenable
			@send <<enable>>
			@echo password:
			@send <<enable_password>>
			@prompt
			@exec echo abcd
	  		`,
			expected: []ExecuteResult{
				{LineNumber: 2, LineText: "@connect skipprompt skipenable", Incomming: "ABC>"},
				{LineNumber: 3, LineText: "@send <<enable>>", Outgoing: "enable\r\n"},
				{LineNumber: 4, LineText: "@echo password:", Incomming: "enable\r\npassword:"},
				{LineNumber: 5, LineText: "@send <<enable_password>>", Outgoing: "********\r\n"},
				{LineNumber: 6, LineText: "@prompt", Incomming: "\r\nenable OK\r\nabc#"},
				{LineNumber: 7, LineText: "@exec echo abcd", Command: "echo abcd", Incomming: "echo abcd\r\nprint abcd\r\nabc#", Outgoing: "echo abcd\r\n"},
			},
		},

		{
			testname: "test auto",
			scriptText: `
			@connect auto
			@exec echo abcd
			`,
			expected: []ExecuteResult{
				{
					LineNumber: 2,
					LineText:   "@connect auto",
					Incomming:  "ABC>enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:   "enable\r\n********\r\n",
				},
				{
					LineNumber: 3,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},
		{
			testname: "test skiplogin",
			scriptText: `
			@connect skiplogin
			@login
			@exec echo abcd
		    `,
			expected: []ExecuteResult{
				{
					LineNumber: 2,
					LineText:   "@connect skiplogin",
				},
				{
					LineNumber: 3,
					LineText:   "@login",
					Incomming:  "ABC>enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:   "enable\r\n********\r\n",
				},
				{
					LineNumber: 4,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},
		{
			testname: "test skipenable",
			scriptText: `
			@connect skiplogin
			@login skipenable
			@enable
			@exec echo abcd
		    `,
			expected: []ExecuteResult{
				{
					LineNumber: 2,
					LineText:   "@connect skiplogin",
				},
				{
					LineNumber: 3,
					LineText:   "@login skipenable",
					Incomming:  "ABC>",
				},
				{
					LineNumber: 4,
					LineText:   "@enable",
					Incomming:  "enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:   "enable\r\n********\r\n",
				},
				{
					LineNumber: 5,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},
	} {
		t.Run(test.testname, func(t *testing.T) {
			script, err := ParseScript(strings.NewReader(test.scriptText))
			if err != nil {
				t.Error(err)
				return
			}

			options := &sshd.Options{}
			options.AddUserPassword("abc", "123")

			options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
			//options.WithNoEnable("ABC>", sshd.Echo)

			listener, err := sshd.StartServer(":", options)
			if err != nil {
				t.Error(err)
				return
			}
			defer listener.Close()

			port := listener.Port()
			ctx := context.Background()

			params := &SSHParam{
				// Timeout: 30 * time.Second,
				Address: "127.0.0.1",
				Port:    port,
				// UserQuest: "",
				Username: "abc",
				// PasswordQuest: "",
				Password:            "123",
				PrivateKey:          "",
				Prompt:              "",
				EnableCommand:       "enable",
				EnablePasswordQuest: "",
				EnablePassword:      "testsx",
				EnablePrompt:        "",
				UseExternalSSH:      false,
				UseCRLF:             true,
			}

			t.Run("default", func(t *testing.T) {
				if test.skipNotExternal {
					return
				}

				conn := &Shell{SSHParams: params}
				results, err := script.Run(ctx, conn)
				if err != nil {
					t.Error(err)

					if !cmp.Equal(results, test.expected) {
						t.Error(cmp.Diff(results, test.expected))
					}
					return
				}

				if !cmp.Equal(results, test.expected) {
					t.Error(cmp.Diff(results, test.expected))
				}
			})
			t.Run("use_external_ssh", func(t *testing.T) {
				if test.skipExternal {
					return
				}

				params.UseExternalSSH = true
				conn := &Shell{SSHParams: params}
				results, err := script.Run(ctx, conn)
				if err != nil {
					t.Error(err)

					if !cmp.Equal(results, test.expected) {
						t.Error(cmp.Diff(results, test.expected))
					}
					return
				}

				if len(results) != len(test.expected) {
					t.Error(cmp.Diff(results, test.expected))
					return
				}

				for idx := range results {
					if !cmp.Equal(results[idx], test.expected[idx]) {
						if strings.Contains(results[idx].Incomming, "Store key in cache?") {
							continue
						}
						t.Error("[", idx, "]", cmp.Diff(results[idx], test.expected[idx]))
					}
				}
			})
		})
	}
}

func TestScriptSimpleTelnet(t *testing.T) {
	for _, test := range []struct {
		testname   string
		scriptText string
		expected   []ExecuteResult
	}{
		{
			testname: "test raw",
			scriptText: `
			@connect skiplogin
			@echo username:
			@send <<username>>
			@echo password:
			@send <<password>>
			@prompt
			@send <<enable>>
			@echo password:
			@send <<enable_password>>
			@prompt
			@exec echo abcd
	  		`,
			expected: []ExecuteResult{
				{LineNumber: 2, LineText: "@connect skiplogin"},
				{LineNumber: 3, LineText: "@echo username:", Incomming: "username:"},
				{LineNumber: 4, LineText: "@send <<username>>", Outgoing: "abc\r\n"},
				{LineNumber: 5, LineText: "@echo password:", Incomming: "password:"},
				{LineNumber: 6, LineText: "@send <<password>>", Outgoing: "********\r\n"},
				{LineNumber: 7, LineText: "@prompt", Incomming: "ABC>"},
				{LineNumber: 8, LineText: "@send <<enable>>", Outgoing: "enable\r\n"},
				{LineNumber: 9, LineText: "@echo password:", Incomming: "enable\r\npassword:"},
				{LineNumber: 10, LineText: "@send <<enable_password>>", Outgoing: "********\r\n"},
				{LineNumber: 11, LineText: "@prompt", Incomming: "\r\nenable OK\r\nabc#"},
				{LineNumber: 12, LineText: "@exec echo abcd", Command: "echo abcd", Incomming: "echo abcd\r\nprint abcd\r\nabc#", Outgoing: "echo abcd\r\n"},
			},
		},
		{
			testname: "test auto",
			scriptText: `
			@connect auto
			@exec echo abcd
	  		`,
			expected: []ExecuteResult{
				{
					LineNumber: 2,
					LineText:   "@connect auto",
					Incomming:  "username:password:ABC>enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:   "abc\r\n********\r\nenable\r\n********\r\n",
				},
				{
					LineNumber: 3,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},

		{
			testname: "test skiplogin",
			scriptText: `
			@connect skiplogin
			@login
			@exec echo abcd
	  		`,
			expected: []ExecuteResult{
				{
					LineNumber: 2,
					LineText:   "@connect skiplogin",
				},
				{
					LineNumber: 3,
					LineText:   "@login",
					Incomming:  "username:password:ABC>enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:   "abc\r\n********\r\nenable\r\n********\r\n",
				},
				{
					LineNumber: 4,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},

		{
			testname: "test skipenable",
			scriptText: `
			@connect skiplogin
			@login skipenable
			@enable
			@exec echo abcd
	  		`,
			expected: []ExecuteResult{
				{
					LineNumber: 2,
					LineText:   "@connect skiplogin",
				},
				{
					LineNumber: 3,
					LineText:   "@login skipenable",
					Incomming:  "username:password:ABC>",
					Outgoing:   "abc\r\n********\r\n",
				},
				{
					LineNumber: 4,
					LineText:   "@enable",
					Incomming:  "enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:   "enable\r\n********\r\n",
				},
				{
					LineNumber: 5,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},
	} {
		t.Run(test.testname, func(t *testing.T) {
			script, err := ParseScript(strings.NewReader(test.scriptText))
			if err != nil {
				t.Error(err)
				return
			}

			options := &telnetd.Options{}
			options.AddUserPassword("abc", "123")

			options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", telnetd.Echo)
			//options.WithNoEnable("ABC>", telnetd.Echo)

			listener, err := telnetd.StartServer(":", options)
			if err != nil {
				t.Error(err)
				return
			}
			defer listener.Close()

			port := listener.Port()
			ctx := context.Background()

			params := &TelnetParam{
				// Timeout: 30 * time.Second,
				Address: "127.0.0.1",
				Port:    port,
				// UserQuest: "",
				Username: "abc",
				// PasswordQuest: "",
				Password:            "123",
				Prompt:              "",
				EnableCommand:       "enable",
				EnablePasswordQuest: "",
				EnablePassword:      "testsx",
				EnablePrompt:        "",
				UseCRLF:             true,
			}

			conn := &Shell{TelnetParams: params}
			results, err := script.Run(ctx, conn)
			if err != nil {

				if !cmp.Equal(results, test.expected) {
					t.Error(cmp.Diff(results, test.expected))
				}

				t.Error(err)
				return
			}

			if !cmp.Equal(results, test.expected) {
				t.Error(cmp.Diff(results, test.expected))
			}
		})
	}
}

func TestScriptSimpleSSHMore(t *testing.T) {
	for _, test := range []struct {
		skipNotExternal bool
		skipExternal    bool
		testname        string
		scriptText      string
		expected        []ExecuteResult
	}{

		{
			skipNotExternal: false,
			skipExternal:    true,
			testname:        "test raw",
			scriptText: `
			@trigger "abc? [Y/N]:" {
				@write N\r\n
			}
			@connect skipprompt skipenable
			@prompt
			@send <<enable>>
			@echo password:
			@send <<enable_password>>
			@prompt
			@exec show
	  		`,
			expected: []ExecuteResult{
				{LineNumber: 2, LineText: "@trigger \"abc? [Y/N]:\" {"},
				{LineNumber: 5, LineText: "@connect skipprompt skipenable"},
				{LineNumber: 6, LineText: "@prompt", Incomming: "abc? [Y/N]:ABC>", Outgoing: "N\r\n"},
				{LineNumber: 7, LineText: "@send <<enable>>", Incomming: "", Outgoing: "enable\r\n"},
				{LineNumber: 8, LineText: "@echo password:", Incomming: "enable\r\npassword:"},
				{LineNumber: 9, LineText: "@send <<enable_password>>", Incomming: "", Outgoing: "********\r\n"},
				{LineNumber: 10, LineText: "@prompt", Incomming: "\r\nenable OK\r\nabc#"},
				{LineNumber: 11, LineText: "@exec show", Command: "show", Incomming: "show\r\nshow\r\nabcd\r\n-- more --efgh\r\n-- more --ijklmn\r\nabc#", Outgoing: "show\r\nyy"},
			},
		},

		{
			skipNotExternal: true,
			skipExternal:    false,
			testname:        "test raw",
			scriptText: `
			@trigger "abc? [Y/N]:" {
				@write N\r\n
			}
			@connect skipenable
			@send <<enable>>
			@echo password:
			@send <<enable_password>>
			@prompt
			@exec show
	  		`,
			expected: []ExecuteResult{
				{LineNumber: 2, LineText: "@trigger \"abc? [Y/N]:\" {"},
				{LineNumber: 5, LineText: "@connect skipenable"},
				{LineNumber: 6, LineText: "@send <<enable>>", Incomming: "", Outgoing: "enable\r\n"},
				{LineNumber: 7, LineText: "@echo password:", Incomming: "enable\r\npassword:"},
				{LineNumber: 8, LineText: "@send <<enable_password>>", Incomming: "", Outgoing: "********\r\n"},
				{LineNumber: 9, LineText: "@prompt", Incomming: "\r\nenable OK\r\nabc#"},
				{LineNumber: 10, LineText: "@exec show", Command: "show", Incomming: "show\r\nshow\r\nabcd\r\n-- more --efgh\r\n-- more --ijklmn\r\nabc#", Outgoing: "show\r\nyy"},
			},
		},

		{
			testname: "test auto",
			scriptText: `
			@trigger "abc? [Y/N]:" {
				@write N\r\n
			}
			@connect auto
			@exec echo abcd
			`,
			expected: []ExecuteResult{
				{LineNumber: 2, LineText: "@trigger \"abc? [Y/N]:\" {"},
				{LineNumber: 5, LineText: "@connect auto",
					Incomming: "abc? [Y/N]:ABC>enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:  "N\r\nenable\r\n********\r\n",
				},
				{
					LineNumber: 6,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},
		{
			testname: "test skiplogin",
			scriptText: `
			@trigger "abc? [Y/N]:" {
				@write N\r\n
			}
			@connect skiplogin
			@login
			@exec echo abcd
		    `,
			expected: []ExecuteResult{
				{LineNumber: 2, LineText: "@trigger \"abc? [Y/N]:\" {"},
				{
					LineNumber: 5,
					LineText:   "@connect skiplogin",
				},
				{
					LineNumber: 6,
					LineText:   "@login",
					Incomming:  "abc? [Y/N]:ABC>enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:   "N\r\nenable\r\n********\r\n",
				},
				{
					LineNumber: 7,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},
		{
			testname: "test skipenable",
			scriptText: `
			@trigger "abc? [Y/N]:" {
				@write N\r\n
			}
			@connect skiplogin
			@login skipenable
			@enable
			@exec echo abcd
		    `,
			expected: []ExecuteResult{
				{LineNumber: 2, LineText: "@trigger \"abc? [Y/N]:\" {"},
				{
					LineNumber: 5,
					LineText:   "@connect skiplogin",
				},
				{
					LineNumber: 6,
					LineText:   "@login skipenable",
					Incomming:  "abc? [Y/N]:ABC>",
					Outgoing:   "N\r\n",
				},
				{
					LineNumber: 7,
					LineText:   "@enable",
					Incomming:  "enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:   "enable\r\n********\r\n",
				},
				{
					LineNumber: 8,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},
	} {
		t.Run(test.testname, func(t *testing.T) {
			script, err := ParseScript(strings.NewReader(test.scriptText))
			if err != nil {
				t.Error(err)
				return
			}

			options := &sshd.Options{}
			options.AddUserPassword("abc", "123")
			options.WithQuest("abc? [Y/N]:", "N", "ABC>",
				sshd.WithEnable("enable", "password:", "testsx", "", "abc#", sshd.OS(sshd.Commands{
					"show": sshd.WithMore([]string{
						"abcd",
						"efgh",
						"ijklmn",
					}, []byte("-- more --"), nil),
				})))

			listener, err := sshd.StartServer(":", options)
			if err != nil {
				t.Error(err)
				return
			}
			defer listener.Close()

			port := listener.Port()
			ctx := context.Background()

			params := &SSHParam{
				// Timeout: 30 * time.Second,
				Address: "127.0.0.1",
				Port:    port,
				// UserQuest: "",
				Username: "abc",
				// PasswordQuest: "",
				Password:            "123",
				PrivateKey:          "",
				Prompt:              "",
				EnableCommand:       "enable",
				EnablePasswordQuest: "",
				EnablePassword:      "testsx",
				EnablePrompt:        "",
				UseExternalSSH:      false,
				UseCRLF:             true,
			}

			t.Run("default", func(t *testing.T) {
				if test.skipNotExternal {
					return
				}

				conn := &Shell{SSHParams: params}
				results, err := script.Run(ctx, conn)
				if err != nil {
					t.Error(err)

					if !cmp.Equal(results, test.expected) {
						t.Error(cmp.Diff(results, test.expected))
					}
					return
				}

				if !cmp.Equal(results, test.expected) {
					t.Error(cmp.Diff(results, test.expected))

					for _, a := range results {
						t.Errorf("%#v", a)
					}
				}
			})
			t.Run("use_external_ssh", func(t *testing.T) {
				if test.skipExternal {
					return
				}

				params.UseExternalSSH = true
				conn := &Shell{SSHParams: params}
				results, err := script.Run(ctx, conn)
				if err != nil {
					t.Error(err)

					if !cmp.Equal(results, test.expected) {
						t.Error(cmp.Diff(results, test.expected))
					}
					return
				}

				if len(results) != len(test.expected) {
					t.Error(cmp.Diff(results, test.expected))
					return
				}

				for idx := range results {
					if !cmp.Equal(results[idx], test.expected[idx]) {
						if strings.Contains(results[idx].Incomming, "Store key in cache?") {

							if !cmp.Equal(results[idx].SubResults, test.expected[idx].SubResults) {
								t.Error("[", idx, "]", cmp.Diff(results[idx].SubResults, test.expected[idx].SubResults))
							}

							continue
						}
						t.Error("[", idx, "]", cmp.Diff(results[idx], test.expected[idx]))
					}
				}
			})
		})
	}
}

func TestScriptTelnetWithGb18030(t *testing.T) {
	quest, _, err := transform.String(simplifiedchinese.GB18030.NewEncoder(), "中文测试? [Y/N]:")
	if err != nil {
		t.Error(err)
		return
	}

	for _, test := range []struct {
		skipNotExternal bool
		skipExternal    bool
		testname        string
		scriptText      string
		expected        []ExecuteResult
	}{
		{
			testname: "test auto",
			scriptText: `
			@trigger GB18030"中文测试? [Y/N]:" {
				@write N\r\n
			}
			@connect auto
			@exec echo abcd
			`,
			expected: []ExecuteResult{
				{LineNumber: 2, LineText: "@trigger GB18030\"中文测试? [Y/N]:\" {"},
				{LineNumber: 5, LineText: "@connect auto",
					Incomming: quest + "ABC>enable\r\npassword:\r\nenable OK\r\nabc#",
					Outgoing:  "N\r\nenable\r\n********\r\n",
				},
				{
					LineNumber: 6,
					LineText:   "@exec echo abcd",
					Command:    "echo abcd",
					Incomming:  "echo abcd\r\nprint abcd\r\nabc#",
					Outgoing:   "echo abcd\r\n",
				},
			},
		},
	} {
		t.Run(test.testname, func(t *testing.T) {
			script, err := ParseScript(strings.NewReader(test.scriptText))
			if err != nil {
				t.Error(err)
				return
			}

			options := &sshd.Options{}
			options.AddUserPassword("abc", "123")
			options.WithQuest(quest, "N", "ABC>",
				sshd.WithEnable("enable", "password:", "testsx", "", "abc#", sshd.OS(sshd.Commands{
					"show": sshd.WithMore([]string{
						"abcd",
						"efgh",
						"ijklmn",
					}, []byte("-- more --"), nil),
				})))

			listener, err := sshd.StartServer(":", options)
			if err != nil {
				t.Error(err)
				return
			}
			defer listener.Close()

			port := listener.Port()
			ctx := context.Background()

			params := &SSHParam{
				// Timeout: 30 * time.Second,
				Address: "127.0.0.1",
				Port:    port,
				// UserQuest: "",
				Username: "abc",
				// PasswordQuest: "",
				Password:            "123",
				PrivateKey:          "",
				Prompt:              "",
				EnableCommand:       "enable",
				EnablePasswordQuest: "",
				EnablePassword:      "testsx",
				EnablePrompt:        "",
				UseExternalSSH:      false,
				UseCRLF:             true,
			}

			t.Run("default", func(t *testing.T) {
				if test.skipNotExternal {
					return
				}

				conn := &Shell{SSHParams: params}
				results, err := script.Run(ctx, conn)
				if err != nil {
					t.Error(err)

					if !cmp.Equal(results, test.expected) {
						t.Error(cmp.Diff(results, test.expected))
					}
					return
				}

				if !cmp.Equal(results, test.expected) {
					t.Error(cmp.Diff(results, test.expected))
				}
			})
			t.Run("use_external_ssh", func(t *testing.T) {
				if test.skipExternal {
					return
				}

				params.UseExternalSSH = true
				conn := &Shell{SSHParams: params}
				results, err := script.Run(ctx, conn)
				if err != nil {
					t.Error(err)

					if !cmp.Equal(results, test.expected) {
						t.Error(cmp.Diff(results, test.expected))
					}
					return
				}

				if len(results) != len(test.expected) {
					t.Error(cmp.Diff(results, test.expected))
					return
				}

				for idx := range results {
					if !cmp.Equal(results[idx], test.expected[idx]) {
						if strings.Contains(results[idx].Incomming, "Store key in cache?") {
							continue
						}
						t.Error("[", idx, "]", cmp.Diff(results[idx], test.expected[idx]))
					}
				}
			})
		})
	}
}
