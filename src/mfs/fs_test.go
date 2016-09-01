package mfs

import (
	"fmt"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	j := NewIPfsfs()
	fmt.Println(j)
	fmt.Println(j.Stat())
	j.Mfs("share")
	t.Errorf("FAIL")

}

func TestDate(t *testing.T) {
	const layout = "/2006/01/02/15/04/"
	n := time.Now()
	fmt.Println(n.Format(layout))

}
