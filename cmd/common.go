package cmd

import (
	"errors"
	"regexp"
	"strings"

	"github.com/99designs/keyring"
	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/utils/slices"
	"github.com/brawaru/marct/xbox"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

type ctxKey string

const (
	accountsStoreKey ctxKey = "accounts"
	instanceKey      ctxKey = "instance"
	workDirKey       ctxKey = "workDir"
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
			Other: "Select credentials storage",
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

	if s.Keyring.Backend == nil || !slices.Includes(keyring.AvailableBackends(), *s.Keyring.Backend) {
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

type selectAccountFlowOptions struct {
	SkipSelected bool    // Whether to skip the selected account.
	Message      *string // Message to display. If empty, the default "Select profile" message will be used.
}

type SelectProfileFlowOption func(*selectAccountFlowOptions)

func WithoutSelected() SelectProfileFlowOption {
	return func(o *selectAccountFlowOptions) {
		o.SkipSelected = true
	}
}

func WithMessage(v string) SelectProfileFlowOption {
	return func(o *selectAccountFlowOptions) {
		o.Message = &v
	}
}

func SelectProfileFlow(p *launcher.Profiles, options ...SelectProfileFlowOption) (*launcher.Profile, error) {
	var o selectAccountFlowOptions
	for _, opt := range options {
		opt(&o)
	}

	if p == nil || len(p.Profiles) == 0 {
		return nil, cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"Command": strings.Join([]string{
					app.Name,
					profileCommand.Name,
					profileCreateCommand.Name,
				}, " "),
			},
			DefaultMessage: &i18n.Message{
				ID:    "cli.flows.select-account.error.no-profiles",
				Other: "No profiles found. Create your first profile using `{{ .Command }}`.",
			},
		}), 1)
	}

	if !o.SkipSelected && p.SelectedProfile != nil {
		p, ok := p.Profiles[*p.SelectedProfile]
		if ok {
			return &p, nil
		}
	}

	var selection string
	var variants []string
	variantMappings := make(map[string]string)
	for i, v := range p.Profiles {
		name := v.Name
		if name == "" {
			name = locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-profile.unnamed",
				Other: "Unnamed profile",
			})
		}

		option := locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"Name":    name,
				"Version": v.LastVersionID,
			},
			DefaultMessage: &i18n.Message{
				ID:    "cli.prompts.select-profile.option",
				Other: "{{ .Name }} ({{ .Version }})",
			},
		})

		variants = append(variants, option)
		variantMappings[option] = i
	}

	var msg string
	if o.Message != nil {
		msg = *o.Message
	} else {
		msg = locales.TranslateUsing(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "cli.prompts.select-profile.message",
				Other: "Select profile",
			},
		})
	}

	err := survey.AskOne(&survey.Select{
		Message: msg,
		Options: variants,
	}, &selection, survey.WithValidator(func(ans interface{}) error {
		i, ok := ans.(survey.OptionAnswer)
		if !ok {
			return errors.New(locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-profile.error.invalid-type",
				Other: "Invalid answer type",
			}))
		}

		_, ok = variantMappings[i.Value]

		if !ok {
			return errors.New(locales.Translate(&i18n.Message{
				ID:    "cli.prompts.select-profile.error.invalid-option",
				Other: "Invalid option",
			}))
		}

		return nil
	}))

	if err != nil {
		return nil, cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"Error": err.Error(),
			},
			DefaultMessage: &i18n.Message{
				ID:    "cli.prompts.select-profile.error.survey-failed",
				Other: "Cannot read your answer: {{ .Error }}",
			},
		}), 1)
	}

	s, ok := p.Profiles[variantMappings[selection]]

	if !ok {
		return nil, cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"Selection": selection,
			},
			DefaultMessage: &i18n.Message{
				ID:    "cli.prompts.select-profile.error.invalid-selection",
				Other: "Invalid selection: {{ .Selection }}",
			},
		}), 1)
	}

	return &s, nil
}

// Regular Expression for checking the valid Minecraft username.
//
// Current rules for usernames are (as per https://help.minecraft.net/hc/en-us/articles/4408950195341):
// - must have 3-16 characters
// - must not have spaces
// - A-Z, a-z, 0-9
// - the only allowed special character is _
//
var minecraftUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,16}$`)

func offlineUsernamePrompt() (string, error) {
	var username string
	err := survey.AskOne(&survey.Input{
		Message: locales.Translate(&i18n.Message{
			ID:    "cli.prompts.offline-username",
			Other: "Username",
		}),
		Help: locales.Translate(&i18n.Message{
			ID:    "cli.prompts.offline-username.help",
			Other: "Enter your username to use for offline account.",
		}),
	}, &username, survey.WithValidator(func(ans interface{}) error {
		i, ok := ans.(string)

		if !ok {
			return errors.New(locales.Translate(&i18n.Message{
				ID:    "cli.prompts.offline-username.error.invalid-type",
				Other: "Invalid answer type",
			}))
		}

		if !minecraftUsernameRegex.MatchString(i) {
			return errors.New(locales.Translate(&i18n.Message{
				ID:    "cli.prompts.offline-username.error.invalid-format",
				Other: "Invalid username",
			}))
		}

		return nil
	}))

	if err != nil {
		return "", cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"Error": err.Error(),
			},
			DefaultMessage: &i18n.Message{
				ID:    "cli.prompts.offline-username.error.survey-failed",
				Other: "Cannot read your answer: {{ .Error }}",
			},
		}), 1)
	}

	return username, nil
}
