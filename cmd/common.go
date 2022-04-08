package cmd

import (
	"errors"

	"github.com/99designs/keyring"
	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/utils/slices"
	"github.com/brawaru/marct/xbox"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
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

func keyringBackendSelectPrompt(backends []keyring.BackendType) (resp keyring.BackendType, err error) {
	var options []string
	optionsMappings := make(map[string]keyring.BackendType)
	for _, b := range backends {
		option := string(b)

		switch b {
		case keyring.FileBackend:
			option = locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-keyring-backend.file",
				Other: "File",
			})
		case keyring.WinCredBackend:
			option = locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-keyring-backend.wincred",
				Other: "Windows Credential Manager",
			})
		case keyring.KWalletBackend:
			option = locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-keyring-backend.kwallet",
				Other: "KWallet",
			})
		case keyring.SecretServiceBackend:
			option = locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-keyring-backend.secretservice",
				Other: "Secret Service",
			})
		case keyring.InvalidBackend:
			continue
		case keyring.KeyCtlBackend:
			option = locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-keyring-backend.keyctl",
				Other: "Keyctl",
			})
		case keyring.KeychainBackend:
			option = locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-keyring-backend.keychain",
				Other: "Keychain",
			})
		case keyring.PassBackend:
			option = locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-keyring-backend.pass",
				Other: "Pass",
			})
		}

		options = append(options, option)
		optionsMappings[option] = b
	}

	var selection string

	err = survey.AskOne(&survey.Select{
		Message: locales.Translate(&i18n.Message{
			ID:    "cli.prompts.select-keyring-backend.message",
			Other: "Select keyring backend",
		}),
		Default: options[0],
		Options: options,
		Help: locales.Translate(&i18n.Message{
			ID: "cli.prompts.select-keyring-backend.help",
			Other: "Select the keyring backend to use for storing your credentials.\n" +
				"If you don't know which one to use, use the default one.",
		}),
	}, &selection, survey.WithValidator(func(answer interface{}) error {
		i, ok := answer.(survey.OptionAnswer)
		if !ok {
			return errors.New(locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-keyring-backend.error.survey-validation-invalid-type",
				Other: "Invalid response type.",
			}))
		}

		_, ok = optionsMappings[i.Value]

		if !ok {
			return errors.New(locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-keyring-backend.error.survey-validation-invalid-option",
				Other: "Invalid option.",
			}))
		}

		return nil
	}))

	resp = optionsMappings[selection]

	return
}

// keyringOpenFlow provides a standard CLI flow for opening a keyring.
// It accepts launcher instance as an argument which it uses to open settings
// and read existing backend preference, or, if it's not set, ask the user to
// select one of the available backends and then saves the settings.
// It returns the selected backend or an error if something went wrong.
// Error will be wrapped as a cli.ExitCoder, so there's no need to handle it.
func keyringOpenFlow(instance *launcher.Instance) (keyring.Keyring, error) {
	s, err := instance.OpenSettings()
	if err != nil {
		return nil, cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]any{
				"Error": err.Error(),
			},
			DefaultMessage: &i18n.Message{
				ID:    "cli.flows.open-keyring.error.open-settings",
				Other: "Cannot read your settings: {{ .Error }}",
			},
		}), 1)
	}

	var bt keyring.BackendType

	if s.Keyring.Backend == nil || slices.Includes(keyring.AvailableBackends(), *s.Keyring.Backend) {
		bt, err = launcher.SelectKeyringBackend(keyringBackendSelectPrompt)
		if err != nil {
			return nil, cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]any{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "cli.flows.open-keyring.error.select-backend",
					Other: "Cannot select keyring backend: {{ .Error }}",
				},
			}), 1)
		}

		s.Keyring.Backend = &bt

		if err := s.Save(); err != nil {
			return nil, cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]any{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "cli.flows.open-keyring.error.save-settings",
					Other: "Cannot save your preference: {{ .Error }}",
				},
			}), 1)
		}
	} else {
		bt = *s.Keyring.Backend
	}

	k, err := instance.OpenKeyring(launcher.KeyringOpenOptions{
		PromptFunc: keyringOpenPrompt,
		Backend:    bt,
		PassCmd:    s.Keyring.PassCmd,
		PassDir:    s.Keyring.PassDir,
	})

	if err != nil {
		return nil, cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"Error": err.Error(),
			},
			DefaultMessage: &i18n.Message{
				ID:    "cli.flows.open-keyring.error.open-keyring",
				Other: "Cannot open your keyring: {{ .Error }}",
			},
		}), 1)
	}

	return k, nil
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
