package cmd

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/validfile"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

func mapVersions(entries []launcher.VersionDescriptor) []string {
	var res []string
	for _, entry := range entries {
		res = append(res, entry.ID)
	}
	return res
}

var versionInstallCommand = createCommand(&cli.Command{
	Name:    "install",
	Aliases: []string{"i", "download", "dl"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.version-download.usage",
		Other: "Downloads or verifies version of the game",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.version-download.description",
		Other: "Downloads or verifies previously downloaded version of the game",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.version-download.args-usage",
		Other: "<version ID>",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value("workDir").(*launcher.Instance)
		versions, err := workDir.FetchVersions(false)

		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.version-download.err.manifest-fetch-failed",
					Other: "failed to fetch versions manifest due to error: {{ .Error }}",
				},
			}), 1)
		}

		if ctx.NArg() > 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.version-download.error.too-many-arguments",
				Other: "Too many arguments provided, only version ID is expected",
			}), 1)
		}

		versionId := ctx.Args().First()

		if versionId == "" {
			err := survey.AskOne(&survey.Select{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.version-download.survey.version",
					Other: "Version to download",
				}),
				Options: mapVersions(versions.Versions),
			}, &versionId)

			if err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.version-download.survey.error.cannot-read-answer",
						Other: "Cannot read your answer: {{ .Error }}",
					},
				}), 1)
			}
		}

		versionDescriptor := versions.GetVersion(versionId)

		if versionDescriptor == nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"VersionId": versionId,
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.version-download.err.version-not-found",
					Other: "cannot find version \"{{ .VersionId }}\"",
				},
			}), 5)
		}

		manifestDlErr := workDir.DownloadVersionFile(*versionDescriptor)

		if manifestDlErr != nil {
			{
				var validationErr *validfile.ValidateError
				if errors.As(manifestDlErr, &validationErr) {
					return cli.Exit(locales.Translate(&i18n.Message{
						ID:    "command.version-download.err.manifest-validate-error",
						Other: "downloaded version file is malformed",
					}), 1)
				}
			}

			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": manifestDlErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.version-download.err.unknown-error-fetch-version",
					Other: "unknown error when downloading the version: {{ .Error }}",
				},
			}), 1)
		}

		version, err := workDir.ReadVersionFile(versionDescriptor.ID)
		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.version-download.err.version-file-read-failed",
					Other: "Cannot read version file: {{ .Error }}",
				},
			}), 1)
		}

		versionDlErr := workDir.DownloadVersion(*version)

		if versionDlErr != nil {
			{
				var validationErr *validfile.ValidateError
				if errors.As(versionDlErr, &validationErr) {
					return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"File":  validationErr.Name,
							"Error": validationErr.Err.Error(),
						},
						DefaultMessage: &i18n.Message{
							ID:    "command.version-download.err.file-verification-failed",
							Other: "downloaded file {{ .File }} failed validation: {{ .Error }}",
						},
					}), 1)
				}
			}

			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": versionDlErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.version-download.err.unknown-error-download",
					Other: "unknown error when downloading client.json: {{ .Error }}",
				},
			}), 1)
		}

		return nil
	},
})

func init() {
	versionCommand.Subcommands = append(versionCommand.Subcommands, versionInstallCommand)
}
