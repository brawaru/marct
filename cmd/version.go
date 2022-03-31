package cmd

import (
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var versionCommand = createCommand(&cli.Command{
	Name: "version",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.version.usage",
		Other: "Manage game versions",
	}),
	Aliases: []string{"ver", "versions"},
})

func init() {
	app.Commands = append(app.Commands, versionCommand)
}
