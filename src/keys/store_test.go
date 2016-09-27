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
	blob, _ := k.MakeSigned()
	fmt.Println(blob)
	//blob.Data[3] = 0x23
	err = blob.Check()
	if err == nil {
		fmt.Println("Valid Sig BLOB")
	}
	fmt.Println(err)
	//sk, err := blob.Encode()
	//fmt.Println(string(sk), err)
	ks.Close()
	t.Errorf("FAIL")

}
