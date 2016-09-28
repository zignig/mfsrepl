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
	//fmt.Println(k, err)
	//data, err := k.Encode()
	//fmt.Println(data, err)
	//rk, err := Decode(data)
	//fmt.Println(rk, err)
	ks.Insert(k, "local")
	ks.ListKeys("local")
	blob, _ := k.MakeSigned()
	//blob.Data[90] = 0x0FF
	err = blob.Check()
	if err == nil {
		fmt.Println("Valid Sig BLOB")
	}
	fmt.Println(err)
	ks.PutPublic(blob, "public")
	ks.ListKeys("public")
	ks.Close()
	t.Errorf("FAIL")

}
