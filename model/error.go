package model

// HTTPError allows a plain Go error to be associated with an HTTP status code.
type HTTPError struct {
	Err            error
	HTTPStatusCode int
}

func (e HTTPError) Error() string {
	return e.Err.Error()
}
