package funny

import "fmt"

type FunnyContent interface {
	String() string
}

type FunnyRuntimeError struct {
	FunnyContent FunnyContent
	Msg          string
}

func (fre *FunnyRuntimeError) Error() string {
	if v, ok := fre.FunnyContent.(Token); ok {
		return fmt.Sprintf("funny runtime error: %s %d:%d", fre.Msg, v.Position.Line, v.Position.Col)
	}
	if v, ok := fre.FunnyContent.(Statement); ok {
		return fmt.Sprintf("funny runtime error: %s %d:%d", fre.Msg, v.Position().Line, v.Position().Col)
	}
	return "funny runtime error: unknow"
}

// P panic
func P(keyword string, posContent FunnyContent) error {
	return &FunnyRuntimeError{
		Msg:          keyword,
		FunnyContent: posContent,
	}
}
