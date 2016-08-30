package mfs

import (
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	j := NewIPfsfs()
	fmt.Println(j)
	fmt.Println(j.Stat())
	j.mfs("share")
	t.Errorf("FAIL")

}
