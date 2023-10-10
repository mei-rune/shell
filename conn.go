package shell

import (
	"context"
	"io"
	"time"
)

type SendPasswordWriter interface {
	SendPassword(s []byte) error
}

type Conn interface {
	io.ReadWriteCloser

	SetReadDeadline(t time.Duration) error
	SetWriteDeadline(t time.Duration) error

	SetTeeWriter(w io.Writer) context.CancelFunc
	SetTeeReader(w io.Writer) context.CancelFunc
	SetTeeOutput(w io.Writer) context.CancelFunc

	UseCRLF()
	Send([]byte) error
	Sendln([]byte) error
	SendPasswordWriter
	DrainOff() (int, error)
	Expect([][]byte) (int, []byte, error)
}

type DoFunc func(conn Conn, bs []byte, idx int) (bool, error)
