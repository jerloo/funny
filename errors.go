package funny

import "fmt"

type FunnyRuntimeError struct {
	Postion Position
	Msg     string
}

func (fre *FunnyRuntimeError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s\n", fre.Postion.File, fre.Postion.Line, fre.Postion.Col, fre.Msg)
}

// P panic
func P(keyword string, pos Position) error {
	return &FunnyRuntimeError{
		Msg:     keyword,
		Postion: pos,
	}
}
