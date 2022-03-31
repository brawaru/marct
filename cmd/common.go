package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/xbox"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func keyringOpenPrompt(req string) (resp string, err error) {
	// FIXME: translate request
	err = survey.AskOne(&survey.Password{
		Message: locales.Translate(&i18n.Message{
			ID:    "cli.prompts.open-keyring-password.message",
			Other: "Enter password to unlock keyring",
		}),
		Help: locales.Translate(&i18n.Message{
			ID: "cli.prompts.open-keyring-password.help",
			Other: "Keyring is a secure storage in your system to store important credentials.\n" +
				"We use it to store your Microsoft account keys since they cannot be stored in regular files.",
		}),
	}, &resp)

	return
}

func xboxDeviceAuthPrompt(devAuth xbox.DeviceAuthResponse) error {
	println(locales.TranslateUsing(&i18n.LocalizeConfig{
		TemplateData: map[string]string{
			"Link": devAuth.VerificationURI,
			"Code": devAuth.UserCode,
		},
		// this is the exact thing Microsoft sends, but made into a translatable string by us
		DefaultMessage: &i18n.Message{
			ID:    "cli.prompts.xbox-login-how-to",
			Other: "To sign in, use a web browser to open the page {{ .Link }} and enter the code {{ .Code }} to authenticate.",
		},
	}))

	return nil
}
