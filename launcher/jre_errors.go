package launcher

import "fmt"

type javaUnavailableErrno int

const (
	_                                          = iota
	ErrSystemUnsupported  javaUnavailableErrno = iota
	ErrVersionUnavailable                      = iota
)

type JavaUnavailableError struct {
	System  string
	Version string
	Errno   javaUnavailableErrno
}

func (j *JavaUnavailableError) Error() (text string) {
	if j.Errno == ErrSystemUnsupported {
		text = fmt.Sprintf("no runtimes support system \"%s\"", j.System)
	} else if j.Errno == ErrVersionUnavailable {
		text = fmt.Sprintf("version \"%s\" is not supported on system \"%s\"", j.Version, j.System)
	}

	return
}

func (j *JavaUnavailableError) Is(target error) bool {
	t, ok := target.(*JavaUnavailableError)

	return ok &&
		(t.System == "" || j.System == t.System) &&
		(t.Version == "" || j.Version == t.Version) &&
		(t.Errno == 0 || j.Errno == t.Errno)
}
