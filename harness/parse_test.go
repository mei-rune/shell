package harness

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, test := range []struct {
		text      string
		cmdCount  int
		err       string
		questions [][]string
	}{
		{
			text: `@trigger {}`,
			err:  "非预期的块结束符",
		},
		{
			text: `@trigger {`,
			err:  "没有块结束符",
		},
		{
			text: `@trigger "abc" {
				@write aaa
			}`,
			cmdCount: 1,
			questions: [][]string{
				[]string{"abc"},
			},
		},
		{
			text: `@trigger "\r\n\tabc" {
				@write aaa
			}`,
			cmdCount: 1,
			questions: [][]string{
				[]string{"\r\n\tabc"},
			},
		},
		{
			text: `@trigger "abc {
				@write aaa
			}`,
			err: "参数语法不正确",
		},
		{
			text: `@trigger "abc" abc {
				@write aaa
			}`,
			err: "选项 'abc' 是未知的",
		},
	} {
		t.Run(test.text, func(t *testing.T) {
			spt, err := ParseScript(strings.NewReader(test.text))
			if err != nil {
				t.Log(err, test.text)
				if test.err == "" {
					t.Error(err)
					return
				}
				if !strings.Contains(err.Error(), test.err) {
					t.Error(err)
				}
				return
			}

			if len(spt.Cmds) != test.cmdCount {
				t.Error("want", test.cmdCount, "got", len(spt.Cmds))
			}

			if len(test.questions) > 0 {
				sh := &Shell{}
				spt.Run(nil, sh)

				if len(test.questions) != len(sh.questions) {
					t.Error("want", len(test.questions), "got", len(sh.questions))
				}
				var questions [][]string
				for _, q := range sh.questions {
					questions = append(questions, q.Strings())
				}
				if !reflect.DeepEqual(test.questions, questions) {
					t.Error("want", test.questions, "got", questions)
				}
			}
		})
	}
}
