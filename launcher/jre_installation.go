package launcher

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brawaru/marct/launcher/download"
	"github.com/brawaru/marct/launcher/java"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/validfile"
	"github.com/itchio/lzma"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// an enum of file download states
type FileState int

const (
	FileStateUnknown       FileState = iota // File has not yet been checked
	FileStateNotDownloaded                  // File has not yet been downloaded
	FileStateCorrupted                      // File is corrupted and must be deleted
	FileStateDownloaded                     // File is downloaded and ready to be mapped
	FileStateReady                          // File is mapped and ready
)

type JREObject struct {
	JavaFile                    // File to download
	Destination       string    // Where the object is stored
	ObjectDestination string    // Where the unmapped object is stored
	State             FileState // State of the file
	IsRaw             bool      // Whether the file stores is raw or compressed
}

type JreInstallation struct {
	Classifier string
	Selector   string
	Descriptor *JavaVersionDescriptor
	Manifest   *JavaManifest
	Path       string // Root path of the installation

	stagingPath string                // Path to the staging directory where all the objects are downloaded
	objects     map[string]*JREObject // All objects in the installation
	filesPath   string                // Path to the files' directory where the final files are stored
}

func NewInstallation(classifier string, selector string, descriptor *JavaVersionDescriptor, path string) *JreInstallation {
	stagingPath := filepath.Join(path, classifier+"_staging")
	filesPath := filepath.Join(path, classifier)

	return &JreInstallation{
		Classifier:  classifier,
		Selector:    selector,
		Descriptor:  descriptor,
		Path:        path,
		objects:     make(map[string]*JREObject),
		stagingPath: stagingPath,
		filesPath:   filesPath,
	}
}

func (i *JreInstallation) fetchManifest() error {
	if i.Descriptor == nil {
		return errors.New("descriptor is nil")
	}

	manifestPath := filepath.Join(i.Path, ".manifest")

	if dl, err := download.WithSHA1(i.Descriptor.Manifest.URL, manifestPath, i.Descriptor.Manifest.SHA1); err == nil {
		if dlErr := dl.Download(); dlErr != nil {
			return fmt.Errorf("failed to download manifest: %w", dlErr)
		}
	} else {
		return fmt.Errorf("cannot create download: %w", err)
	}

	if manifestBytes, manifestReadErr := os.ReadFile(manifestPath); manifestReadErr != nil {
		return fmt.Errorf("could not read manifest file: %w", manifestReadErr)
	} else if manifestUnmarshalErr := json.Unmarshal(manifestBytes, &i.Manifest); manifestUnmarshalErr != nil {
		return fmt.Errorf("could not unmarshal manifest file: %w", manifestUnmarshalErr)
	}

	return nil
}

func (i *JreInstallation) prepareObjects() error {
	if i.Manifest == nil {
		return errors.New("manifest is nil")
	}

	i.objects = make(map[string]*JREObject, 0)

	for dest, file := range i.Manifest.Files {
		i.objects[dest] = &JREObject{
			JavaFile:    file,
			Destination: filepath.Join(i.filesPath, filepath.FromSlash(dest)),
			State:       FileStateUnknown,
		}
	}

	return nil
}

func (i *JreInstallation) validateObject(fp string, object *JREObject) error {
	state := FileStateNotDownloaded

	switch object.Type {
	case java.TypeDir:
		if exists, statErr := validfile.DirExists(object.Destination); statErr != nil {
			if errors.Is(statErr, &validfile.WrongPathType{}) {
				state = FileStateCorrupted
			} else {
				return fmt.Errorf("dir %q stat failed: %w", object.Destination, statErr)
			}
		} else if exists {
			state = FileStateReady
		}
	case java.TypeLink:
		tfp := path.Clean(path.Join(path.Dir(fp), object.Target))

		target, exists := i.objects[tfp]

		if !exists {
			return fmt.Errorf("object %q targets %q, which is not present in object set", fp, tfp)
		} else {
			if runtime.GOOS == "windows" {
				// cannot check where hardlink points to, so have to actually verify the file
				if pathExists, err := validfile.PathExists(object.Destination); err != nil {
					return fmt.Errorf("path %q stat failed: %w", object.Destination, err)
				} else if pathExists {
					rawDl := target.Downloads["raw"]

					if validateErr := validfile.ValidateFileHex(object.Destination, sha1.New(), rawDl.SHA1); validateErr != nil {
						if !utils.DoesNotExist(validateErr) {
							var v *validfile.ValidateError
							if !errors.As(validateErr, &v) || v.Mismatch() {
								state = FileStateCorrupted // either refers to a different file or is actually corrupted
							} else {
								return fmt.Errorf("cannot validate linked file %q: %w", object.Destination, validateErr)
							}
						}
					} else {
						state = FileStateReady
					}
				} else {
					state = FileStateNotDownloaded
				}
			} else {
				linkDestination, err := os.Readlink(object.Destination)
				if err != nil {
					if !utils.DoesNotExist(err) {
						return fmt.Errorf("cannot read link %q: %w", object.Destination, err)
					}
				} else if linkDestination == object.Target || linkDestination == target.Destination {
					state = FileStateReady
				} else {
					state = FileStateCorrupted
				}
			}
		}
	case java.TypeFile:
		rawDl := object.Downloads["raw"]

		if validateErr := validfile.ValidateFileHex(object.Destination, sha1.New(), rawDl.SHA1); validateErr != nil {
			if !utils.DoesNotExist(validateErr) {
				var v *validfile.ValidateError
				if !errors.As(validateErr, &v) || v.Mismatch() {
					state = FileStateCorrupted
				} else {
					return fmt.Errorf("cannot validate file %q: %w", object.Destination, validateErr)
				}
			}
		} else {
			state = FileStateReady
		}
	}

	object.State = state

	return nil
}

func (i *JreInstallation) validateObjects() (bool, error) {
	success := true

	for fp, object := range i.objects {
		if err := i.validateObject(fp, object); err != nil {
			return false, fmt.Errorf("failed to validate object %q: %w", fp, err)
		}

		if object.State != FileStateReady {
			success = false
		}
	}

	return success, nil
}

func (i *JreInstallation) deleteCorrupted() error {
	for _, object := range i.objects {
		if object.State == FileStateCorrupted {
			if err := os.RemoveAll(object.Destination); err != nil {
				return fmt.Errorf("cannot delete corrupted file %q: %w", object.Destination, err)
			}
		}
	}

	return nil
}

func (i *JreInstallation) downloadFiles() error {
	for fp, object := range i.objects {
		if !object.Type.IsFile() {
			continue
		}

		var objectDl Download

		isRaw := false

		if lzmaDl, hasLzma := object.Downloads["lzma"]; hasLzma {
			objectDl = lzmaDl
		} else if rawDl, hasRaw := object.Downloads["raw"]; hasRaw {
			objectDl = rawDl
			isRaw = true
		} else {
			return fmt.Errorf("file %q has no downloads?! O_o", fp)
		}

		dest := filepath.Join(i.stagingPath, objectDl.SHA1)

		if dl, err := download.WithSHA1(objectDl.URL, dest, objectDl.SHA1); err == nil {
			if dlErr := dl.Download(); dlErr != nil {
				return dlErr
			}
		} else {
			return err
		}

		object.IsRaw = isRaw
		object.ObjectDestination = dest
		object.State = FileStateDownloaded
	}

	return nil
}

func (i *JreInstallation) mapDir(_ string, object *JREObject) error {
	if mdErr := os.MkdirAll(object.Destination, 0755); mdErr != nil {
		return fmt.Errorf("cannot create directory %q: %w", object.Destination, mdErr)
	}

	return nil
}

func (i *JreInstallation) mapLink(fp string, object *JREObject) error {
	tfp := path.Clean(path.Join(path.Dir(fp), object.Target))

	target, exists := i.objects[tfp]

	if !exists {
		return fmt.Errorf("expected %q to exist", tfp)
	}

	if target.State != FileStateReady {
		if mapErr := i.mapObject(tfp, target); mapErr != nil {
			return fmt.Errorf("target %q cannot be mapped: %w", tfp, mapErr)
		}
	}

	if err := os.MkdirAll(filepath.Dir(object.Destination), 0755); err != nil {
		return fmt.Errorf("cannot create parent directories for %q: %w", target.Destination, err)
	}

	if runtime.GOOS == "windows" {
		// on Windows, usage of symbolic links requires elevated privileges, thus we use hard links instead
		if linkErr := os.Link(target.Destination, object.Destination); linkErr != nil {
			return fmt.Errorf("hardlink failed: %w", linkErr)
		}
	} else {
		if linkErr := os.Symlink(object.Target, object.Destination); linkErr != nil {
			return fmt.Errorf("symlink failed: %w", linkErr)
		}
	}

	return nil
}

func (i *JreInstallation) mapFile(_ string, object *JREObject) error {
	if dir := filepath.Dir(object.Destination); dir != "." {
		if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
			return fmt.Errorf("cannot create parent directories for %q: %w", object.Destination, mkdirErr)
		}
	}

	file, createErr := os.Create(object.Destination)
	if createErr != nil {
		return fmt.Errorf("cannot create file %q: %w", object.Destination, createErr)
	}

	objectFile, openErr := os.Open(object.ObjectDestination)
	if openErr != nil {
		return fmt.Errorf("cannot open file %q: %w", object.ObjectDestination, openErr)
	}

	srcReader := io.Reader(objectFile)

	if !object.IsRaw {
		srcReader = lzma.NewReader(objectFile)
	}

	if _, copyErr := io.Copy(file, srcReader); copyErr != nil {
		return fmt.Errorf("cannot copy %q to %q: %w", object.ObjectDestination, object.Destination, copyErr)
	}

	utils.DClose(objectFile)
	utils.DClose(file)

	rawDl := object.Downloads["raw"]

	if validateErr := validfile.ValidateFileHex(object.Destination, sha1.New(), rawDl.SHA1); validateErr != nil {
		return validateErr
	}

	if object.Executable {
		if stat, statErr := os.Stat(object.Destination); statErr != nil {
			return fmt.Errorf("cannot stat file %q: %w", object.Destination, statErr)
		} else if chmodErr := os.Chmod(object.Destination, stat.Mode()|0b1000000); chmodErr != nil {
			return fmt.Errorf("cannot mark file %q as executable: %w", object.Destination, chmodErr)
		}
	}

	return nil
}

func (i *JreInstallation) mapObject(fp string, object *JREObject) error {
	if object.State == FileStateReady {
		return nil
	}

	switch object.Type {
	case java.TypeLink:
		if err := i.mapLink(fp, object); err != nil {
			return err
		}
	case java.TypeFile:
		if err := i.mapFile(fp, object); err != nil {
			return err
		}
	case java.TypeDir:
		if err := i.mapDir(fp, object); err != nil {
			return err
		}
	}

	object.State = FileStateReady

	return nil
}

func (i *JreInstallation) mapObjects() error {
	for fp, object := range i.objects {
		if err := i.mapObject(fp, object); err != nil {
			return fmt.Errorf("cannot map %q: %w", fp, err)
		}
	}

	return nil
}

func (i *JreInstallation) deleteStaging() error {
	if err := os.RemoveAll(i.stagingPath); err != nil {
		return fmt.Errorf("cannot delete staging directory %q: %w", i.stagingPath, err)
	}

	return nil
}

func (i *JreInstallation) writeVersion() error {
	versionFile, createErr := os.Create(filepath.Join(i.Path, ".version"))

	if createErr == nil {
		defer utils.DClose(versionFile)

		if _, writeErr := versionFile.WriteString(i.Descriptor.Version.Name); writeErr != nil {
			return fmt.Errorf("cannot write version file: %w", writeErr)
		}
	}

	return createErr
}

type PostValidationError struct {
	BadObjects []JREObject
}

func (p *PostValidationError) Error() string {
	return fmt.Sprintf("%d objects failed post-validation", len(p.BadObjects))
}

func (i *JreInstallation) Install() error {
	// 0. fetch manifest
	if i.Manifest == nil {
		if err := i.fetchManifest(); err != nil {
			return err
		}
	}

	if err := i.prepareObjects(); err != nil {
		return fmt.Errorf("cannot prepare objects: %w", err)
	}

	// 1. validate all files

	if allValid, err := i.validateObjects(); err != nil {
		return fmt.Errorf("cannot validate objects: %w", err)
	} else if allValid {
		return nil
	}

	// 2. delete corrupted ones

	if err := i.deleteCorrupted(); err != nil {
		return fmt.Errorf("cannot delete corrupted files: %w", err)
	}

	// 3. download missing ones

	if err := i.downloadFiles(); err != nil {
		return fmt.Errorf("cannot download files: %w", err)
	}

	// 4. map them

	if err := i.mapObjects(); err != nil {
		return fmt.Errorf("cannot map objects: %w", err)
	}

	// 5. delete staging directory

	if err := i.deleteStaging(); err != nil {
		return fmt.Errorf("cannot delete staging directory: %w", err)
	}

	// 6. write version file

	if err := i.writeVersion(); err != nil {
		return fmt.Errorf("cannot write version file: %w", err)
	}

	if allValid, err := i.validateObjects(); err != nil {
		return err
	} else if !allValid {
		var faulty []JREObject
		for _, object := range i.objects {
			if object.State != FileStateReady {
				faulty = append(faulty, *object)
			}
		}
		return &PostValidationError{faulty}
	}

	return nil
}
