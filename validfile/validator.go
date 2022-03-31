package validfile

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/brawaru/marct/utils"
	"hash"
	"io"
	"os"
)

type HashMatchError struct {
	Expected []byte
	Actual   []byte
}

func (h *HashMatchError) Error() string {
	return fmt.Sprintf("hash does not match: expected %x, got %x", h.Expected, h.Actual)
}

func (h *HashMatchError) Is(target error) bool {
	t, ok := target.(*HashMatchError)
	return ok &&
		(t.Expected == nil || bytes.Equal(h.Expected, t.Expected)) &&
		(t.Actual == nil || bytes.Equal(h.Actual, t.Actual))
}

type ValidateError struct {
	Name string
	Err  error
}

func (v *ValidateError) Error() string {
	return fmt.Sprintf("validate %s: %s", v.Name, v.Err)
}

func (v *ValidateError) Unwrap() error {
	return v.Err
}

func (v *ValidateError) Is(target error) bool {
	t, ok := target.(*ValidateError)
	return ok &&
		(t.Name == "" || v.Name == t.Name) &&
		(t.Err == nil || v.Err == t.Err)
}

// Mismatch returns whether the error is just about mismatch.
// This value will be true if file does not exist, or if it does, but hash does not match.
// In any other case, the value is false.
func (v *ValidateError) Mismatch() bool {
	if utils.DoesNotExist(v.Err) {
		return true
	}

	return errors.Is(v.Err, &HashMatchError{})
}

// FIXME: missing documentation

type InvalidHashSumError struct {
	Sum []byte
	Err error
}

func (i *InvalidHashSumError) Error() string {
	if i.Err != nil {
		return fmt.Sprintf("invalid input hash sum: %s", i.Err)
	}

	return "invalid input hash sum"
}

func (i *InvalidHashSumError) Unwrap() error {
	return i.Err
}

func (i *InvalidHashSumError) Is(target error) bool {
	t, ok := target.(*InvalidHashSumError)

	return ok &&
		(t.Sum == nil || bytes.Equal(i.Sum, t.Sum)) &&
		(t.Err == nil || i.Err == t.Err)
}

func validationFail(name string, err error) *ValidateError {
	return &ValidateError{name, err}
}

func ValidateFileHex(name string, hash hash.Hash, expected string) error {
	var input []byte = nil

	decodedSum, decodeErr := hex.DecodeString(expected)

	if decodeErr != nil {
		return validationFail(name, &InvalidHashSumError{
			Sum: nil,
			Err: decodeErr,
		})
	}

	input = decodedSum

	return ValidateFile(name, hash, input)
}

func ValidateFile(name string, hash hash.Hash, expected []byte) error {
	if expected == nil {
		return validationFail(name, &InvalidHashSumError{
			Sum: expected,
			Err: errors.New("expected sum cannot be nil"),
		})
	}

	file, openErr := os.Open(name)

	if openErr != nil {
		return validationFail(name, openErr)
	}

	defer file.Close()

	hash.Reset()

	if _, copyErr := io.Copy(hash, file); copyErr != nil {
		return validationFail(name, copyErr)
	}

	actual := hash.Sum(nil)

	if !bytes.Equal(actual, expected) {
		return validationFail(name, &HashMatchError{
			Expected: expected,
			Actual:   actual,
		})
	}

	return nil
}
