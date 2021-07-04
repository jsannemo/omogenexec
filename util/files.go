package util

import (
	"fmt"
	"github.com/google/logger"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileBase represents a given base directory, and allows operations to be performed on files and directories within it.
//
// It makes sure that all operations are only performed within that directory, providing extra protection against malicious
// input or erroneous operations.
type FileBase struct {
	base string
	// The group ID files should be owned by.
	OwnerGid int
	// The user ID files should be owned by.
	OwnerUid int
	// Whether newly created files should be marked as writable by its owner group or not.
	GroupWritable bool
}

// NewFileBase returns a FileBase for the given path. By default, the owner UID and GID of files will be the same as of the
// current process, and files will not be group writable.
func NewFileBase(path string) (FileBase, error) {
	fullPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return FileBase{}, err
	}
	return FileBase{
		base:          fullPath,
		OwnerGid:      os.Getgid(),
		OwnerUid:      os.Getuid(),
		GroupWritable: false,
	}, nil
}

// Path returns the base path that the FileBase represents.
func (fb *FileBase) Path() string {
	return fb.base
}

// SubBase returns a new FileBase where the given path has been traversed. The path must be a relative path and may not
// refer to a path above the relative root, i.e. "../" is not an allowed path.
func (fb FileBase) SubBase(path string) (FileBase, error) {
	nbase, err := fb.FullPath(path)
	if err != nil {
		return FileBase{}, fmt.Errorf("could not make subbase to %s: %v", path, err)
	}
	fb.base = nbase
	return fb, nil
}

// FullPath returns the path of the file base after traversing the given relative path. If the path would result in a
// path above the file base, an error is returned.
func (fb *FileBase) FullPath(subPath string) (string, error) {
	fullPath, err := filepath.EvalSymlinks(filepath.Join(fb.base, subPath))
	if err != nil {
		return "", err
	}
	baseComponents := filepath.SplitList(fb.base)
	newComponents := filepath.SplitList(fullPath)
	if len(newComponents) < len(baseComponents) {
		return "", fmt.Errorf("path %s traverses upwards (too few components)", subPath)
	}
	for i := 0; i < len(baseComponents); i++ {
		if baseComponents[i] != newComponents[i] {
			return "", fmt.Errorf("path %s traverses upwards from %s (%s != %s)", subPath, fb.base, baseComponents[i], newComponents[i])
		}
	}
	return fullPath, nil
}

// ReadFile returns the contents of the file at the given subpath.
func (fb *FileBase) ReadFile(subPath string) ([]byte, error) {
	npath, err := fb.FullPath(subPath)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(npath)
}

// LinkInto links the file at the absolute path formPath to the given file base subpath.
func (fb *FileBase) LinkInto(fromPath string, toSubPath string) error {
	npath, err := fb.FullPath(toSubPath)
	if err != nil {
		return err
	}
	return os.Link(fromPath, npath)
}

// RemoveContents removes all files and directories within the file base.
func (fb *FileBase) RemoveContents(subPath string) error {
	npath, err := fb.FullPath(subPath)
	if err != nil {
		return err
	}
	dir, err := ioutil.ReadDir(npath)
	if err != nil {
		return err
	}
	for _, d := range dir {
		if err := os.RemoveAll(filepath.Join(npath, d.Name())); err != nil {
			return err
		}
	}
	return nil
}

// Exists checks if the given subpath exists.
func (fb *FileBase) Exists(subPath string) (bool, error) {
	npath, err := fb.FullPath(subPath)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(npath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		// stat should never fail
		logger.Fatalf("stat failed: %v", err)
	}
	return true, nil
}

// Copy copies the file at the absolute source path into the given filebase subpath.
func (fb *FileBase) Copy(srcFile, dstFile string) error {
	npath, err := fb.FullPath(dstFile)
	if err != nil {
		return err
	}
	out, err := os.Create(npath)
	defer out.Close()
	if err != nil {
		return err
	}

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	if err := fb.FixOwners(dstFile); err != nil {
		return err
	}
	if err := fb.FixMode(dstFile); err != nil {
		return err
	}
	return nil
}

// FixOwners updates the file at the given subpath to have the correct owner user and group.
func (fb *FileBase) FixOwners(subPath string) error {
	npath, err := fb.FullPath(subPath)
	if err != nil {
		return err
	}
	return os.Chown(npath, fb.OwnerUid, fb.OwnerGid)
}

// SetMode sets the mode of a file at the given subpath.
func (fb *FileBase) SetMode(subPath string, mode os.FileMode) error {
	npath, err := fb.FullPath(subPath)
	if err != nil {
		return err
	}
	return os.Chmod(npath, mode)
}

// FixModeExec updates the file at the given subpath to be rwx for its user and rx (or rwx depending on if the base is GroupWritable) for its group.
func (fb *FileBase) FixModeExec(subPath string) error {
	mode := 0750
	if fb.GroupWritable {
		mode = mode | 0020
	}
	return fb.SetMode(subPath, os.FileMode(mode))
}

// FixMode updates the file at the given subpath to be rw for its user and r (or rw depending on if the base is GroupWritable) for its group.
func (fb FileBase) FixMode(subPath string) error {
	mode := 0640
	if fb.GroupWritable {
		mode = mode | 0020
	}
	return fb.SetMode(subPath, os.FileMode(mode))
}

// Mkdir creates a directory at the given subpath.
func (fb *FileBase) Mkdir(subPath string) error {
	npath, err := fb.FullPath(subPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(npath, 0750); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	if err := fb.FixOwners(subPath); err != nil {
		return err
	}
	if err := fb.FixModeExec(subPath); err != nil {
		return err
	}
	return nil
}

// WriteFile writes binary data to the file at the given subpath.
func (fb *FileBase) WriteFile(subPath string, contents []byte) error {
	npath, err := fb.FullPath(subPath)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(npath, contents, 0750); err != nil {
		return err
	}
	if err := fb.FixOwners(subPath); err != nil {
		return err
	}
	if err := fb.FixMode(subPath); err != nil {
		return err
	}
	return nil
}

// CopyInto copies the contents of the source directory into the file base.
func (fb *FileBase) CopyInto(srcDir string) error {
	entries, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(srcDir, entry.Name())
		if entry.IsDir() {
			if err := fb.Mkdir(entry.Name()); err != nil {
				return err
			}
			subfb, err := fb.SubBase(entry.Name())
			if err != nil {
				return err
			}
			if err := subfb.CopyInto(filepath.Join(srcDir, entry.Name())); err != nil {
				return err
			}
		} else {
			if err := fb.Copy(sourcePath, entry.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Remove removes the file at the given subpath.
func (fb *FileBase) Remove(subPath string) error {
	npath, err := fb.FullPath(subPath)
	if err != nil {
		return err
	}
	return os.Remove(npath)
}
