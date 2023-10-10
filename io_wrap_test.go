package shell

import (
	"bytes"
	//"fmt"
	"strings"
	"testing"
)

var testMatchs = []struct {
	pattern string
	data    []string
	found   bool
}{
	{pattern: "a", data: []string{"a"}, found: true},
	{pattern: "a", data: []string{"abc"}, found: true},
	{pattern: "b", data: []string{"abc"}, found: true},
	{pattern: "c", data: []string{"abc"}, found: true},
	{pattern: "d", data: []string{"abc"}, found: false},
	{pattern: "ab", data: []string{"abc"}, found: true},
	{pattern: "bc", data: []string{"abc"}, found: true},
	{pattern: "abc", data: []string{"abc"}, found: true},
	{pattern: "abcd", data: []string{"abc"}, found: false},
	{pattern: "b", data: []string{"a", "b", "c"}, found: true},
	{pattern: "ab", data: []string{"a", "b", "c"}, found: true},
	{pattern: "ab", data: []string{"ab", "c"}, found: true},
	{pattern: "ab", data: []string{"a", "bc"}, found: true},
	{pattern: "ab", data: []string{"abc"}, found: true},

	{pattern: "abc", data: []string{"a", "b", "c"}, found: true},
	{pattern: "abc", data: []string{"ab", "c"}, found: true},
	{pattern: "abc", data: []string{"a", "bc"}, found: true},
	{pattern: "abc", data: []string{"abc"}, found: true},
	{pattern: "abbc", data: []string{"aa", "abbc"}, found: true},
	{pattern: "abbbc", data: []string{"aa", "abbbc"}, found: true},
	{pattern: "aabc", data: []string{"aa", "abc"}, found: true},
	{pattern: "aaabc", data: []string{"aa", "abc"}, found: true},
	{pattern: "aaabc", data: []string{"aaaaaaa", "abc"}, found: true},
	{pattern: "aaabc", data: []string{"aaaaaaa", "bc"}, found: true},
	{pattern: "aaabc", data: []string{"aaaaaaa", "aabc"}, found: true},
	{pattern: "aaabc", data: []string{"aaaaaaa", "aaabc"}, found: true},
	{pattern: "aaabc", data: []string{"aaaaaaa", "aaa", "bc"}, found: true},
	{pattern: "aaabbbc", data: []string{"aaaaaaa", "bb", "bc"}, found: true},
	{pattern: "aaabbbc", data: []string{"aaaaaaa", "bbb", "bc"}, found: false},
	{pattern: "aaabbbc", data: []string{"aaaaaaa", "bbbb", "bc"}, found: false},
	{pattern: "aaabbbc", data: []string{"aaaaaaa", "bb", "bbc"}, found: false},
	{pattern: "aaabbbc", data: []string{"aaaaaaa", "aabb", "bbc"}, found: false},
	{pattern: "aaabbbc", data: []string{"aaaaaab", "bb", "bc"}, found: false},
	{pattern: "aaabbbc", data: []string{"aaaaaa", "bb", "cc"}, found: false},
	{pattern: "aaabbcc", data: []string{"aaaaaa", "bb", "cc"}, found: true},
	{pattern: "aaabbccdd", data: []string{"aaaaaa", "bb", "cc", "dd"}, found: true},
	{pattern: "aaabbccdd", data: []string{"aaaaaa", "bb", "cc", "ddddd"}, found: true},
	{pattern: "aaabbccdd", data: []string{"aaaaaabbccddddd"}, found: true},
	{pattern: "abcd", data: []string{"abc"}, found: false},
}

func TestMatch(t *testing.T) {
	for _, m := range testMatchs {

		//fmt.Printf("\"%v\" match \"%v\"\r\n", m.data, m.pattern)

		var buf bytes.Buffer
		var matched bool
		var matchedIdx int
		writer := &matchWriter{patterns: [][]byte{[]byte("adsbadsfsadfs22"), []byte(m.pattern)},
			offset: make([]int, 2),
			out:    &buf, cb: func(idx int, b []byte) {
				matched = true
				matchedIdx = idx
			}}
		for _, p := range m.data {
			writer.Write([]byte(p))
		}

		if m.found != matched {
			t.Errorf("\"%v\" match \"%v\" failed, excepted is %v, actual is %v %v", m.data, m.pattern, m.found, matched, matchedIdx)
			//fmt.Printf("\"%v\" match \"%v\" failed, excepted is %v, actual is %v\r\n", m.data, m.pattern, m.found, writer.Matched())
		} else if strings.Join(m.data, "") != buf.String() {
			t.Errorf("\"%v\" match \"%v\" failed, result is error", m.data, m.pattern)
		}
	}
}
