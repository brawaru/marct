package network

import (
	"errors"
	"math"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/brawaru/marct/network/concheck"
	"github.com/brawaru/marct/network/concheck/mozchecker"
)

type RequestRetrierOptions struct {
	MaxRetries            int              // Maximum number of retries, 0 means infinite.
	AllowNonNetworkErrors bool             // Whether to allow non-network errors to be retried.
	RetryDelay            time.Duration    // Initial delay between retries. It cannot be less than a zero.
	RetryDelayMultiplier  float64          // Multiplier of retry delay for each retry. It cannot be less than one.
	RetryDelayMax         time.Duration    // Maximum delay between retries. It cannot be less than RetryDelay.
	ConnectionChecker     concheck.Checker // Network connection checker in case of network error. If nil, then connection is not checked.
}

type RetrierOption func(*RequestRetrierOptions)

func WithMaxRetries(maxRetries int) RetrierOption {
	return func(options *RequestRetrierOptions) {
		options.MaxRetries = maxRetries
	}
}

func WithRetryDelay(delay time.Duration, multiplier float64, max time.Duration) RetrierOption {
	return func(options *RequestRetrierOptions) {
		options.RetryDelay = delay
		options.RetryDelayMultiplier = math.Max(1, multiplier)
		if max < delay {
			options.RetryDelayMax = delay
		} else {
			options.RetryDelayMax = max
		}
	}
}

func WithConstRetryDelay(delay time.Duration) RetrierOption {
	return WithRetryDelay(delay, 1, delay)
}

func WithConnectionChecker(checker concheck.Checker) RetrierOption {
	return func(options *RequestRetrierOptions) {
		options.ConnectionChecker = checker
	}
}

var isNetworkReset = func(errno syscall.Errno) bool {
	return errno == syscall.ECONNRESET
}

func IsNetworkError(e error) bool {
	{
		var dnsErr *net.DNSError
		if errors.As(e, &dnsErr) {
			return dnsErr.IsTemporary || dnsErr.IsNotFound || dnsErr.IsTimeout
		}
	}

	{
		var syscallErr syscall.Errno
		if errors.As(e, &syscallErr) {
			return isNetworkReset != nil && isNetworkReset(syscallErr)
		}
	}

	return false
}

func WithRetries(options ...RetrierOption) Option {
	o := &RequestRetrierOptions{
		MaxRetries:            0,
		AllowNonNetworkErrors: false,
		RetryDelay:            time.Second,
		RetryDelayMultiplier:  2,
		RetryDelayMax:         time.Minute,
		ConnectionChecker:     mozchecker.NewMozChecker(),
	}

	for _, option := range options {
		option(o)
	}

	return func(req *http.Request, options *ActionOptions) {
		retries := 0 // how many retries we've done

		options.ErrorHandlers = append(options.ErrorHandlers, func(err error) error {
			isNetErr := IsNetworkError(err)

			if (o.AllowNonNetworkErrors || isNetErr) && (o.MaxRetries == 0 || retries < o.MaxRetries) {
				delay := o.RetryDelay * time.Duration(math.Max(1, o.RetryDelayMultiplier*float64(retries)))

				if delay > o.RetryDelayMax {
					delay = o.RetryDelayMax
				}

				if isNetErr && o.ConnectionChecker != nil {
					netCheckStart := time.Now()

					if err := concheck.WaitForConnection(req.Context(), o.ConnectionChecker); err != nil {
						return err
					}

					if time.Since(netCheckStart) > delay {
						// There is no need to delay if connection checking took longer than the delay itself.
						delay = 0
					}
				}

				if delay > 0 {
					time.Sleep(delay)
				}

				retries += 1

				return ErrRetryRequest
			}

			return nil // Pass error to the next handler
		})
	}
}
