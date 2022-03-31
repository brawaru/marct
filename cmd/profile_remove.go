package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var profileRemoveCommand = createCommand(&cli.Command{
	Name:    "remove",
	Aliases: []string{"rm"},
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.profile-remove.args-usage",
		Other: "[profile identifier]",
	}),
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.profile-remove.usage",
		Other: "Remove profile",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.profile-remove.description",
		Other: "Remove existing profile",
	}),
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-remove.args.force",
				Other: "Remove profile without confirmation",
			}),
		},
	},
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value("workDir").(*launcher.Instance)

		if ctx.NArg() < 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.profile-remove.error.no-profile-identifier",
				Other: "You must specify profile identifier",
			}), 1)
		} else if ctx.NArg() > 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.profile-remove.error.too-many-arguments",
				Other: "Too many arguments, only profile identifier is expected",
			}), 1)
		}

		profileID := ctx.Args().First()

		profiles := ctx.Context.Value("profiles").(*launcher.Profiles)

		profile, ok := profiles.Profiles[profileID]

		if !ok {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"ProfileID": profileID,
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.profile-remove.error.profile-not-found",
					Other: "Profile with identifier \"{{ .ProfileID }}\" not found",
				},
			}), 1)
		}

		profileName := profile.Name

		if profileName == "" {
			profileName = locales.Translate(&i18n.Message{
				ID:    "command.profile-remove.unnamed-profile",
				Other: "Unnamed profile",
			})
		}

		if !ctx.Bool("force") {
			var answer bool
			err := survey.AskOne(&survey.Confirm{
				Message: locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"ProfileName": profileName,
						"ProfileID":   profileID,
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.profile-remove.confirm-remove",
						Other: "Are you sure you want to remove {{ .ProfileName }} ({{ .ProfileID }})? Associated files will not be deleted.",
					},
				}),
			}, &answer)

			if err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.profile-remove.error.ask-confirmation",
						Other: "Cannot ask for confirmation: {{ .Error }}",
					},
				}), 1)
			}

			if !answer {
				return nil
			}
		}

		delete(profiles.Profiles, profileID)

		if err := workDir.WriteProfiles(profiles); err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.profile-remove.error.write-profiles",
					Other: "Cannot save profiles: {{ .Error }}",
				},
			}), 1)
		}

		return nil
	},
})

func init() {
	profileCommand.Subcommands = append(profileCommand.Subcommands, profileRemoveCommand)
}
