package shell

import (
	"bytes"
	"context"
	"strconv"
	"testing"
)

func TestExpect2(t *testing.T) {

	var idx string
	makeFunc := func(prefix string) func(conn Conn, bs []byte, i int) (bool, error) {
		return func(conn Conn, bs []byte, i int) (bool, error) {
			idx = prefix + strconv.FormatInt(int64(i), 10)
			return false, nil
		}
	}

	for _, test := range []struct {
		text          string
		matchs        []Matcher
		exceptedIndex string
		shouldError   bool
	}{
		{
			text: "a",
			matchs: []Matcher{
				Match("a", makeFunc("a_")),
			},
			exceptedIndex: "a_0",
		},
		{
			text: "a",
			matchs: []Matcher{
				Match([]string{"a", "b"}, makeFunc("a_")),
			},
			exceptedIndex: "a_0",
		},
		{
			text: "a",
			matchs: []Matcher{
				Match([]string{"b", "a"}, makeFunc("a_")),
			},
			exceptedIndex: "a_1",
		},
		{
			text: "a",
			matchs: []Matcher{
				Match([]string{"c", "b", "a"}, makeFunc("a_")),
			},
			exceptedIndex: "a_2",
		},

		{
			text: "a",
			matchs: []Matcher{
				Match("b", makeFunc("b_")),
				Match("a", makeFunc("a_")),
			},
			exceptedIndex: "a_0",
		},

		{
			text: "b",
			matchs: []Matcher{
				Match("b", makeFunc("b_")),
				Match("a", makeFunc("a_")),
			},
			exceptedIndex: "b_0",
		},

		{
			text: "b",
			matchs: []Matcher{
				Match("b", makeFunc("b_")),
				Match("b", makeFunc("a_")),
			},
			exceptedIndex: "b_0",
		},

		{
			text: "b",
			matchs: []Matcher{
				Match("c", makeFunc("c_")),
				Match("b", makeFunc("b_")),
				Match("a", makeFunc("a_")),
			},
			exceptedIndex: "b_0",
		},
	} {
		idx = ""

		var buf bytes.Buffer
		buf.WriteString(test.text)

		conn := MakeConnWrapper(nil, &buf, &buf)
		err := Expect(context.Background(), &conn, test.matchs...)

		if err != nil {
			if !test.shouldError {
				t.Error(err)
			}
			continue
		}

		if idx != test.exceptedIndex {
			t.Error("want", test.exceptedIndex, "got", idx)
		}
	}
}
