package concheck

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// FIXME: locales and prints should not be in here, instead prefer handler approach

var currentlyChecking = false

type Status int

const (
	StatusUnknown Status = iota // Network is in unknown status
	StatusCaptive               // Network is locked behind captive portal
	StatusOK                    // Network is uncompromised
)

type CheckResult struct {
	Status   Status
	Redirect *string
}

type Checker interface {
	// Checks whether the connection is established and not behind the captive portal.
	Run() (CheckResult, error)
}

type checkInProgress struct {
	m sync.Mutex
	// All the waiters until the connection is established.
	waiters []chan error
}

func (c *checkInProgress) Wait() chan error {
	c.m.Lock()
	defer c.m.Unlock()
	ch := make(chan error)
	c.waiters = append(c.waiters, ch)
	return ch
}

func (c *checkInProgress) Result(err error) {
	c.m.Lock()
	defer c.m.Unlock()

	for _, ch := range c.waiters {
		ch <- err
	}
}

var currentChecks = make(map[Checker]*checkInProgress)

// Waits until the connection is available. The checking is performed in a separate goroutine, however, the calling
// routine will be blocked until the context is cancelled/expired or the connection is available.
func WaitForConnection(ctx context.Context, checker Checker) error {
	c := currentChecks[checker]

	if c == nil {
		n := checkInProgress{
			waiters: []chan error{},
			m:       sync.Mutex{},
		}

		currentChecks[checker] = &n
		c = &n

		go func() {
			c.Result(checkLoop(ctx, checker))
		}()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-c.Wait():
		return e
	}
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

func checkLoop(ctx context.Context, checker Checker) error {
	debounce := time.Second

checkingLoop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			res, err := checker.Run()

			if err != nil {
				return err
			}

			switch res.Status {
			case StatusOK:
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
	}

	if debounce != time.Second {
		println(locales.Translate(&i18n.Message{
			ID:    "network-checker.reconnected",
			Other: "Welcome back to the Internet!",
		}))
	}

	return nil
}
