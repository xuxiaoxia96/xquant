package labs

import (
	"fmt"
	"reflect"
	"testing"

	"xquant/factors"
)

func TestReflect(t *testing.T) {
	//lazyInit()
	fia := reflect.TypeOf((*factors.Feature)(nil)).Elem()
	v := FindImplements(fia)
	fmt.Println(v)
}
