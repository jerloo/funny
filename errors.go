package funny

type FunnyContent interface {
	String() string
}

type FunnyRuntimeError struct {
	FunnyContent FunnyContent
	Msg          string
}

func (fre *FunnyRuntimeError) Error() string {
	return "funny runtime error"
}

// P panic
func P(keyword string, posContent FunnyContent) error {
	return &FunnyRuntimeError{
		Msg:          keyword,
		FunnyContent: posContent,
	}
}
