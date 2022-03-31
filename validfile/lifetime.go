package validfile

import (
	"fmt"
	"time"
)

type FileExpiredError struct {
	Name     string
	Modified time.Time
	TTL      time.Duration
}

func (e *FileExpiredError) Error() string {
	return fmt.Sprintf("file \"%s\" has expired, was last modified %s, which is longer than %s", e.Name, e.Modified.Format(time.RFC1123), e.TTL.String())
}

func (e *FileExpiredError) Is(target error) bool {
	t, ok := target.(*FileExpiredError)

	return ok == true &&
		(t.Name == "" || t.Name == e.Name) &&
		(t.Modified.Equal(time.Time{}) || t.Modified.Equal(e.Modified)) &&
		(t.TTL == 0 || t.TTL == e.TTL)
}

func NotExpired(name string, ttl time.Duration) error {
	stat, statErr := TryStat(name)

	if statErr != nil {
		return statErr
	}

	modTime := stat.ModTime()

	if time.Since(modTime) > ttl {
		return &FileExpiredError{
			Name:     name,
			Modified: modTime,
			TTL:      ttl,
		}
	}

	return nil
}
