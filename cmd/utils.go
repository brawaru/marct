package cmd

import (
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var utilsCommand = createCommand(&cli.Command{
	Name: "utils",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.utils.usage",
		Other: "Utils for advanced users",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.utils.description",
		Other: "This command contains various tools that are meant for more advanced users.",
	}),
})

func init() {
	app.Commands = append(app.Commands, utilsCommand)
}
