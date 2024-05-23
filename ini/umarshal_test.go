package ini

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNewDecoder(t *testing.T) {
	var tmp struct {
		A    string `ini:"a"`
		Test struct {
			K string `ini:"k"`
		} `ini:"test"`
	}
	a := NewDecoder(bytes.NewReader([]byte("a=b\n[test]\nk=v\n;l=p")))
	err := a.Decode(&tmp)
	fmt.Println(err, tmp)
	tmp.A = "3"
	fmt.Println(tmp)
	a.Show()
}
