package cmd

import (
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var javaCommand = createCommand(&cli.Command{
	Name: "java",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.java.usage",
		Other: "Manage Java versions",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.java.description",
		Other: "This command allows you to manage Mojang JRE installations",
	}),
})

func init() {
	app.Commands = append(app.Commands, javaCommand)
}
