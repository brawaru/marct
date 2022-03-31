package cmd

import (
	"errors"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
	"os"
)

var utilsVirtualizeCommand = createCommand(&cli.Command{
	Name: "virtualize",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.utils-virtualize.usage",
		Other: "Virtualizes a version index file",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.utils-virtualize.description",
		Other: "Virtualizes a version index file, mapping all assets to their appropriate locations.",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.utils-virtualize.args-usage",
		Other: "<index ID>",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value("workDir").(*launcher.Instance)

		if ctx.NArg() != 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.utils-virtualize.error.illegal-num-of-args",
				Other: "Illegal number of arguments: expected only index ID",
			}), 1)
		}
		i := ctx.Args().First()
		ai, err := workDir.ReadAssetIndex(i)
		if err != nil {
			return err // FIXME: wrap error
		}

		if virtualizeErr := ai.Virtualize(workDir.DefaultAssetsObjectResolver(), workDir.AssetsVirtualPath(i)); virtualizeErr != nil {
			var pathErr *os.PathError
			if errors.As(virtualizeErr, &pathErr) && errors.Is(pathErr, os.ErrNotExist) {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"File": pathErr.Path,
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.utils-virtualize.error.not-exists",
						Other: "File {{ .File }} does not exist",
					},
				}), 1)
			}

			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": virtualizeErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.utils-virtualize.error.unknown",
					Other: "Unknown error: {{ .Error }}",
				},
			}), 1)
		}

		return nil
	},
})

func init() {
	utilsCommand.Subcommands = append(utilsCommand.Subcommands, utilsVirtualizeCommand)
}
