package shell

import (
	"bytes"
	"encoding/hex"
	"sync"
	"unicode"

	//"fmt"
	"io"
)

func crossingMatch(s, pattern []byte) int {
	l := len(s)
	if l > len(pattern) {
		l = len(pattern)
	}

	for ; l > 0; l-- {
		if bytes.Equal(s[len(s)-l:], pattern[:l]) {
			return l
		}
	}
	return 0
}

// crossingMatch2(s, b, pattern) 等价于
//
//	crossingMatch(append(s, b), pattern)
func crossingMatch2(s []byte, b byte, pattern []byte) int {
	l := len(s)
	if l > len(pattern)-1 {
		l = len(pattern) - 1
	}

	for ; l > 0; l-- {
		if pattern[l] == b && bytes.Equal(s[len(s)-l:], pattern[:l]) {
			return l + 1
		}
	}
	return 0
}

func match(s, pattern []byte, offset int) (bool, int) {
	orignOffset := offset
	for offset > 0 {
		//fmt.Printf("s=%v\tpattern=%v\toffset=%v\r\n", string(s), string(pattern), offset)
		if (offset + len(s)) < len(pattern) {
			if bytes.Equal(s, pattern[offset:offset+len(s)]) {
				return false, offset + len(s)
			}
		} else {
			//fmt.Printf("s=%v\tpattern=%v\toffset=%v, left=%v, right=%v\r\n",
			// string(s), string(pattern), offset, string(s[:len(pattern)-offset]), string(pattern[offset:]))
			if bytes.Equal(s[:len(pattern)-offset], pattern[offset:]) {
				return true, 0
			}
		}
		offset = crossingMatch(pattern[orignOffset-offset+1:orignOffset], pattern)
	}

	if -1 != bytes.Index(s, pattern) {
		return true, 0
	}
	return false, crossingMatch(s, pattern)
}

type matchWriter struct {
	maxSize  uint64
	size     uint64
	mu       sync.Mutex
	patterns [][]byte
	offset   []int
	out      io.Writer
	cb       func(idx int, pattern []byte)
}

func (mw *matchWriter) Reset(out io.Writer, patterns [][]byte, cb func(idx int, pattern []byte)) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.out = out
	mw.patterns = patterns
	if len(patterns) > 0 {
		mw.offset = make([]int, len(patterns))
	} else {
		mw.offset = nil
	}
	mw.cb = cb
	mw.size = 0
}

func (mw *matchWriter) Write(p []byte) (n int, err error) {
	mw.mu.Lock()
	out := mw.out
	mw.mu.Unlock()
	if nil != out {
		n, err = out.Write(p)
	} else {
		n = len(p)
	}

	mw.mu.Lock()
	defer mw.mu.Unlock()

	if len(mw.patterns) > 0 {
		mw.size += uint64(n)
		if mw.maxSize > 0 && mw.size >= mw.maxSize && nil != mw.cb {
			func() {
				mw.mu.Unlock()
				defer mw.mu.Lock()
				mw.cb(-1, nil)
			}()
		}

		for idx, pattern := range mw.patterns {
			matched, offset := match(p[:n], pattern, mw.offset[idx])
			if matched && nil != mw.cb {

				func() {
					mw.mu.Unlock()
					defer mw.mu.Lock()

					mw.cb(idx, pattern)
				}()

				mw.offset[idx] = 0
				break
			} else {
				mw.offset[idx] = offset
			}
		}
	}
	return n, err
}

func Wrap(out io.Writer, patterns [][]byte, cb func(idx int, pattern []byte)) io.Writer {
	if 0 == len(patterns) || nil == cb {
		return out
	}
	return &matchWriter{patterns: patterns, offset: make([]int, len(patterns)), out: out, cb: cb}
}

type safeWriter struct {
	sync.Mutex
	out io.Writer
}

func (mw *safeWriter) Write(p []byte) (n int, e error) {
	mw.Lock()
	defer mw.Unlock()
	return mw.out.Write(p)
}

func safeIO(out io.Writer) io.Writer {
	if w, ok := out.(*safeWriter); ok {
		return w
	}
	if w, ok := out.(*SafeWriter); ok {
		return w
	}
	return &safeWriter{out: out}
}

type SafeReader struct {
	sync.Mutex
	tag string
	R   io.Reader
}

func (srw *SafeReader) Read(p []byte) (int, error) {
	srw.Lock()
	defer srw.Unlock()

	return srw.R.Read(p)
}

type SafeWriter struct {
	sync.Mutex
	tag string
	W   io.Writer
}

func (srw *SafeWriter) WriteWithTag(tag string, p []byte) (int, error) {
	srw.Lock()
	defer srw.Unlock()

	if srw.tag != tag {
		srw.tag = tag

		srw.W.Write([]byte("\r\n"))
		srw.W.Write([]byte(tag))
		srw.W.Write([]byte(": "))
	}

	var e error

	for idx, c := range p {
		switch c {
		case '\\':
			_, e = io.WriteString(srw.W, "\\")
		// case '\r':
		// 	_, e = io.WriteString(srw.W, "\\r")
		// case '\n':
		// 	_, e = io.WriteString(srw.W, "\\n")
		// case '\t':
		// 	_, e = io.WriteString(srw.W, "\\t")
		case '[':
			_, e = io.WriteString(srw.W, "\\[")
		case ']':
			_, e = io.WriteString(srw.W, "\\]")
		default:
			if unicode.IsDigit(rune(c)) || (c >= 32 && c <= 127) {
				_, e = srw.W.Write([]byte{c})
			} else {
				_, e = io.WriteString(srw.W, "["+hex.EncodeToString([]byte{c})+"]")
			}
		}

		if e != nil {
			return idx, e
		}
	}
	return len(p), nil
}

func (srw *SafeWriter) Write(p []byte) (int, error) {
	srw.Lock()
	defer srw.Unlock()
	return srw.W.Write(p)
}

type TagWriter struct {
	W interface {
		WriteWithTag(tag string, p []byte) (int, error)
	}

	Tag string
}

func (tw *TagWriter) Write(p []byte) (int, error) {
	return tw.W.WriteWithTag(tw.Tag, p)
}

func toHexIfNeed(p []byte) []byte {
	need := false
	for _, c := range p {
		if !unicode.IsPrint(rune(c)) {
			need = true
			break
		}
	}
	if !need {
		return p
	}
	var newCopy = make([]byte, 0, len(p)+64)
	for _, c := range p {
		switch c {
		case '\\':
			newCopy = append(newCopy, '\\')
		// case '\r':
		// 	_, e = io.WriteString(srw.W, "\\r")
		// case '\n':
		// 	_, e = io.WriteString(srw.W, "\\n")
		// case '\t':
		// 	_, e = io.WriteString(srw.W, "\\t")
		case '[':
			newCopy = append(newCopy, '\\', '[')
		case ']':
			newCopy = append(newCopy, '\\', ']')
		default:
			if unicode.IsDigit(rune(c)) || (c >= 32 && c <= 127) {
				newCopy = append(newCopy, c)
			} else {
				newCopy = append(newCopy, '[')
				newCopy = append(newCopy, hex.EncodeToString([]byte{c})...)
				newCopy = append(newCopy, ']')
			}
		}
	}
	return newCopy
}

func ToHexStringIfNeed(recvBytes []byte) string {
	if len(recvBytes) == 0 {
		return ""
	}
	return string(toHexIfNeed(recvBytes))
}
