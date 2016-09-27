package keys

import (
	"fmt"
	"testing"
)

func TestStore(t *testing.T) {
	ks, err := NewKeyStore("./testing.db")
	if err != nil {
		fmt.Println(err)
	}
	k, err := ks.NewLocalKey()
	fmt.Println(k, err)
	data, err := k.Encode()
	fmt.Println(data, err)
	rk, err := Decode(data)
	fmt.Println(rk, err)
	ks.Insert(k, "local")
	ks.ListKeys("local")
	ks.Close()
	t.Errorf("FAIL")

}
