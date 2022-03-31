package launcher

import (
	"github.com/brawaru/marct/globstate"
	"github.com/brawaru/marct/locales"
	"github.com/matishsiao/goInfo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"regexp"
	"runtime"
	"strconv"
)

type OS struct {
	Name    *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
	Arch    *string `json:"arch,omitempty"`
}

var sysInfo goInfo.GoInfoObject

func init() {
	info, err := goInfo.GetInfo()
	if err == nil {
		sysInfo = info
	} else {
		println(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"Error": err.Error(),
			},
			DefaultMessage: &i18n.Message{
				ID:    "warning.sys-info-query-failed",
				Other: "failed to read system information (this might break installation process!). reason: {{ .Error }}",
			},
		}))
	}
}

func currentOS() string {
	v := runtime.GOOS
	switch v {
	case "darwin":
		return "osx"
	default:
		return v
	}
}

func currentArch() string {
	// this is probably incorrect
	if strconv.IntSize == 32 {
		return "x86"
	} else {
		return "x64"
	}
}

func (o *OS) Matches() bool {
	if o.Name != nil {
		if t, err := regexp.Compile(*o.Name); err == nil { // <- is it regular expression!?
			v := currentOS()

			if !t.MatchString(v) {
				if globstate.VerboseLogs {
					println(locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"RegularExpression": t.String(),
							"Value":             v,
						},
						DefaultMessage: &i18n.Message{
							ID:    "log.minecraft.versions.match-failed.os",
							Other: "OS does not match: excepted to match `{{ .RegularExpression }}`, but `{{ .Value }}` doesn't",
						},
					}))
				}

				return false
			}
		} else {
			if globstate.VerboseLogs {
				println(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"RegularExpression": t.String(),
						"Error":             err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "log.minecraft.versions.match-failed.os-regex-fail",
						Other: "OS does not match: cannot build regular expression `{{ .RegularExpression }}` due to `{{ .Error }}`",
					},
				}))
			}

			return false
		}
	}

	if o.Version != nil {
		if t, err := regexp.Compile(*o.Version); err == nil {
			v := sysInfo.Core

			if !t.MatchString(v) {
				if globstate.VerboseLogs {
					println(locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"RegularExpression": t.String(),
							"Value":             v,
						},
						DefaultMessage: &i18n.Message{
							ID:    "log.minecraft.versions.match-failed.version",
							Other: "OS version does not match: excepted to match `{{ .RegularExpression }}`, but `{{ .Value }}` doesn't",
						},
					}))
				}
			}
		} else {
			if globstate.VerboseLogs {
				println(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"RegularExpression": t.String(),
						"Error":             err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "log.minecraft.versions.match-failed.version-regex-fail",
						Other: "OS version does not match: cannot build regular expression `{{ .RegularExpression }}` due to `{{ .Error }}`",
					},
				}))
			}

			return false
		}
	}

	if o.Arch != nil {
		v := *o.Arch

		if v != currentArch() {
			if globstate.VerboseLogs {
				println(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Expected": v,
						"Value":    v,
					},
					DefaultMessage: &i18n.Message{
						ID:    "log.minecraft.versions.match-failed.version",
						Other: "OS arch does not match: excepted `{{ .Expected }}`, got `{{ .Value }}`",
					},
				}))
			}

			return false
		}
	}

	return true
}
