package csy_formatcode_util_test

import (
	"fmt"

	"github.com/front-ck996/csy/utils/csy_formatcode_util"

	"testing"
)

func TestFormatCode(t *testing.T) {
	in := []byte(`
package p
var ()
func f() {
	for _ = range v {
	}
}
`[1:])
	fmt.Println(csy_formatcode_util.FormatCode(in))
}
