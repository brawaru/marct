package validfile

import (
	"errors"
	"fmt"
	"github.com/brawaru/marct/utils"
	"os"
)

type StatErr struct {
	Name string
	Err  error
}

func (s *StatErr) Error() string {
	return fmt.Sprintf("failed to stat file \"%s\": %s", s.Name, s.Err)
}

func (s *StatErr) Unwrap() error {
	return s.Err
}

func (s *StatErr) Is(target error) bool {
	t, ok := target.(*StatErr)

	return ok &&
		(t.Name == "" || s.Name == t.Name) &&
		(t.Err != nil && errors.Is(s.Err, t.Err))
}

// DoesNotExist returns whether the error is related to file being absent rather than any other reason.
func (s *StatErr) DoesNotExist() bool {
	return utils.DoesNotExist(s.Err)
}

type WrongPathType struct {
	Name        string
	IsDirectory bool
}

func (f *WrongPathType) Error() string {
	if f.IsDirectory {
		return fmt.Sprintf("%s is a directory", f.Name)
	}

	return fmt.Sprintf("%s is a file", f.Name)
}

func (f *WrongPathType) Is(target error) bool {
	t, ok := target.(*WrongPathType)
	return ok && (t.Name == "" || f.Name == t.Name)
}

type exp int

const (
	expNone exp = iota
	expFile
	expDir
)

func TryStat(name string) (stat os.FileInfo, err error) {
	stat, err = os.Stat(name)

	if err != nil {
		err = &StatErr{name, err}
	}

	return
}

func wrappedStat(name string, exp exp) error {
	stat, statErr := TryStat(name)

	if statErr == nil {
		isDir := stat.IsDir()

		if exp == expFile && isDir {
			return &StatErr{name, &WrongPathType{name, isDir}}
		}
	}

	return statErr
}

// // TryStatDir queries information about directory at specified location.
// //
// // If query fails, the error is returned as is.
// //
// // If location does exist and is a file, a WrongPathType error is returned.
// //
// // nil is returned otherwise.
// func TryStatDir(name string) error {
// 	return wrappedStat(name, true)
// }
//
// // TryStatFile queries information about file at specified location.
// //
// // If query fails, the error is returned as is.
// //
// // If file does exist and is a directory, a WrongPathType error is returned.
// //
// // nil is returned otherwise.
// func TryStatFile(name string) error {
// 	return wrappedStat(name, false)
// }

func isExists(statErr error) (exists bool, err error) {
	exists = statErr == nil

	if !utils.DoesNotExist(statErr) {
		err = statErr
	}

	return
}

// DirExists uses TryStat function to check if the directory exists at location.
//
// If stat fails with error related to directory absence, it does not return that error,
// only false boolean; in any other case both false and the error are returned.
//
// If stat succeeds, only true boolean is returned.
func DirExists(name string) (exists bool, err error) {
	return isExists(wrappedStat(name, expDir))
}

// FileExists uses TryStat function to check if the file exists at location.
//
// If stat fails with error related to file absence, it does not return that error,
// only false boolean; in any other case both false and the error are returned.
//
// If stat succeeds, only true boolean is returned.
func FileExists(name string) (exists bool, err error) {
	return isExists(wrappedStat(name, expFile))
}

// PathExists uses TryStat function to check if the file or directory exists at location.
//
// If stat fails with error related to file or directory absence, it does not return that error,
// only false boolean; in any other case both false and the error are returned.
//
// If stat succeeds, only true boolean is returned.
func PathExists(name string) (exists bool, err error) {
	return isExists(wrappedStat(name, expNone))
}

// FIXME: documentation is lacking in places

func ValidateExistsFile(name string) (err error) {
	if err := wrappedStat(name, expFile); err != nil {
		return validationFail(name, err)
	}

	return nil
}
