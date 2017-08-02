package keys

import (
	"fmt"
	"testing"
)


func TestPeer( t *testing.T ){
    p := NewPeer("keys",logger)
    r := p.st.GetRand(5)
    fmt.Println("rand",r)
	t.Errorf("FAIL")
}

func testStore(t *testing.T) {
	ks, err := NewKeyStore("keys")
	if err != nil {
		fmt.Println(err)
	}
	k, err := ks.NewLocalKey()
	fmt.Println(k, err)
	ks.Save(k)
	//data, err := k.Encode()
	//fmt.Println(data, err)
	//rk, err := Decode(data)
	//fmt.Println(rk, err)
	//ks.Insert(k, "local")
	//local, _ := ks.ListKeys("local")
	//fmt.Println(local)
	blob, _ := k.MakeSigned()
	//blob.Data[90] = 0x0FF
	//err = blob.Check()
	//if err == nil {
	//	fmt.Println("Valid Sig BLOB")
	//}
	//fmt.Println(err)
	ks.PutPublic(blob, "public")
	public, _ := ks.ListKeys("public")
	fmt.Println(public)
    list,err := ks.ListKeys("public")
    fmt.Println(list,err)
	ks.Close()
	t.Errorf("FAIL")

}
