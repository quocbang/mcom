package pda

import (
	"fmt"
	"strings"
)

// RootStringConvert convert to xml format.
func RootStringConvert(resp string) string {
	return strings.ReplaceAll(strings.ReplaceAll(resp, "&lt;", "<"), "&gt;", ">")
}

func newErrorf(url, msg string, args ...interface{}) error {
	return fmt.Errorf("pda web service-"+url+": "+msg, args...)
}
