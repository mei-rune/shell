package shell

import (
	"bytes"
	"testing"
)

func TestExpect(t *testing.T) {
	for _, m := range testMatchs {

		p := MakePipe(1024)
		wrapper := MakeConnWrapper(nil, nil, p)

		for _, d := range m.data {
			p.Write([]byte(d))
		}
		p.Close()

		var buf bytes.Buffer
		wrapper.SetTeeReader(&buf)
		matchedIdx, _, err := wrapper.Expect([][]byte{[]byte("adsbadsfsadfs22"), []byte(m.pattern)})
		if err != nil {
			if m.found {
				t.Errorf("\"%v\" match \"%v\" failed, %v", m.data, m.pattern, err)
			}
			continue
		}

		if matchedIdx != 1 {
			t.Errorf("\"%v\" match \"%v\" failed, excepted is %v, actual is %v", m.data, m.pattern, 1, matchedIdx)
			continue
		}

		if !m.found {
			t.Errorf("\"%v\" match \"%v\" failed, excepted is %v, actual is %v", m.data, m.pattern, m.found, true)
			continue
		}

		// if strings.Join(m.data, "") != buf.String() {
		// 	t.Errorf("\"%v\" match \"%v\" failed, result is error", m.data, m.pattern)
		// }
	}
}

// ,
