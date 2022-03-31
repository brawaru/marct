package utils

// TraverseCauses calls a check function on error, which supposed to read the cause error and return it,
// a filter might be added, which would prevent to traversing down, this method will also return if the same
// error as input is returned.
func TraverseCauses(err error, check func(input error) error) error {
	if err == nil {
		return nil
	}

	current := err

	for {
		ret := check(current)

		if ret == nil || current == ret {
			return current
		}

		current = ret
	}
}

func NewWrappedError(message string, cause error) error {
	return &wrappedErr{
		text:  message,
		error: cause,
	}
}

type wrappedErr struct {
	text string
	error
}

func (w *wrappedErr) Error() string {
	return w.text
}

func (w *wrappedErr) Unwrap() error {
	return w.error
}
