package cmd

import (
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var initCommand = createCommand(&cli.Command{
	Name: "init",
	Usage: locales.Translate(&i18n.Message{
		ID:    "marct.command.init.usage",
		Other: "Initialises empty Minecraft directory",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "marct.command.init.description",
		Other: "Initialises empty Minecraft directory", // TODO: probably add better description?
	}),
	Action: func(ctx *cli.Context) error {
		panic("not implemented")
	},
})

func init() {
	app.Commands = append(app.Commands, initCommand)
}
