package locales

import (
	"embed"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/jeandeaual/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"os"
)

var GlobalLocalizer *i18n.Localizer

//go:embed *.toml
var locales embed.FS

func init() {
	var userLocales []string

	if l, found := os.LookupEnv("LANGUAGE"); found && l != "" {
		fmt.Printf("[localizer] info: found locale override: %q\n", l)

		if tag, err := language.Parse(l); err == nil {
			userLocales = append(userLocales, tag.String())
		}
	}

	if len(userLocales) == 0 {
		if l, err := locale.GetLocales(); err == nil {
			userLocales = l
		} else {
			fmt.Println("[localizer] warn: failed to acquire user locales, using en-US")
			userLocales = []string{"en-US"}
		}
	}

	var bundle = i18n.NewBundle(language.AmericanEnglish)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	dirEntries, err := locales.ReadDir(".")
	if err == nil {
		for _, entry := range dirEntries {
			if entry.IsDir() {
				continue
			}

			fileName := entry.Name()

			contents, err := locales.ReadFile(fileName)

			if err == nil {
				bundle.MustParseMessageFileBytes(contents, fileName)
			} else {
				fmt.Printf("[localizer] warn: failed to read file %s: %v\n", fileName, err)
			}
		}
	} else {
		fmt.Printf("[localizer] warn: failed to read locales, using default messages. %v\n", err)
	}

	GlobalLocalizer = i18n.NewLocalizer(bundle, userLocales...)
}

func Translate(msg *i18n.Message) string {
	return TranslateUsing(&i18n.LocalizeConfig{
		DefaultMessage: msg,
	})
}

func TranslateWith(msg *i18n.Message, replacements map[string]string) string {
	return TranslateUsing(&i18n.LocalizeConfig{
		TemplateData:   replacements,
		DefaultMessage: msg,
	})
}

func Pluralise(msg *i18n.Message, count int) string {
	return TranslateUsing(&i18n.LocalizeConfig{
		PluralCount:    count,
		DefaultMessage: msg,
	})
}

func TranslateUsing(config *i18n.LocalizeConfig) string {
	var str, err = GlobalLocalizer.Localize(config)

	if len(str) == 0 {
		panic(err)
	} else {
		return str
	}
}
