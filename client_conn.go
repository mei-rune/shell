package shell

// type Observable struct {
// 	Except [][]byte
// 	Cb     func(w io.Writer) error
// }

// func (ob *Observable) Then(cb func(w io.Writer) error) Observable {
// 	return Observable{Except: ob.Except, Cb: cb}
// }

// func When(except ...string) Observable {
// 	return Observable{Except: except}
// }

// type ClientConn interface {
// 	Execute(cmd string, ob Observable)
// }

// type ClientConn struct {
// 	w       io.Writer
// 	matcher matchWriter
// }

// func (c *ClientConn) Expect(buf *bytes.Buffer, timeout time.Duration, delims [][]byte) (int, error) {
// 	prompts := make([][]byte, 0, len(more_prompts)+1+len(delims))
// 	prompts[0] = c.prompt
// 	copy(prompts[1:], more_prompts)
// 	if len(delims) > 0 {
// 		copy(prompts[1+len(more_prompts):], delims)
// 	}
// 	c.expect(buf, delims, func(idx int, prompt []byte) {

// 	})
// }

// func (c *ClientConn) expect(buf *bytes.Buffer, timeout time.Duration, delims [][]byte,
// 	cb func(idx int, prompt []byte)) {
// 	c.matcher.Reset(buf, delims, cb)
// }

// func (c *ClientConn) Sendln(buf *bytes.Buffer, s []byte) error {
// 	if !bytes.HasSuffix(s, []byte("\n")) {
// 		if e := c.Send(buf, s); nil != e {
// 			return e
// 		}
// 		return c.Send(buf, []byte("\n"))
// 	}
// 	return c.Send(buf, s)
// }

// func (c *ClientConn) Send(buf *bytes.Buffer, s []byte) error {
// 	if nil != buf {
// 		buf.Write(s)
// 	}

// 	_, err := c.w.Write(s)
// 	return err
// }
