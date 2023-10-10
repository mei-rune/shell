package shell

import (
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	PlinkPath = "runtime_env/putty/plink.exe"
	binDir    string

	IsWindows = runtime.GOOS == "windows"
)

func SetPlinkPath(s string) {
	PlinkPath = s
}

func init() {
	pa, _ := os.Executable()
	binDir = filepath.Dir(pa)
}

func init() {
	if plinkPath := os.Getenv("mei_plink"); plinkPath != "" {
		PlinkPath = plinkPath
	}

	if fi, err := os.Stat(PlinkPath); err != nil && os.IsNotExist(err) {
		var files []string
		if IsWindows {
			files = []string{
				filepath.Join(binDir, "plink.exe"),
				filepath.Join(binDir, "runtime_env\\putty\\plink.exe"),
				filepath.Join(binDir, "..\\runtime_env\\putty\\plink.exe"),
				"C:\\Program Files\\mei\\runtime_env\\putty\\plink.exe",


				filepath.Join(binDir, "plink_old.exe"),
				filepath.Join(binDir, "runtime_env\\putty\\plink_old.exe"),
				filepath.Join(binDir, "..\\runtime_env\\putty\\plink_old.exe"),
				"C:\\Program Files\\mei\\runtime_env\\putty\\plink_old.exe",
			}
		} else {
			files = []string{
				filepath.Join(binDir, "plink"),
				filepath.Join(binDir, "runtime_env/putty/plink"),
				filepath.Join(binDir, "../runtime_env/putty/plink"),
				"/usr/local/tpt/runtime_env/putty/plink",


				filepath.Join(binDir, "plink_old"),
				filepath.Join(binDir, "runtime_env/putty/plink_old"),
				filepath.Join(binDir, "../runtime_env/putty/plink_old"),
				"/usr/local/tpt/runtime_env/putty/plink_old",
			}
		}
		for _, pa := range files {
			if fi, err = os.Stat(pa); err == nil && !fi.IsDir() {
				PlinkPath = pa
				break
			}
		}
	}
}

type PlinkClient struct {
	cmd *exec.Cmd
	ConnWrapper
}

func (c *PlinkClient) Close() error {
	if e := c.cmd.Process.Kill(); nil != e {
		return e
	}

	c.ConnWrapper.Close()
	return c.cmd.Wait()
}

var tmpseed = time.Now().Unix()

func ConnectPlink(host, username, password, privateKey string, sWriter, cWriter io.Writer) (*PlinkClient, error) {
	// if privateKey != "" {
	// 	return nil, errors.New("兼容模式不支持 证书登录")
	// }
	p := MakePipe(2048)
	address, port, err := net.SplitHostPort(host)

	var cmd *exec.Cmd

	if privateKey != "" {
		filename := filepath.Join(os.TempDir(), strconv.FormatInt(atomic.AddInt64(&tmpseed, 1), 10))
		err = ioutil.WriteFile(filename, []byte(privateKey), 0600)
		if err != nil {
			return nil, err
		}
		defer os.Remove(filename)

		if err != nil {
			cmd = exec.Command(PlinkPath, "-t", username+"@"+host, "-i", filename)
		} else {
			cmd = exec.Command(PlinkPath, "-t", username+"@"+address, "-P", port, "-i", filename)
		}
	} else {
		if err != nil {
			cmd = exec.Command(PlinkPath, "-t", username+"@"+host)
		} else {
			cmd = exec.Command(PlinkPath, "-t", username+"@"+address, "-P", port)
		}
	}

	if sWriter != nil {
		cmd.Stderr = MultWriters(p, sWriter)
	} else {
		cmd.Stderr = p //MultWriters(w, os.Stdout)
	}
	cmd.Stdout = cmd.Stderr

	stdin, e := cmd.StdinPipe()
	if nil != e {
		return nil, e
	}
	if e := cmd.Start(); nil != e {
		return nil, e
	}

	if cWriter != nil {
		cWriter = MultWriters(stdin, cWriter)
	} else {
		cWriter = stdin
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			p.CloseWithError(err)
		}
	}()

	pClient := &PlinkClient{
		cmd: cmd,
	}
	pClient.Init(closeFunc(func() error {
		if e := cmd.Process.Kill(); nil != e {
			return e
		}
		return nil
	}), cWriter, p)
	return pClient, nil
}
