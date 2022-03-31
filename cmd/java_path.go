package cmd

import (
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/validfile"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

var javaPathCommand = createCommand(&cli.Command{
	Name: "path",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.java-path.usage",
		Other: "Print Java installation location",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.java-path.description",
		Other: "Prints location where Java was installed",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.java-path.args-usage",
		Other: "<type>",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value("workDir").(*launcher.Instance)

		if ctx.NArg() != 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.java-path.error.illegal-num-of-args",
				Other: "Illegal number of arguments: expected only type of Java",
			}), 1)
		}

		t := ctx.Args().First()

		platforms, readErr := workDir.ReadJREs()
		if readErr != nil {
			if utils.DoesNotExist(readErr) {
				return cli.Exit(locales.Translate(&i18n.Message{
					ID:    "command.java-path.error.manifest-not-exists",
					Other: "Cannot find cached manifest file, please use `marct java refresh` to create it.",
				}), 1)
			}

			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": readErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.java-path.error.unknown",
					Other: "Unknown error when reading cached manifest file: {{ .Error }}",
				},
			}), 1)
		}

		classifiers, selector := platforms.GetMatching()

		_, exists := classifiers[t]

		if !exists {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Type": t,
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.java-path.error.version-does-not-exist",
					Other: "Version {{ .Type }} does not exist",
				},
			}), 1)
		}

		path := filepath.Join(workDir.JREPath(t, selector), t)

		if _, existErr := validfile.DirExists(path); existErr != nil {
			return cli.Exit(&i18n.Message{
				ID:    "command.java-path.error.not-downloaded",
				Other: "That type of JRE is not downloaded",
			}, 1)
		}

		// FIXME: check that exists

		_, _ = os.Stdout.WriteString(path)

		return nil
	},
})

func init() {
	javaCommand.Subcommands = append(javaCommand.Subcommands, javaPathCommand)
}
