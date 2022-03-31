package network

import (
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"math"
	"time"
)

var currentlyChecking = false

type Status int

const (
	StatusUnknown Status = iota // Network is in unknown status
	StatusCaptive               // Network is locked behind captive portal
	StatusOk                    // Network is uncompromised
)

type CheckResult struct {
	Status   Status
	Redirect *string
}

type Checker interface {
	Run() CheckResult
}

func nextAttemptLocalize(duration time.Duration) string {
	seconds := math.Round(float64(duration / time.Second))

	return locales.TranslateUsing(&i18n.LocalizeConfig{
		PluralCount: int(seconds),
		DefaultMessage: &i18n.Message{
			ID:    "relative.future.seconds",
			One:   "in {{ .PluralCount }} second",
			Other: "in {{ .PluralCount }} seconds",
		},
	})
}

func checkLoop(checker Checker) {
	debounce := time.Second

checkingLoop:
	for {
		res := checker.Run()

		switch res.Status {
		case StatusOk:
			break checkingLoop
		case StatusCaptive:
			debounce = time.Second * 32

			redirect := "[]"
			if res.Redirect != nil {
				redirect = *res.Redirect
			}

			println(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Redirect":    redirect,
					"NextAttempt": nextAttemptLocalize(debounce),
				},
				DefaultMessage: &i18n.Message{
					ID:    "network-checker.captive-portal-detected",
					Other: "Your network connection is limited!\n  Please visit {{ .Redirect }} to restore network connection.\n  We'll try connecting again {{ .NextAttempt }}",
				},
			}))
		case StatusUnknown:
			println(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"NextAttempt": nextAttemptLocalize(debounce),
				},
				DefaultMessage: &i18n.Message{
					ID:    "network-checker.unknown-network-error",
					Other: "There appears a problem with your network connection! We'll try to connect again {{ .NextAttempt }}",
				},
			}))
		}

		time.Sleep(debounce)

		debounce *= 2

		if debounce > (time.Second * 32) {
			debounce = time.Second * 32
		}
	}

	if debounce != time.Second {
		println(locales.Translate(&i18n.Message{
			ID:    "network-checker.reconnected",
			Other: "Welcome back to the Internet!",
		}))
	}
}

func WaitForConnection() {
	if currentlyChecking {
		for !currentlyChecking {
			time.Sleep(time.Second * 1) // fire every second otherwise busy work
		}

		return
	}

	currentlyChecking = true

	var checker MozChecker

	checkLoop(checker)

	currentlyChecking = false
}
