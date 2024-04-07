package cmd

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	offlineAccount "github.com/brawaru/marct/offline/account"
	offlineAuthFlow "github.com/brawaru/marct/offline/authflow"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/utils/pointers"
	xboxAccount "github.com/brawaru/marct/xbox/account"
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
		instance := ctx.Context.Value(instanceKey).(*launcher.Instance)

		profiles, err := instance.ReadProfiles()
		if err != nil {
			if utils.DoesNotExist(err) {
				err = errors.New(locales.Translate(&i18n.Message{
					ID:    "command.launch.error.profiles-file-does-not-exist",
					Other: "profiles file does not exist",
				}))
			}

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

		// TODO: add -i option that prompts user to select profile to launch

		var profile launcher.Profile
		if ctx.NArg() == 0 {
			if profiles.SelectedProfile == nil || ctx.Bool("i") {
				p, err := SelectProfileFlow(profiles, WithMessage(locales.Translate(&i18n.Message{
					ID:    "command.launch.prompt.select-profile",
					Other: "Select profile to launch",
				})))

				if err != nil {
					return err
				}

				profile = *p
			} else {
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

		if len(accountsStore.Accounts) == 0 {
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
			if newSelection, selectionErr := SelectAccountFlow(accountsStore.Accounts); selectionErr != nil {
				return selectionErr
			} else {
				selectedAccount = newSelection
			}
		}

		switch selectedAccount.Type {
		case xboxAccount.AccountType:
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
						Other: "Cannot authorize your Xbox account: {{ .Error }}",
					},
				}), 1)
			}
		case offlineAccount.AccountType:
			authFlow := offlineAuthFlow.CreateAuthFlow(&offlineAuthFlow.Options{
				UsernameRequestHandler: offlineUsernamePrompt,
			})

			err := authFlow.RefreshAccount(selectedAccount)

			if err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.launch.error.offline-account-refresh-failed",
						Other: "Cannot authorize your offline account: {{ .Error }}",
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

		versionID := profile.LastVersionID
		if versionID == "latest-release" || versionID == "latest-snapshot" {
			// this is a special ID and we'll have to fetch versions list
			v, err := instance.FetchVersions(false)
			if err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.launch.error.versions-fetch-failed",
						Other: "Cannot acquire a list of latest versions: {{ .Error }}",
					},
				}), 1)
			}

			vd := v.GetVersion(versionID)
			if vd == nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"VersionID": versionID,
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.launch.error.version-not-found",
						Other: "Cannot get recent version for ID {{ .VersionID }}.",
					},
				}), 1)
			}

			if err := instance.DownloadVersionFile(*vd); err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.launch.error.version-fetch-failed",
						Other: "Cannot get version manifest for {{ .VersionID }}: {{ .Error }}",
					},
				}), 1)
			}

			versionID = vd.ID
		}

		version, err := instance.ReadVersionWithInherits(versionID)
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

		if err := instance.DownloadVersion(*version); err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.launch.error.cannot-download-version",
					Other: "Unable to download and verify version files: {{ .Error }}",
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
				var e *exec.ExitError

				if errors.As(err, &e) && e.ExitCode() != 0 {
					println(locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"ExitCode": strconv.Itoa(e.ExitCode()),
						},
						DefaultMessage: &i18n.Message{
							ID:    "command.launch.warn.non-zero-exit",
							Other: "Game process exited with code {{ .ExitCode }}",
						},
					}))
				}

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
