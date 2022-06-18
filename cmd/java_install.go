package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/launcher/java"
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var javaInstallCommand = createCommand(&cli.Command{
	Name: "install",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.java-install.usage",
		Other: "Install Java version",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.java-install.description",
		Other: "Installs a Mojang JRE either interactively or by passed flag",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.java-install.args",
		Other: "<type>",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value(instanceKey).(*launcher.Instance)

		if ctx.IsSet("version") && ctx.IsSet("latest") {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.java-install.error.incompatible-flags",
				Other: "Both version and latest flags are provided, only one is allowed",
			}), 1)
		}

		if ctx.Args().Len() > 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.java-install.error.too-many-args",
				Other: "Too many arguments: excepted only type of Java to install",
			}), 1)
		}

		t := ctx.Args().First()

		platforms, fetchErr := workDir.FetchJREs(false)
		if fetchErr != nil || platforms == nil {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.java-install.error.fetch-failed",
				Other: "Cannot retrieve a list of available JREs",
			}), 1)
		}

		classifiers, selector := platforms.GetMatching()

		if len(classifiers) == 0 {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Selector": selector,
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.java-install.error.no-jres-available",
					Other: "No JREs available for the platform {{ .Selector }}",
				},
			}), 1)
		}

		if len(t) == 0 {
			var options []string
			optionsMappings := map[string]string{}

			for classifier, v := range classifiers {
				mostRecent := v.MostRecent()

				if mostRecent == nil {
					continue
				}

				option := fmt.Sprintf("%s (%s)", classifier, mostRecent.Version.Name)
				optionsMappings[option] = classifier
				options = append(options, option)
			}

			sort.Strings(options)

			var resp string

			if surveyingErr := survey.AskOne(&survey.Select{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.java-install.survey.type",
					Other: "Select type of JRE to install",
				}),
				Options: options,
			}, &resp); surveyingErr != nil {
				return cli.Exit(locales.Translate(&i18n.Message{
					ID:    "command.java-install.error.survey-fail-type",
					Other: "Failed to read response for 'type'",
				}), 1)
			}

			t = optionsMappings[resp]
		}

		if installErr := workDir.InstallJRE(*platforms, t); installErr != nil {
			{
				var v *launcher.PostValidationError
				if errors.As(installErr, &v) {
					buf := new(strings.Builder)

					for i, object := range v.BadObjects {
						if i != 0 {
							buf.WriteByte('\n')
						}

						var objectType string
						var objectState string

						switch object.Type {
						case java.TypeDir:
							objectType = locales.Translate(&i18n.Message{
								ID:    "command.java-install.error.post-validation.object-type.dir",
								Other: "directory",
							})
						case java.TypeFile:
							objectType = locales.Translate(&i18n.Message{
								ID:    "command.java-install.error.post-validation.object-type.file",
								Other: "file",
							})
						case java.TypeLink:
							objectType = locales.Translate(&i18n.Message{
								ID:    "command.java-install.error.post-validation.object-type.link",
								Other: "link",
							})
						}

						switch object.State {
						case launcher.FileStateCorrupted:
							objectState = locales.Translate(&i18n.Message{
								ID:    "command.java-install.error.post-validation.object-state.corrupted",
								Other: "corrupted",
							})
						case launcher.FileStateNotDownloaded:
							objectState = locales.Translate(&i18n.Message{
								ID:    "command.java-install.error.post-validation.object-state.not-downloaded",
								Other: "does not exist",
							})
						default:
							objectState = locales.Translate(&i18n.Message{
								ID:    "command.java-install.error.post-validation.object-state.unknown",
								Other: "unknown",
							})
						}

						buf.WriteString(locales.TranslateUsing(&i18n.LocalizeConfig{
							TemplateData: map[string]string{
								"Type":  objectType,
								"Path":  object.Destination,
								"State": objectState,
							},
							DefaultMessage: &i18n.Message{
								ID:    "command.java-install.error.install-failed-dir",
								Other: "- {{ .Type }} {{ .Path }}: {{ .State }}",
							},
						}))
					}

					return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"List": buf.String(),
						},
						DefaultMessage: &i18n.Message{
							ID:    "command.java-install.error.post-validation-failed",
							Other: "Installation failed as the following objects cannot be validated after all steps are complete:\n{{ .List }}",
						},
					}), 1)
				}
			}

			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": installErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.java-install.error.installation-failed",
					Other: "Installation failed: {{ .Error }}",
				},
			}), 1)
		}

		return nil
	},
})

func init() {
	javaCommand.Subcommands = append(javaCommand.Subcommands, javaInstallCommand)
}
