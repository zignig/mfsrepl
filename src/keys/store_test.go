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
	ks.NewKey()
	ks.Close()
	t.Errorf("FAIL")

}
