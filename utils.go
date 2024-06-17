package shell

import (
	"bufio"
	"bytes"
	"io"
	"unicode"

	"github.com/runner-mei/errors"
)

func hasMore(bs []byte) bool {
	for _, more := range MorePrompts {
		if bytes.Contains(bs, more) {
			return true
		}
	}
	return false
}

func isMoreLine(bs []byte) bool {
	for {
		oldLen := len(bs)
		bs = bytes.TrimSpace(bs)
		bs = bytes.Trim(bs, "-")

		if len(bs) == oldLen {
			break
		}
		if len(bs) == 0 {
			return false
		}
	}

	return bytes.Equal(bytes.ToLower(bs), []byte("more"))
}

func isAll(bs []byte, charset []byte) bool {
	for _, c := range bs {
		if -1 == bytes.IndexByte(charset, c) {
			return false
		}
	}
	return true
}

func RemoveCtrlCharByLine(bs [][]byte, capacity int) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	for _, line := range bs {
		cidx := bytes.IndexByte(line, 8)
		if -1 != cidx {
			if hasMore(line) {
				buf.Write(removeCtrlCharWithIndexAndOffset(line, cidx, 0))
				buf.WriteString("\n")
				continue
			}
		} else if isMoreLine(line) {
			continue
		} else if cidx = bytes.IndexByte(line, 13); cidx >= 0 && isMoreLine(line[:cidx]) {
			buf.Write(line[cidx+1:])
			buf.WriteString("\n")
			continue
		}

		buf.Write(RemoveCtrlChar(line))
		buf.WriteString("\n")
	}
	return buf.Bytes()
}

func RemoveCtrlChar(bs []byte) []byte {
	return removeCtrlCharWithIndexAndOffset(bs, 0, 0)
}

func splitEscapeKey(bs []byte) ([]byte, byte, int) {
	for i := 0; i < len(bs); i++ {
		if !unicode.IsDigit(rune(bs[i])) {
			if i > 0 {
				return bs[0 : i-1], bs[i], i + 1
			}
			return nil, bs[i], i + 1
		}
	}
	return nil, byte(0), 0
}

func removeCtrlCharWithIndexAndOffset(bs []byte, idx, offset int) []byte {
	for i := idx; i < len(bs); i++ {
		switch bs[i] {
		case 0:
			// don't some thing.
		case 8:
			if offset > 0 {
				offset--
			}
		case 27: // ESC
			j := i + 1
			if j < len(bs) {
				if '[' == bs[j] {
					_, c, o := splitEscapeKey(bs[j+1:])
					switch c {
					//case 'C':

					case 'D', 'J', 'K':
						offset = 0
						i++
						i += o
					}
				} else {
					i = j

					if offset > 0 {
						offset--
					}
					if offset > 0 {
						offset--
					}
				}
			}
		default:
			if offset != i {
				bs[offset] = bs[i]
			}
			offset++
		}
	}
	return bs[:offset]
}

func RemoveNullChar(bs []byte) []byte {
	copy_idx := 0
	for i := 0; i < len(bs); i++ {
		if 0 == bs[i] {
			continue
		}
		if copy_idx != i {
			bs[copy_idx] = bs[i]
		}

		copy_idx++
	}

	return bs[:copy_idx]
}

func SplitLines(bs []byte) [][]byte {
	if nil == bs {
		return nil
	}
	scanner := bufio.NewScanner(bytes.NewReader(bs))
	res := make([][]byte, 0, 10)
	for scanner.Scan() {
		bs := make([]byte, len(scanner.Bytes()))
		copy(bs, scanner.Bytes())

		res = append(res, bs)
	}

	if nil != scanner.Err() {
		panic(scanner.Err())
	}
	return res
}

type closeFunc func() error

func (c closeFunc) Close() error {
	return c()
}

func MultWriters(w1 io.Writer, list ...io.Writer) io.Writer {
	if len(list) == 0 {
		return w1
	}

	var passwordWriters []SendPasswordWriter
	var noPasswordWriters []io.Writer
	var closers []io.Closer

	wc1, ok1 := w1.(io.WriteCloser)
	if ok1 {
		closers = append(closers, wc1)
	}
	if pw, ok := w1.(SendPasswordWriter); ok {
		passwordWriters = append(passwordWriters, pw)
	} else {
		noPasswordWriters = append(noPasswordWriters, w1)
	}

	for _, w := range list {
		if c, ok := w.(withWriteCloser); ok {
			if len(c.closers) > 0 {
				closers = append(closers, c.closers...)
			}
			if len(c.passwordWriters) > 0 {
				passwordWriters = append(passwordWriters, c.passwordWriters...)
			}
			if len(c.noPasswordWriters) > 0 {
				noPasswordWriters = append(noPasswordWriters, c.noPasswordWriters...)
			}
			continue
		}
		wc, ok := w.(io.WriteCloser)
		if ok {
			closers = append(closers, wc)
		}
		if pw, ok := w.(SendPasswordWriter); ok {
			passwordWriters = append(passwordWriters, pw)
		} else {
			noPasswordWriters = append(noPasswordWriters, w)
		}
	}

	if len(passwordWriters) == 0 && len(closers) == 0 {
		return io.MultiWriter(append(list, w1)...)
	}

	return withWriteCloser{
		singleWriter:      io.MultiWriter(append(list, w1)...),
		passwordWriters:   passwordWriters,
		noPasswordWriters: noPasswordWriters,
		closers:           closers,
	}
}

var _ io.WriteCloser = withWriteCloser{}

type withWriteCloser struct {
	singleWriter io.Writer

	passwordWriters   []SendPasswordWriter
	noPasswordWriters []io.Writer
	closers           []io.Closer
}

func (w withWriteCloser) Write(p []byte) (int, error) {
	return w.singleWriter.Write(p)
}

func (w withWriteCloser) WriteString(s string) (int, error) {
	return io.WriteString(w.singleWriter, s)
}

func (w withWriteCloser) SendPassword(s []byte) error {
	if w.noPasswordWriters == nil {
		_, err := w.Write(s)
		return err
	}
	for _, pw := range w.passwordWriters {
		if err := pw.SendPassword(s); err != nil {
			return err
		}
	}

	for _, npw := range w.noPasswordWriters {
		n, err := npw.Write(s)
		if err != nil {
			return err
		}
		if n != len(s) {
			return io.ErrShortWrite
		}
	}
	return nil
}

func (w withWriteCloser) Close() error {
	var errList []error
	for _, a := range w.closers {
		if e := a.Close(); e != nil {
			errList = append(errList, e)
		}
	}

	if len(errList) == 0 {
		return nil
	}

	return errors.ErrArray(errList)
}

type WriteFunc func([]byte) (int, error)

func (c WriteFunc) Write(p []byte) (int, error) {
	return c(p)
}
func (c WriteFunc) SendPassword(p []byte) (int, error) {
	return c([]byte("********"))
}

type passwordWriter struct {
	io.Writer
}

func (w passwordWriter) SendPassword(s []byte) error {
	_, err := w.Write([]byte("********"))
	return err
}

var _ io.Writer = passwordWriter{}
var _ passwordWriter = passwordWriter{}

func PasswordWriter(w io.Writer) io.Writer {
	_, ok := w.(SendPasswordWriter)
	if ok {
		return w
	}
	return passwordWriter{w}
}

func WriteFull(w io.Writer, bs []byte) error {
	for len(bs) > 0 {
		n, e := w.Write(bs)
		if nil != e {
			return e
		}
		bs = bs[n:]
	}
	return nil
}

func ParseCmdOutput(bs []byte, cmd, prompt, characteristic []byte) ([]byte, error) {
	if nil == bs || 0 == len(bs) {
		return nil, errors.New("console output is empty")
	}

	lineArray := SplitLines(bs)
	if nil == lineArray || 0 == len(lineArray) {
		return nil, errors.New("console output is empty")
	}

	// for idx := range lineArray {
	// 	fmt.Printf("%q\r\n", ToHexStringIfNeed(lineArray[idx]))
	// }

	fullPrompt := bytes.TrimRightFunc(lineArray[len(lineArray)-1], unicode.IsSpace)
	if len(prompt) > 0 && !bytes.Contains(fullPrompt, prompt) {
		return nil, errors.New("last line of '" + string(bs) + "' isn't prompt.")
	}

	lineArray = lineArray[:len(lineArray)-1]

	// find the last prompt
	foundIdx := -1
	for idx := 0; idx < len(lineArray); idx++ {
		line := lineArray[idx]
		if bytes.HasPrefix(line, fullPrompt) {
			foundIdx = idx
		}
	}
	lineArray = lineArray[foundIdx+1:]

	if len(characteristic) > 0 {
		foundIdx = -1
		for idx := 0; idx < len(lineArray); idx++ {
			line := lineArray[idx]

			if bytes.Contains(line, characteristic) {
				foundIdx = idx
				break
			}
		}

		if foundIdx < 0 {
			return nil, errors.New("characteristic '" + string(characteristic) + "' isn't found in '" + string(bs) + "'.")
		}
	}

	if len(cmd) > 0 {
		foundIdx = -1
		for idx := 0; idx < 3 && idx < len(lineArray); idx++ {
			if !bytes.Contains(lineArray[idx], cmd) {
				foundIdx = idx
				break
			}
		}

		if foundIdx >= 0 {
			lineArray = lineArray[foundIdx:]
		}
	}

	// fmt.Println("======")
	// fmt.Printf("%q\r\n", bs)
	// fmt.Println("======")
	// for idx := range lineArray {
	// 	fmt.Printf("%q\r\n", ToHexStringIfNeed(lineArray[idx]))
	// }
	// tmp := RemoveCtrlCharByLine(lineArray, len(bs))
	// fmt.Println("======")
	// fmt.Printf("%q\r\n", ToHexStringIfNeed(tmp))
	// fmt.Println("======")

	return RemoveCtrlCharByLine(lineArray, len(bs)), nil
}
