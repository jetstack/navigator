package errors

type transientError error

func IsTransient(err error) bool {
	_, ok := err.(transientError)
	return ok
}

func Transient(err error) error {
	return transientError(err)
}
