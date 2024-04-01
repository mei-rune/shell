package shell

import (
	"testing"
)

func TestRemoveMore(t *testing.T) {
	excepted := "280 deny udp any any eq 5554\n290 deny udp any any eq 9996\n"

	bs := RemoveCtrlCharByLine([][]byte{[]byte("280 deny udp any any eq 5554"),
		[]byte("---MORE- --\r290 deny udp any any eq 9996")}, 10)
	if string(bs) != excepted {
		t.Error(string(bs))
	}
}

// func TestRemoveCtrlCharByLine(t *testing.T) {

// 	for _, test := range []struct {
// 		contents []string
// 		excepted string
// 	}{
// 		{
// 			contents: []string{"616263640d",
// 				"2d2d206d6f7265202d2d1b5b34324420202020202020202020202020202020202020202020202020",
// 				"20202020202020202020202020202020201b5b343244656667680d",
// 				"2d2d206d6f7265202d2d1b5b34324420202020202020202020202020202020202020202020202020",
// 				"20202020202020202020202020202020201b5b343244696a6b6c6d6e0d",
// 				"72657475726e0d",
// 			},
// 			excepted: "abcd\nefgh\nijklmn\r\nreturn",
// 		},
// 	} {

// 		var a [][]byte
// 		for _, s := range test.contents {
// 			b, _ := hex.DecodeString(s)
// 			a = append(a, b)
// 		}

// 		actual := RemoveCtrlCharByLine(a, 1204)

// 		if test.excepted != string(actual) {
// 			t.Error("excepted is", test.excepted)
// 			t.Error("actual   is", string(actual))
// 		}

// 	}
// }
