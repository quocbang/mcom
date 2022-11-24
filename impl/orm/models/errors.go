package models

import (
	"fmt"
)

type scanBadTypeError struct {
	structName string
	src        interface{}
}

func (e *scanBadTypeError) Error() string {
	return fmt.Sprintf("bad src type [%T] for struct [%s]", e.src, e.structName)
}
