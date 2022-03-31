package launcher

import "fmt"

type DownloadUnavailableError struct {
	Download string
}

func (d *DownloadUnavailableError) Error() string {
	return fmt.Sprintf("download %s is not available", d.Download)
}

func (d *DownloadUnavailableError) Is(target error) bool {
	t, ok := target.(*DownloadUnavailableError)
	return ok && (t.Download == "" || d.Download == t.Download)
}
