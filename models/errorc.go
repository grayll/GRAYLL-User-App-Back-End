package models

// ErrorC is a customized error interface
// We usually want errcode beside error message.
// Passing error code back to client for quickly checking error
type ErrorC struct {
	ErrMsg  string
	ErrCode int
}

func (e ErrorC) Error() string {
	return e.ErrMsg
}

func NewError(msg string, code int) *ErrorC {
	return &ErrorC{msg, code}
}
