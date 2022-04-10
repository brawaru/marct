package cmd

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	minecraftAccount "github.com/brawaru/marct/minecraft/account"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/utils/pointers"
	xboxAuthFlow "github.com/brawaru/marct/xbox/authflow"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var launchCommand = createCommand(&cli.Command{
	Name: "launch",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.launch.usage",
		Other: "Launch the game",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.launch.description",
		Other: "Launches the game either using selected profile or specified using the argument.",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.launch.args-usage",
		Other: "[profile id]",
	}),
	Flags: nil,
	Action: func(ctx *cli.Context) error {
		instance := ctx.Context.Value("workDir").(*launcher.Instance)

		profiles, err := instance.ReadProfiles()
		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.launch.error.profiles-read-failed",
					Other: "Failed to read launcher profiles file: {{ .Error }}",
				},
			}), 1)
		}

		var profile launcher.Profile
		if ctx.NArg() == 0 {
			if profiles.SelectedProfile != nil {
				i := *profiles.SelectedProfile
				p, ok := profiles.Profiles[i]
				if !ok {
					return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"Name": i,
						},
						DefaultMessage: &i18n.Message{
							ID: "command.launch.error.default-profile-not-found",
							Other: "Selected profile \"{{ .Name }}\" is missing." +
								" Select existing profile or specify profile to launch via argument.",
						},
					}), 1)
				}
				profile = p
			} else {
				return cli.Exit(locales.Translate(&i18n.Message{
					ID:    "command.launch.error.no-selected-profile",
					Other: "No profile selected to launch",
				}), 1)
			}
		} else if ctx.NArg() == 1 {
			i := ctx.Args().First()
			p, ok := profiles.Profiles[i]
			if !ok {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Name": i,
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.launch.error.invalid-profile-specified",
						Other: "Profile \"{{ .Name }}\" does not exist.",
					},
				}), 1)
			}
			profile = p
		} else {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.launch.error.invalid-args-number",
				Other: "Invalid number of arguments",
			}), 1)
		}

		accountsStore, err := instance.OpenAccountsStore()

		utils.DClose(accountsStore) // we won't be making any changes so closing the file immediately

		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.launch.error.accounts-store-open-failed",
					Other: "Failed to open read your accounts: {{ .Error }}",
				},
			}), 1)
		}

		if accountsStore.Accounts == nil || len(accountsStore.Accounts) == 0 {
			command := strings.Join([]string{app.Name, accountCommand.Name, accountAddCommand.Name}, " ")
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Command": command,
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.launch.error.no-accounts",
					Other: "You have no accounts. Please add one using \"{{ .Command }}\" command.",
				},
			}), 1)
		}

		selectedAccount := accountsStore.GetSelectedAccount()

		if selectedAccount == nil {
			// let's ask user to select account using survey library
			var options []string
			optionsMappings := make(map[string]string)
			for accountID, account := range accountsStore.Accounts {
				option := ""
				properties, err := minecraftAccount.ReadProperties(&account)
				if err != nil {
					// set option to translated 'Unknown account ({{ .ID }})'
					option = locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"ID": accountID,
						},
						DefaultMessage: &i18n.Message{
							ID:    "command.launch.unknown-account",
							Other: "Unknown account ({{ .ID }})",
						},
					})
				} else {
					option = properties.Username
				}

				options = append(options, option)
				optionsMappings[option] = accountID
			}

			var selectedID string
			if err := survey.AskOne(&survey.Select{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.launch.select-account",
					Other: "Select account to launch:",
				}),
				Options: options,
			}, &selectedID, survey.WithValidator(func(ans interface{}) error {
				// if answer is not of type string, then return error about invalid type
				i, ok := ans.(survey.OptionAnswer)
				if !ok {
					return errors.New(locales.Translate(&i18n.Message{
						ID:    "command.launch.error.survey-validation-invalid-type",
						Other: "Invalid response type.",
					}))
				}

				// if optionMappings doesn't have a key i, then return error about invalid selection
				_, ok = optionsMappings[i.Value]
				if !ok {
					return errors.New(locales.Translate(&i18n.Message{
						ID:    "command.launch.error.invalid-selection",
						Other: "Invalid selection. Please select one of the options.",
					}))
				}

				return nil
			})); err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.launch.error.account-select-failed",
						Other: "Failed to read your selection: {{ .Error }}",
					},
				}), 1)
			}

			selection := accountsStore.Accounts[optionsMappings[selectedID]]
			selectedAccount = &selection
		}

		print("Logging you in...\r")

		switch selectedAccount.Type {
		case "xbox":
			k, err := keyringOpenFlow(instance)

			if err != nil {
				return err
			}

			authFlow := xboxAuthFlow.CreateAuthFlow(&xboxAuthFlow.Options{
				DeviceAuthHandler: xboxDeviceAuthPrompt,
				Keyring:           k,
			})

			err = authFlow.RefreshAccount(selectedAccount)

			if err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.launch.error.xbox-account-refresh-failed",
						Other: "Cannot authorize your account: {{ .Error }}",
					},
				}), 1)
			}
		default:
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "command.launch.error.account-type-not-supported",
					Other: "Account type {{ .AccountType }} is not supported.",
				},
			}), 1)
		}

		version, err := instance.ReadVersionWithInherits(profile.LastVersionID)
		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.launch.error.cannot-inherit-versions",
					Other: "Unable to prepare for launch: {{ .Error }}",
				},
			}), 1) // FIXME: translate error to message
		}

		if lr, err := instance.Launch(*version, launcher.LaunchOptions{
			Background:    ctx.Bool("background"),
			JavaPath:      pointers.DerefOrDefault(profile.JavaPath),
			Resolution:    profile.Resolution,
			Authorization: *selectedAccount.Authorization,
			GameDirectory: filepath.Join(instance.Path, filepath.FromSlash(profile.GameDir)), // MCL compat: no sanitization
			JavaArgs:      profile.JavaArgs,
		}); err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.launch.error.launch-failed",
					Other: "Cannot launch game: {{ .Error }}",
				},
			}), 1) // FIXME: translate error to message
		} else {
			if err := lr.Command.Wait(); err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.launch.error.wait-error",
						Other: "Cannot wait for child process: {{ .Error }}",
					},
				}), 1)
			}

			if err := lr.Clean(); err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.launch.error.clean-error",
						Other: "Cannot clean up temporary files: {{ .Error }}",
					},
				}), 1)
			}
		}

		return nil
	},
})

func init() {
	app.Commands = append(app.Commands, launchCommand)
}
