package errors

// Public wraps the original error with a new error that has a 
// `Public() string` method that will return a message that is 
// acceptable to display to the public. This error can also be 
// unwrapped using the traditional `errors` package approach. 
func Public(err error, msg string) error{
    return publicError{err, msg}
}

type publicError struct {
    err error
    msg string
}

func (pe publicError) Error() string {
    return pe.err.Error()
}

func (pe publicError) Public() string {
    return pe.msg
}

func (pe publicError) Unwrap() error {
    return pe.err
}
