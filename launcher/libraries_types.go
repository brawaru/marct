package launcher

import (
	"strconv"
	"strings"

	"github.com/brawaru/marct/maven"
)

type Artifact struct {
	Download
	Path string `json:"path" validate:"required"`
}

type LibraryDownloads struct {
	Artifact    *Artifact            `json:"artifact,omitempty"`
	Classifiers map[string]*Artifact `json:"classifiers,omitempty"`
}

type ExtractOptions struct {
	Exclude []string `json:"exclude,omitempty"`
}

func (e *ExtractOptions) Include(name string) bool {
	if e != nil {
		for _, s := range e.Exclude {
			if strings.HasPrefix(name, s) {
				return false
			}
		}
	}

	return true
}

type Library struct {
	Downloads   *LibraryDownloads `json:"downloads,omitempty"`
	Coordinates maven.Coordinates `json:"name"`
	URL         *string           `json:"url,omitempty"`
	Rules       Rules             `json:"rules,omitempty"`
	Extract     *ExtractOptions   `json:"extract,omitempty"`
	Natives     map[string]string `json:"natives,omitempty"`
}

// GetMatchingNatives returns the natives matching the current running OS and arch.
func (l *Library) GetMatchingNatives() *Artifact {
	if l == nil {
		panic("library is nil")
	}

	if l.Natives == nil || l.Downloads == nil || l.Downloads.Classifiers == nil {
		return nil
	}

	if classifierName, has := l.Natives[currentOS()]; has {
		c := strings.Replace(classifierName, "${arch}", strconv.Itoa(strconv.IntSize), 1)
		return l.Downloads.Classifiers[c]
	}

	return nil
}
