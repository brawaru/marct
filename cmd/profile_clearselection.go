package cmd

import (
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var profileSelectionClearCommand = createCommand(&cli.Command{
	Name:    "selection-clear",
	Aliases: []string{"deselect"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.profile-selection-clear.usage",
		Other: "Clear default profile selection",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.profile-selection-clear.description",
		Other: "Allows to clear default profile used in management or launch commands",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value("workDir").(*launcher.Instance)
		profiles := ctx.Context.Value("profiles").(*launcher.Profiles)

		profiles.SelectedProfile = nil
		err := workDir.WriteProfiles(profiles)
		if err != nil {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.profile-selection-clear.profiles-write-failed",
				Other: "Failed to save changes to profile file: {{ .Error }}",
			}), 1)
		}

		return cli.Exit(locales.Translate(&i18n.Message{
			ID:    "command.profile-selection-clear.success",
			Other: "Cleared profile selection",
		}), 0)
	},
})

func init() {
	profileCommand.Subcommands = append(profileCommand.Subcommands, profileSelectionClearCommand)
}
