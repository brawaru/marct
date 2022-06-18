package cmd

import (
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var javaRefreshCommand = createCommand(&cli.Command{
	Name: "refresh",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.java-refresh.usage",
		Other: "Refresh Java Runtimes manifest",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.java-refresh.description",
		Other: "Fetches fresh version of the Java Runtimes manifest from Mojang",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value(instanceKey).(*launcher.Instance)

		if _, fetchErr := workDir.FetchJREs(true); fetchErr != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": fetchErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.java-refresh.error.fetch-error",
					Other: "Error while refreshing: {{ .Error }}",
				},
			}), 1)
		}

		return nil
	},
})

func init() {
	javaCommand.Subcommands = append(javaCommand.Subcommands, javaRefreshCommand)
}
