package keys

import (
	"fmt"
	"testing"
)

func TestStore(t *testing.T) {
	ks, err := NewKeyStore("./keys.db")
	if err != nil {
		fmt.Println(err)
	}
	k, err := ks.NewLocalKey()
	fmt.Println(k, err)
	ks.Close()
	t.Errorf("FAIL")

}
