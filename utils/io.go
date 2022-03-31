package utils

import (
	"github.com/brawaru/marct/globstate"
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"io"
)

func DClose(c io.Closer) {
	if closeErr := c.Close(); closeErr != nil {
		if globstate.VerboseLogs {
			println(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": closeErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "log.verbose.io-close-failed",
					Other: "failed to close stream: {{ .Error }}",
				},
			}))
		}
	}
}
