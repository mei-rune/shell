package shell

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"time"
	"unicode"
)

const (
	CR = byte('\r')
	LF = byte('\n')
)

const (
	// SE                  240    End of subnegotiation parameters.
	cmdSE = 240
	// NOP                 241    No operation.
	cmdNOP = 241
	// Data Mark           242    The data stream portion of a Synch.
	//                            This should always be accompanied
	//                            by a TCP Urgent notification.
	cmdData = 242

	// Break               243    NVT character BRK.
	cmdBreak = 243
	// Interrupt Process   244    The function IP.
	cmdIP = 244
	// Abort output        245    The function AO.
	cmdAO = 245
	// Are You There       246    The function AYT.
	cmdAYT = 246
	// Erase character     247    The function EC.
	cmdEC = 247
	// Erase Line          248    The function EL.
	cmdEL = 248
	// Go ahead            249    The GA signal.
	cmdGA = 249
	// SB                  250    Indicates that what follows is
	//                            subnegotiation of the indicated
	//                            option.
	cmdSB = 250 // FA

	// WILL (option code)  251    Indicates the desire to begin
	//                            performing, or confirmation that
	//                            you are now performing, the
	//                            indicated option.
	cmdWill = 251 // FB
	// WON'T (option code) 252    Indicates the refusal to perform,
	//                            or continue performing, the
	//                            indicated option.
	cmdWont = 252 // FC
	// DO (option code)    253    Indicates the request that the
	//                            other party perform, or
	//                            confirmation that you are expecting
	//                            the other party to perform, the
	//                            indicated option.
	cmdDo = 253 // FD
	// DON'T (option code) 254    Indicates the demand that the
	//                            other party stop performing,
	//                            or confirmation that you are no
	//                            longer expecting the other party
	//                            to perform, the indicated option.
	cmdDont = 254 // FE

	// IAC                 255    Data Byte 255.
	cmdIAC = 255 //FF

)

const (

	// 1(0x01)    回显(echo)
	optEcho = 1
	// 3(0x03)    抑制继续进行(传送一次一个字符方式可以选择这个选项)
	optSuppressGoAhead = 3
	// 24(0x18)   终端类型
	optWndType = 24
	// 31(0x1F)   窗口大小
	optWndSize = 31
	// 32(0x20)   终端速率
	optRate = 32

// 33(0x21)   远程流量控制
// 34(0x22)   行方式
// 36(0x24)   环境变量
)

// Conn implements net.Conn interface for Telnet protocol plus some set of
// Telnet specific methods.
type Telnet struct {
	nconn net.Conn

	columns, rows byte
	w             io.Writer
	r             *bufio.Reader

	unixWriteMode bool

	cliSuppressGoAhead bool
	cliEcho            bool

	errc chan error
}

func NewTelnet(conn net.Conn) *Telnet {
	return NewTelnet2(conn, conn, conn)
}

func NewTelnet2(conn net.Conn, w io.Writer, r io.Reader) *Telnet {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}
	return &Telnet{
		nconn:   conn,
		columns: 255,
		rows:    255,
		w:       w,
		r:       bufio.NewReaderSize(r, 256),
	}
}

func TelnetWrap(c *Telnet, tees, teec io.Writer) *ConnWrapper {
	c.errc = make(chan error, 1)
	p := MakePipe(2048)
	go func() {

		var a [1]byte
		// 请注意这里不能用 io.Copy()
		for {
			b, err := c.ReadByte()
			if err != nil {
				c.errc <- err
				close(c.errc)

				p.CloseWithError(err)
				break
			}

			if err = p.WriteByte(b); err != nil {
				c.errc <- err
				close(c.errc)

				p.CloseWithError(err)
				break
			}

			if tees != nil {
				a[0] = b
				tees.Write(a[:])
			}
		}
	}()

	var w io.Writer = c
	if teec != nil {
		w = MultWriters(w, teec)
	}

	return &ConnWrapper{
		session:         c,
		w:               w,
		r:               p,
		readByte:        p,
		drainto:         p,
		setReadDeadline: p,
		// setWriteDeadline: p,
	}
}

func DialTelnet(network, addr string) (*Telnet, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return NewTelnet(conn), nil
}

func DialTelnetTimeout(network, addr string, timeout time.Duration) (*Telnet, error) {
	conn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	return NewTelnet(conn), nil
}

func (c *Telnet) Close() error {
	err := c.nconn.Close()

	if c.errc != nil {
		e, _ := <-c.errc

		if err == nil {
			err = e
		}
	}
	return err
}

// SetUnixWriteMode sets flag that applies only to the Write method.
// If set, Write converts any '\n' (LF) to '\r\n' (CR LF).
func (c *Telnet) SetUnixWriteMode(uwm bool) {
	c.unixWriteMode = uwm
}

func (c *Telnet) do(option byte) error {
	//log.Println("do:", option)
	_, err := c.w.Write([]byte{cmdIAC, cmdDo, option})
	return err
}

func (c *Telnet) dont(option byte) error {
	//log.Println("dont:", option)
	_, err := c.w.Write([]byte{cmdIAC, cmdDont, option})
	return err
}

func (c *Telnet) will(option byte) error {
	//log.Println("will:", option)
	_, err := c.w.Write([]byte{cmdIAC, cmdWill, option})
	return err
}

func (c *Telnet) wont(option byte) error {
	//log.Println("wont:", option)
	_, err := c.w.Write([]byte{cmdIAC, cmdWont, option})
	return err
}

func (c *Telnet) sub(opt byte, data ...byte) error {
	if _, err := c.w.Write([]byte{cmdIAC, cmdSB, opt}); err != nil {
		return err
	}
	if _, err := c.Write(data); err != nil {
		return err
	}
	_, err := c.w.Write([]byte{cmdIAC, cmdSE})
	return err
}

func (c *Telnet) deny(cmd, opt byte) (err error) {
	switch cmd {
	case cmdDo:
		err = c.wont(opt)
	case cmdDont:
		// nop
	case cmdWill, cmdWont:
		err = c.dont(opt)
	}
	return
}

func (c *Telnet) subneg() error {

	// Read an option
	o, err := c.r.ReadByte()
	if err != nil {
		return err
	}
	switch o {
	case optWndType:
		for {
			b, err := c.r.ReadByte()
			if err != nil {
				return err
			}
			if b == cmdIAC {
				if b, err = c.r.ReadByte(); err != nil {
					return err
				}
				if b == cmdSE {
					break
				}
			}
		}
		return c.sub(o, 0, 'X', 'T', 'E', 'R', 'M')
	}

	for {
		if b, err := c.r.ReadByte(); err != nil {
			return err
		} else if b == cmdIAC {
			if b, err = c.r.ReadByte(); err != nil {
				return err
			} else if b == cmdSE {
				return nil
			}
		}
	}
}

func (c *Telnet) cmd(cmd byte) error {
	switch cmd {
	case cmdGA:
		return nil
	case cmdDo, cmdDont, cmdWill, cmdWont:
		// Process cmd after this switch.
	case cmdSB:
		return c.subneg()
	case cmdSE:
		return nil
	default:
		fmt.Println("unknown command:", cmd)
		return nil //fmt.Errorf("unknown command: %d", cmd)
	}
	// Read an option
	o, err := c.r.ReadByte()
	if err != nil {
		return err
	}
	//log.Println("received cmd:", cmd, o)
	switch o {
	case optEcho:
		// Accept any echo configuration.
		switch cmd {
		case cmdDo:
			if !c.cliEcho {
				c.cliEcho = true
				err = c.will(o)
			}
		case cmdDont:
			if c.cliEcho {
				c.cliEcho = false
				err = c.wont(o)
			}
		case cmdWill:
			err = c.do(o)
		case cmdWont:
			err = c.dont(o)
		}
	case optSuppressGoAhead:
		// We don't use GA so can allways accept every configuration
		switch cmd {
		case cmdDo:
			if !c.cliSuppressGoAhead {
				c.cliSuppressGoAhead = true
				err = c.will(o)
			}
		case cmdDont:
			if c.cliSuppressGoAhead {
				c.cliSuppressGoAhead = false
				err = c.wont(o)
			}
		case cmdWill:
			err = c.do(o)
		case cmdWont:
			err = c.dont(o)

		}
	case optWndSize: //optNAWS:
		if cmd != cmdDo {
			err = c.deny(cmd, o)
			break
		}
		if err = c.will(o); err != nil {
			break
		}
		// Reply with max window size: 65535x65535
		err = c.sub(o, 0, 255, 0, 255)

	//case optWndSize:
	//	if cmd == cmdDo {
	//		_, err = c.w.Write([]byte{cmdIAC, cmdSB, optWndSize, 0, c.columns, 0, c.rows, cmdIAC, cmdSE})
	//	}
	case optWndType:
		if cmd != cmdDo {
			err = c.deny(cmd, o)
			break
		}
		if err = c.will(o); err != nil {
			break
		}
		// Reply with max window size: 65535x65535
		// err = c.sub(o, 0, 255, 0, 255)

	default:
		// Deny any other option
		err = c.deny(cmd, o)
	}
	return err
}

func (c *Telnet) tryReadByte() (b byte, retry bool, err error) {
	b, err = c.r.ReadByte()
	if err != nil || b != cmdIAC {
		return
	}
	b, err = c.r.ReadByte()
	if err != nil {
		return
	}
	if b != cmdIAC {
		err = c.cmd(b)
		if err != nil {
			return
		}
		retry = true
	}
	return
}

// SetEcho tries to enable/disable echo on server side. Typically telnet
// servers doesn't support this.
func (c *Telnet) SetEcho(echo bool) error {
	if echo {
		return c.do(optEcho)
	}
	return c.dont(optEcho)
}

// ReadByte works like bufio.ReadByte
func (c *Telnet) ReadByte() (b byte, err error) {
	retry := true
	for retry && err == nil {
		b, retry, err = c.tryReadByte()
	}
	return
}

// ReadRune works like bufio.ReadRune
func (c *Telnet) ReadRune() (r rune, size int, err error) {
loop:
	r, size, err = c.r.ReadRune()
	if err != nil {
		return
	}
	if r != unicode.ReplacementChar || size != 1 {
		// Properly readed rune
		return
	}
	// Bad rune
	err = c.r.UnreadRune()
	if err != nil {
		return
	}
	// Read telnet command or escaped IAC
	_, retry, err := c.tryReadByte()
	if err != nil {
		return
	}
	if retry {
		// This bad rune was a begining of telnet command. Try read next rune.
		goto loop
	}
	// Return escaped IAC as unicode.ReplacementChar
	return
}

// Read is for implement an io.Reader interface
func (c *Telnet) Read(buf []byte) (int, error) {
	var n int
	for n < len(buf) {
		b, err := c.ReadByte()
		if err != nil {
			return n, err
		}
		//log.Printf("char: %d %q", b, b)
		buf[n] = b
		n++
		if c.r.Buffered() == 0 {
			// Try don't block if can return some data
			break
		}
	}
	return n, nil
}

// // ReadBytes works like bufio.ReadBytes
// func (c *Telnet) ReadBytes(delim byte) ([]byte, error) {
// 	var line []byte
// 	for {
// 		b, err := c.ReadByte()
// 		if err != nil {
// 			return nil, err
// 		}
// 		line = append(line, b)
// 		if b == delim {
// 			break
// 		}
// 	}
// 	return line, nil
// }

// // SkipBytes works like ReadBytes but skips all read data.
// func (c *Telnet) SkipBytes(delim byte) error {
// 	for {
// 		b, err := c.ReadByte()
// 		if err != nil {
// 			return err
// 		}
// 		if b == delim {
// 			break
// 		}
// 	}
// 	return nil
// }

// // ReadString works like bufio.ReadString
// func (c *Telnet) ReadString(delim byte) (string, error) {
// 	bytes, err := c.ReadBytes(delim)
// 	return string(bytes), err
// }

// func (c *Telnet) readUntil(read bool, delims ...[]byte) ([]byte, int, error) {
// 	if len(delims) == 0 {
// 		return nil, 0, nil
// 	}
// 	p := make([][]byte, len(delims))
// 	for i, s := range delims {
// 		if len(s) == 0 {
// 			return nil, 0, nil
// 		}
// 		p[i] = s
// 	}
// 	var line []byte
// 	for {
// 		b, err := c.ReadByte()
// 		if err != nil {
// 			return nil, 0, err
// 		}
// 		if read {
// 			line = append(line, b)
// 		}
// 		for i, s := range p {
// 			if s[0] == b {
// 				if len(s) == 1 {
// 					return line, i, nil
// 				}
// 				p[i] = s[1:]
// 			} else {
// 				//idx := crossingMatch(p[i], delims[i])
// 				p[i] = delims[i] //[idx:]
// 			}
// 		}
// 	}
// 	panic(nil)
// }

// // ReadUntilIndex reads from connection until one of delimiters occurs. Returns
// // read data and an index of delimiter or error.
// func (c *Telnet) ReadUntilIndex(delims ...[]byte) ([]byte, int, error) {
// 	return c.readUntil(true, delims...)
// }

// // ReadUntil works like ReadUntilIndex but don't return a delimiter index.
// func (c *Telnet) ReadUntil(delims ...[]byte) ([]byte, error) {
// 	d, _, err := c.readUntil(true, delims...)
// 	return d, err
// }

// // SkipUntilIndex works like ReadUntilIndex but skips all read data.
// func (c *Telnet) SkipUntilIndex(delims ...[]byte) (int, error) {
// 	_, i, err := c.readUntil(false, delims...)
// 	return i, err
// }

// func (c *Telnet) Expect(delims [][]byte) (int, error) {
// 	_, i, err := c.readUntil(false, delims...)
// 	return i, err
// }

// // SkipUntil works like ReadUntil but skips all read data.
// func (c *Telnet) SkipUntil(delims ...[]byte) error {
// 	_, _, err := c.readUntil(false, delims...)
// 	return err
// }

// func (c *Telnet) Send(buf []byte) error {
// 	for len(buf) > 0 {
// 		n, err := c.Write(buf)
// 		if err != nil {
// 			return err
// 		}
// 		buf = buf[n:]
// 	}
// 	return nil
// }

// Write is for implement an io.Writer interface
func (c *Telnet) Write(buf []byte) (int, error) {
	search := "\xff"
	if c.unixWriteMode {
		search = "\xff\n"
	}
	var (
		n   int
		err error
	)
	for len(buf) > 0 {
		var k int
		i := bytes.IndexAny(buf, search)
		if i == -1 {
			k, err = c.w.Write(buf)
			n += k
			break
		}
		k, err = c.w.Write(buf[:i])
		n += k
		if err != nil {
			break
		}
		switch buf[i] {
		case LF:
			k, err = c.w.Write([]byte{CR, LF})
		case cmdIAC:
			k, err = c.w.Write([]byte{cmdIAC, cmdIAC})
		}
		n += k
		if err != nil {
			break
		}
		buf = buf[i+1:]
	}
	return n, err
}
