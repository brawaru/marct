package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var profileSelectCommand = createCommand(&cli.Command{
	Name: "select",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.profile-select.usage",
		Other: "Select default profile to manage",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.profile-select.description",
		Other: "Allows to change default profile used in management or launch commands",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.profile-select.args-usage",
		Other: "[profile identifier]",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value(instanceKey).(*launcher.Instance)
		profiles := ctx.Context.Value(profilesKey).(*launcher.Profiles)

		var profileID string
		if ctx.NArg() == 0 {
			var options []string
			optionsMappings := map[string]string{}
			for id, profile := range profiles.Profiles {
				profileName := profile.Name
				if len(profileName) == 0 {
					profileName = locales.Translate(&i18n.Message{
						ID:    "command.profile-select.unnamed-profile",
						Other: "Unnamed profile",
					})
				}

				option := locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"ID":      id,
						"Name":    profileName,
						"Version": profile.LastVersionID,
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.profile-select.select-option",
						Other: "{{ .ID }}: {{ .Name }} ({{ .Version }})",
					},
				})

				options = append(options, option)
				optionsMappings[option] = id
			}

			var selectedOption string
			if err := survey.AskOne(&survey.Select{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.profile-select.select-message",
					Other: "Select default profile",
				}),
				Options: options,
			}, &selectedOption); err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.profile-select.error.survey-fail",
						Other: "Cannot read your selection: {{ .Error }}",
					},
				}), 1)
			}

			selectedID, ok := optionsMappings[selectedOption]
			if !ok {
				return cli.Exit(locales.Translate(&i18n.Message{
					ID:    "command.profile-select.error.invalid-selection",
					Other: "Selected option does not exist",
				}), 1)
			}

			profileID = selectedID
		} else if ctx.NArg() == 1 {
			profileID = ctx.Args().First()
		} else if ctx.NArg() > 1 {
			return cli.Exit(&i18n.Message{
				ID:    "command.profile-select.error.too-many-arguments",
				Other: "You have provided too many arguments",
			}, 1)
		}

		_, profileExists := profiles.Profiles[profileID]

		if !profileExists {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"ID": profileID,
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.profile-select.error.no-profile-with-id",
					Other: "No profile found with ID {{ .ID }}",
				},
			}), 1)
		}

		profiles.SelectedProfile = &profileID

		if err := workDir.WriteProfiles(profiles); err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.profile-select.error.profiles-write-failed",
					Other: "Cannot save your profile selection: {{ .Error }}",
				},
			}), 1)
		}

		return nil
	},
})

func init() {
	profileCommand.Subcommands = append(profileCommand.Subcommands, profileSelectCommand)
}
