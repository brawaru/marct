package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/brawaru/marct/globstate"
	"github.com/brawaru/marct/launcher"
	locales "github.com/brawaru/marct/locales"
	"github.com/imdario/mergo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name: "marct",
	Usage: locales.Translate(&i18n.Message{
		ID:    "app.usage",
		Other: "Minecraft architect tool",
	}),
	Description: locales.Translate(&i18n.Message{
		ID: "app.description",
		Other: `Minecraft architect tool. Manage your game with ease.

It allows you to manage your game versions, install mod loaders, mods and mod packs.

Generally Marct tries to stay compatible with Minecraft Launcher, but no warranties given.`,
	}),
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name: "workDir",
			Usage: locales.Translate(&i18n.Message{
				ID:    "app.command.args.workDir",
				Other: "Working directory",
			}),
			Value: ".",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage: locales.Translate(&i18n.Message{
				ID:    "app.command.args.verbose",
				Other: "Use verbose logging",
			}),
			Value:       false,
			Destination: &globstate.VerboseLogs,
		},
	},
	EnableBashCompletion:   true,
	UseShortOptionHandling: true,
	OnUsageError:           illegalInput,
	Before: func(ctx *cli.Context) error {
		workDirPath := ctx.String("workDir")
		var workDir *launcher.Instance

		if wd, workDirErr := launcher.OpenInstance(workDirPath); workDirErr == nil {
			workDir = wd
		} else {
			return cli.Exit(locales.TranslateWith(&i18n.Message{
				ID:    "app.error.workdir-init-err",
				Other: "Cannot initialise working directory: {{ .Error }}",
			}, map[string]string{
				"Error": workDirErr.Error(),
			}), 1)
		}

		ctx.Context = context.WithValue(ctx.Context, workDirKey, workDirPath)
		ctx.Context = context.WithValue(ctx.Context, instanceKey, workDir)

		return nil
	},
	After: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value(instanceKey).(*launcher.Instance)

		if err := workDir.Close(); err != nil {
			return cli.Exit(locales.TranslateWith(&i18n.Message{
				ID:    "app.error.workdir-close-err",
				Other: "Cannot close working directory: {{ .Error }}",
			}, map[string]string{
				"Error": err.Error(),
			}), 1)
		}

		return nil
	},
}

func init() {
	helpFmt := `##name##:
   {{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

##usage##:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[##args.global_options##]{{end}}{{if .Commands}} ##args.command## [##args.command_options##]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[##args.rest##]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

##version##:
   {{.Version}}{{end}}{{end}}{{if .Description}}

##description##:
   {{.Description | nindent 3 | trim}}{{end}}{{if .VisibleCommands}}

##commands##:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{range .VisibleCommands}}
   {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

##global_options##:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

##copyright##:
   {{.Copyright}}{{end}}
`

	helpCmd := `##name##:
   {{.HelpName}} - {{.Usage}}

##usage##:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [##args.command_options##]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[##args.rest##]{{end}}{{end}}{{if .Category}}

##category##:
   {{.Category}}{{end}}{{if .Description}}

##description##:
   {{.Description | nindent 3 | trim}}{{end}}{{if .VisibleFlags}}

##options##:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`
	helpSubCmd := `##name##:
   {{.HelpName}} - {{.Usage}}

##usage##:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} ##args.command##{{if .VisibleFlags}} [##args.command_options##]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[##args.rest##]{{end}}{{end}}{{if .Description}}

##description##:
   {{.Description | nindent 3 | trim}}{{end}}

##commands##:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{range .VisibleCommands}}
   {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

##options##:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	replacements := regexp.MustCompile("##[a-z_.]+##")

	replacementsMap := map[string]string{
		"name": locales.Translate(&i18n.Message{
			ID:    "cli.help.name",
			Other: "NAME",
		}),
		"usage": locales.Translate(&i18n.Message{
			ID:    "cli.help.usage",
			Other: "USAGE",
		}),
		"description": locales.Translate(&i18n.Message{
			ID:    "cli.help.description",
			Other: "DESCRIPTION",
		}),
		"commands": locales.Translate(&i18n.Message{
			ID:    "cli.help.commands",
			Other: "COMMANDS",
		}),
		"global_options": locales.Translate(&i18n.Message{
			ID:    "cli.help.global_options",
			Other: "GLOBAL OPTIONS",
		}),
		"copyright": locales.Translate(&i18n.Message{
			ID:    "cli.help.copyright",
			Other: "COPYRIGHT",
		}),
		"version": locales.Translate(&i18n.Message{
			ID:    "cli.help.version",
			Other: "VERSION",
		}),
		"args.global_options": locales.Translate(&i18n.Message{
			ID:    "cli.help.args.global_options",
			Other: "global options",
		}),
		"args.command": locales.Translate(&i18n.Message{
			ID:    "cli.help.args.command",
			Other: "command",
		}),
		"args.command_options": locales.Translate(&i18n.Message{
			ID:    "cli.help.args.command_options",
			Other: "command options",
		}),
		"args.rest": locales.Translate(&i18n.Message{
			ID:    "cli.help.args.rest",
			Other: "arguments...",
		}),
		"category": locales.Translate(&i18n.Message{
			ID:    "cli.help.category",
			Other: "CATEGORY",
		}),
		"options": locales.Translate(&i18n.Message{
			ID:    "cli.help.options",
			Other: "OPTIONS",
		}),
	}

	replacer := func(match string) string {
		runes := []rune(match)
		key := string(runes[2 : len(runes)-2])
		res := replacementsMap[key]

		if len(res) == 0 {
			return key
		}

		return res
	}

	cli.AppHelpTemplate = replacements.ReplaceAllStringFunc(helpFmt, replacer)
	cli.CommandHelpTemplate = replacements.ReplaceAllStringFunc(helpCmd, replacer)
	cli.SubcommandHelpTemplate = replacements.ReplaceAllStringFunc(helpSubCmd, replacer)

	cli.HelpFlag = &cli.BoolFlag{
		Name: "help",
		Usage: locales.Translate(&i18n.Message{
			ID:    "cli.flag.help",
			Other: "Show help",
		}),
		Aliases: []string{"h"},
	}

	app.HideHelpCommand = true
}

func Run(argv []string) error {
	return app.Run(argv)
}

type prefixReplacements map[string]func(trimmed string) string

func prefixReplacer(value string, prefixes prefixReplacements) string {
	for prefix, replacer := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return replacer(strings.TrimPrefix(value, prefix))
		}
	}

	return value
}

func illegalInput(_ *cli.Context, err error, isSubcommand bool) error {
	errStr := prefixReplacer(err.Error(), prefixReplacements{
		"flag provided but not defined: -": func(unknownFlag string) string {
			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Flag": unknownFlag,
				},
				DefaultMessage: &i18n.Message{
					ID:    "usage-error.unknown-flag",
					Other: "Option `{{ .Flag }}` is invalid",
				},
			})
		},
		"flag needs an argument: -": func(flagMissingValue string) string {
			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Flag": flagMissingValue,
				},
				DefaultMessage: &i18n.Message{
					ID:    "usage-error.flag-missing-value",
					Other: "Option `{{ .Flag }}` is missing a value",
				},
			})
		},
	})

	return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
		TemplateData: map[string]string{
			"Error": errStr,
		},
		DefaultMessage: &i18n.Message{
			ID:    "usage-error.other",
			Other: "Invalid usage: {{ .Error }}",
		},
	}), 1)
}

func createCommand(cmd *cli.Command) *cli.Command {
	err := mergo.Merge(cmd, cli.Command{
		HideHelpCommand:        true,
		UseShortOptionHandling: true,
		OnUsageError:           illegalInput,
	})

	if err != nil {
		fmt.Printf("[cmd] warn: failed to create command: %v\n", err)
	}

	return cmd
}
