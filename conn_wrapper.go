package shell

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/runner-mei/errors"
)

var ErrTimeout = errors.ErrTimeout
var IsTimeout = errors.IsTimeoutError

func newConnWrapper(conn io.ReadWriteCloser) ConnWrapper {
	return MakeConnWrapper(conn, conn, conn)
}

func MakeConnWrapper(closer io.Closer, w io.Writer, r io.Reader) ConnWrapper {
	conn := ConnWrapper{}
	conn.Init(closer, w, r)
	return conn
}

type ConnWrapper struct {
	session io.Closer
	w       io.Writer
	r       io.Reader

	readByte interface {
		ReadByte() (byte, error)
	}
	setReadDeadline interface {
		SetReadDeadline(t time.Duration) error
	}
	setWriteDeadline interface {
		SetWriteDeadline(t time.Duration) error
	}
	drainto drainto

	teeR atomic.Value
	teeW atomic.Value

	useCRLF bool
}

func (c *ConnWrapper) UseCRLF() {
	c.useCRLF = true
}

func (c *ConnWrapper) Init(closer io.Closer, w io.Writer, r io.Reader) {
	c.session = closer
	c.w = w
	c.r = r
	c.readByte, _ = r.(interface {
		ReadByte() (byte, error)
	})
	c.setReadDeadline, _ = r.(interface {
		SetReadDeadline(t time.Duration) error
	})
	c.drainto, _ = r.(drainto)
	c.setWriteDeadline, _ = w.(interface {
		SetWriteDeadline(t time.Duration) error
	})
}

func (c *ConnWrapper) Close() error {
	if c.session == nil {
		return nil
	}
	return c.session.Close()
}

func (c *ConnWrapper) readUntil(buf *bytes.Buffer, delims [][]byte) (int, error) {
	if len(delims) == 0 {
		return 0, nil
	}
	p := make([][]byte, len(delims))
	for i, s := range delims {
		if len(s) == 0 {
			return i, nil
		}
		p[i] = s
	}

	for {
		b, err := c.ReadByte()
		if err != nil {
			if IsTimeout(err) {
				if buf != nil {
					bs := buf.Bytes()
					bs = bytes.TrimSpace(bs)
					if bytes.HasSuffix(bs, []byte("#")) {
						for i := range delims {
							if bytes.Equal(delims[i], []byte("#")) {
								return i, nil
							}
						}
					}
				}
			}
			return -1, err
		}
		if buf != nil {
			buf.WriteByte(b)
		}

		for i := range p {
			if p[i][0] != b {

				// fmt.Println("b=", string([]byte{b}), ", end=", string(p[i]))
				// fmt.Println("alreadyRecv=", string(alreadyRecv))
				// fmt.Println("crossingMatch=", n, string(delims[i]))
				// fmt.Println("newdelims=", string(delims[i][n:]))

				alreadyRecvSize := len(delims[i]) - len(p[i])
				alreadyRecv := delims[i][:alreadyRecvSize]
				// 注意 下面一句 是不可以改成 crossingMatch(append(alreadyRecv, b), delims[i])
				// 因为 append() 会修改 alreadyRecv，也可以会导致 delims[i] 的内容被修改
				n := crossingMatch2(alreadyRecv, b, delims[i])
				p[i] = delims[i][n:]

				//p[i] = delims[i]
				// if p[i][0] == b {
				// 	p[i] = p[i][1:]
				// }
			} else {
				// fmt.Println("b=", string([]byte{b}), ", end=", string(p[i]))
				p[i] = p[i][1:]
			}

			if len(p[i]) == 0 {
				if buf != nil {
					bs := buf.Bytes()
					if SkipHits(bs, delims[i]) {
						p[i] = delims[i]
						continue
					}
				}

				return i, nil
			}
		}
	}
	// panic(nil)
}

func (c *ConnWrapper) SetReadDeadline(t time.Duration) error {
	if c.setReadDeadline == nil {
		return nil //c.rc.SetTimeout(t)
	}

	return c.setReadDeadline.SetReadDeadline(t)
}

func (c *ConnWrapper) SetWriteDeadline(t time.Duration) error {
	if c.setWriteDeadline == nil {
		return nil //c.rc.SetWriteDeadline(t)
	}

	return c.setWriteDeadline.SetWriteDeadline(t)
}

func (c *ConnWrapper) Expect(delims [][]byte) (int, []byte, error) {
	var buf bytes.Buffer
	idx, err := c.readUntil(&buf, delims)
	return idx, buf.Bytes(), err
}

var ln = []byte("\n")
var crlf = []byte("\r\n")

func (c *ConnWrapper) Sendln(s []byte) error {
	if len(s) > 0 {
		_, err := c.Write(s)
		if err != nil {
			return err
		}
	}

	if bytes.HasSuffix(s, ln) {
		return nil
	}

	lnBytes := ln
	if c.useCRLF {
		lnBytes = crlf
	}

	_, err := c.Write(lnBytes)
	return err
}

func (c *ConnWrapper) Send(s []byte) error {
	_, err := c.Write(s)
	return err
}

func (c *ConnWrapper) SendPassword(s []byte) error {
	var err error
	if pw, ok := c.w.(SendPasswordWriter); ok {
		err = pw.SendPassword(s)
	} else {
		_, err = c.w.Write(s)
	}

	if err != nil {
		return err
	}
	teeWriter := c.teeWriter()
	teeWriter.Write([]byte("********"))

	lnBytes := ln
	if c.useCRLF {
		lnBytes = crlf
	}
	_, err = c.Write(lnBytes)
	if err == nil {
		teeWriter.Write(lnBytes)
	}
	return err
}

func (c *ConnWrapper) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if n > 0 {
		c.teeWriter().Write(p[:n])
	}
	return n, err
}

func (c *ConnWrapper) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 {
		c.teeReader().Write(p[:n])
	}
	return n, err
}

func (c *ConnWrapper) ReadByte() (byte, error) {
	var bs [1]byte

	if c.readByte != nil {
		b, err := c.readByte.ReadByte()
		if err == nil {
			bs[0] = b
			c.teeReader().Write(bs[:])
		}
		return b, err
	}
	n, err := c.Read(bs[:])
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.ErrNoProgress
	}
	return bs[0], nil
}

func (c *ConnWrapper) DrainOff() (int, error) {
	if c.drainto != nil {
		return c.drainto.DrainTo(c.teeReader())
	}
	return 0, nil
}

func (c *ConnWrapper) teeWriter() io.Writer {
	o := c.teeW.Load()
	if o == nil {
		return ioutil.Discard
	}

	w, _ := o.(io.Writer)
	if w != nil {
		return w
	}
	return ioutil.Discard
}

func (c *ConnWrapper) teeReader() io.Writer {
	o := c.teeR.Load()
	if o == nil {
		return ioutil.Discard
	}

	w, _ := o.(io.Writer)
	if w != nil {
		return w
	}
	return ioutil.Discard
}

func (c *ConnWrapper) setTee(tee *atomic.Value, w io.Writer) context.CancelFunc {
	var cancel context.CancelFunc
	old := tee.Load()
	if old != nil {
		if out, ok := old.(*wout); ok && out != nil && out.Writer != nil {
			w = io.MultiWriter(out.Writer, w)
		}
		cancel = func() {
			tee.Store(old)
		}
	} else {
		cancel = func() {
			tee.Store(&wout{})
		}
	}

	tee.Store(&wout{w})
	return cancel
}

func (c *ConnWrapper) SetTeeWriter(w io.Writer) context.CancelFunc {
	return c.setTee(&c.teeW, w)
}

func (c *ConnWrapper) SetTeeReader(w io.Writer) context.CancelFunc {
	return c.setTee(&c.teeR, w)
}

func (c *ConnWrapper) SetTeeOutput(w io.Writer) context.CancelFunc {
	c1 := c.SetTeeWriter(w)
	c2 := c.SetTeeReader(w)

	return func() {
		c1()
		c2()
	}
}

type wout struct {
	io.Writer
}

func (w *wout) Write(p []byte) (int, error) {
	if w == nil {
		return len(p), nil
	}
	if w.Writer == nil {
		return len(p), nil
	}
	return w.Writer.Write(p)
}

var _ io.Writer = &wout{}

func SkipHits(bs, delim []byte) bool {
	//log.Println("==================== SkipHits test", string(bs))

	if bytes.HasSuffix(bs, []byte("Last login:")) ||
		bytes.HasSuffix(bs, []byte("last login:")) {
		return true
	}
	if bytes.HasSuffix(bs, []byte("</>")) {
		return true
	}

	if bytes.HasSuffix(bs, []byte("<myuser>")) ||
		bytes.HasSuffix(bs, []byte("<mypassword>")) {
		//log.Println("==================== skip", string(bs))
		return true
	}
	if lastLF := bytes.LastIndex(bs, []byte("\n")); lastLF >= 0 {
		lastLine := bytes.TrimFunc(bytes.TrimSpace(bs[lastLF+1:]), func(r rune) bool {
			return r == 0 || unicode.IsSpace(r)
		})
		if bytes.HasPrefix(lastLine, []byte("#")) {

			// Remote Management Console
			// login: netscreen
			// password:
			//  ### Login failed                 <-- 这里
			// Remote Management Console
			// login: netscreen
			// password:
			// GGQQ-10.51.71-SSG550M(M)->

			//log.Println("==================== skip", string(lastLine))
			return true
		}

		if bytes.Equal(delim, []byte("$")) && !bytes.Equal(delim, lastLine) {
			if lastLine[0] == '[' {
				lastLine = bytes.TrimFunc(bytes.TrimSuffix(lastLine, []byte("$")), func(r rune) bool {
					return r == 0 || unicode.IsSpace(r)
				})
				if lastLine[len(lastLine)-1] == ']' {
					// [mfk]$
					return false
				}
				// fmt.Println("11 SkipHits ---- true")
			}
			// AAA$AAA

			// fmt.Println("SkipHits ---- true")
			return true
		}
		// 	log.Println("==================== hit", string(lastLine))
		// } else {
		// 	log.Println("==================== hit", string(bs))
	}
	return false
}
