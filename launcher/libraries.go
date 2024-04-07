package launcher

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/brawaru/marct/globstate"
	"github.com/brawaru/marct/launcher/download"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/maven"
	"github.com/brawaru/marct/utils/terrgroup"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// LibrariesPath returns expected path for the libraries to be placed in.
func (w *Instance) LibrariesPath() string {
	return filepath.Join(w.Path, "libraries")
}

// LibraryPath returns expected path for the library to be placed in.
func (w *Instance) LibraryPath(coords maven.Coordinates) string {
	return filepath.Join(w.LibrariesPath(), coords.Path(os.PathSeparator))
}

func hasEmptyPath(u url.URL) bool {
	return u.Path == "" || u.Path == "/"
}

func (w *Instance) DownloadLibrary(library *Library) error {
	if library.URL != nil || library.Downloads == nil {
		dlPath := w.LibraryPath(library.Coordinates)

		var mavenServer string
		if library.URL == nil {
			mavenServer = mojangMavenServer
		} else {
			mavenServer = *library.URL
		}

		src, urlErr := url.Parse(mavenServer)

		if urlErr != nil {
			return urlErr
		}

		if hasEmptyPath(*src) {
			src.Path = library.Coordinates.Path('/')
		}

		if err := download.From(src, dlPath, download.WithRemoteSHA1(), download.WithRemoteMD5()); err != nil {
			return fmt.Errorf("download %s: %s", dlPath, err)
		}
	} else {
		artifact := library.Downloads.Artifact

		if artifact != nil {
			dest := filepath.Join(w.LibrariesPath(), filepath.FromSlash(artifact.Path))

			if err := download.FromURL(artifact.URL, dest, download.WithSHA1(artifact.SHA1)); err != nil {
				return fmt.Errorf("download %s: %s", artifact.URL, err)
			}
		}
	}

	if natives := library.GetMatchingNatives(); natives != nil {
		dest := filepath.Join(w.LibrariesPath(), filepath.FromSlash(natives.Path))

		if err := download.FromURL(natives.URL, dest, download.WithSHA1(natives.SHA1)); err != nil {
			return fmt.Errorf("download native %s: %s", natives.URL, err)
		}
	}

	return nil
}

func (w *Instance) DownloadLibraries(libraries []Library) error {
	g, _ := terrgroup.New(8)

	for _, l := range libraries {
		library := l

		g.Go(func() error {
			if library.Rules != nil && !library.Rules.Matches() {
				if globstate.VerboseLogs {
					println(locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"Library": library.Coordinates.String(),
						},
						DefaultMessage: &i18n.Message{
							ID:    "log.verbose.library-skipped-over-rules",
							Other: "skipping library {{ .Library }} since it does not match rules",
						},
					}))
				}

				return nil
			}

			if dlErr := w.DownloadLibrary(&library); dlErr == nil {
				if globstate.VerboseLogs {
					println(locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"Library": library.Coordinates.String(),
						},
						DefaultMessage: &i18n.Message{
							ID:    "log.verbose.downloaded-library",
							Other: "downloaded library {{ .Library }}",
						},
					}))
				}
			} else {
				return fmt.Errorf("download libraries: %w", dlErr)
			}

			return nil
		})
	}

	return g.Wait()
}
