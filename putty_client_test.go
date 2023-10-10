package shell

import (
	"bytes"
	"context"
	_ "net/http/pprof"
	"strings"
	"testing"
	"time"
)

func TestPlinkConnectHpILO(t *testing.T) {
	ctx := context.Background()

	var buf bytes.Buffer

	out := &SafeWriter{
		W: &buf,
	}
	defer func() {
		t.Log(buf.String())

		// time.Sleep(1 * time.Hour)
	}()

	conn, err := ConnectPlink("192.168.1.15", "Administrator", "123456abc", "",
		PasswordWriter(WriteFunc(func(p []byte) (int, error) {
			return out.WriteWithTag("S:", p)
		})),
		PasswordWriter(WriteFunc(func(p []byte) (int, error) {
			return out.WriteWithTag("C:", p)
		})))
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(5 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("Administrator"), nil, []byte("123456abc"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	if string(prompt) != "</>hpiLO->" {
		t.Errorf("want '</>hpiLO->' got %s", prompt)
		return
	}

	output, err := Exec(ctx, conn, prompt, []byte("echo abcd"))
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(string(output), "status_tag=COMMAND PROCESSING FAILED") {
		t.Errorf("want 'status_tag=COMMAND PROCESSING FAILED' got %s", output)
	}
}
