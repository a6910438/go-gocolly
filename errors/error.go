package errors

type error struct {
	Code code
	Msg  string
}

func (e *error) Error() string {
	return e.Msg
}

func newError(c code) *error {
	return &error{
		Code: c,
		Msg:  c.String(),
	}
}

var (
	ErrRecordNotFound = newError(errRecordNotFound)
)
