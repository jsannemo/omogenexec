package main

import (
	"github.com/google/logger"

	"github.com/jsannemo/omogenexec/util"
)

// fileLinker is used to swap easily link files into two directories, one for read-only files and one for writable files.
//
// The main use case it for easily swapping out readable and writable files that should have the
// same name within a container executing a program multiple times, such as input and output files.
// It keeps all links in a small number of directories, which makes it easy to clear up the file system
// environment between runs.
type fileLinker struct {
	readBase  *util.FileBase
	writeBase *util.FileBase
}

// NewFileLinker returns a new file linker, rooted at the given path.
func NewFileLinker(dir string) (*fileLinker, error) {
	base := util.NewFileBase(dir)
	base.OwnerGid = util.OmogenexecGroupId()
	if err := base.Mkdir("."); err != nil {
		return nil, err
	}
	reader, err := base.SubBase("read")
	if err != nil {
		return nil, err
	}
	writer, err := base.SubBase("write")
	if err != nil {
		return nil, err
	}
	linker := &fileLinker{
		readBase:  &reader,
		writeBase: &writer,
	}
	linker.writeBase.GroupWritable = true
	if err := linker.writeBase.Mkdir("."); err != nil {
		return nil, err
	}
	if err := linker.readBase.Mkdir("."); err != nil {
		return nil, err
	}
	return linker, nil
}

func (fl *fileLinker) base(writeable bool) *util.FileBase {
	if writeable {
		return fl.writeBase
	} else {
		return fl.readBase
	}
}

// PathFor returns the path that a file will get inside the linker.
func (fl *fileLinker) PathFor(inName string, writeable bool) string {
	str, err := fl.base(writeable).FullPath(inName)
	if err != nil {
		logger.Fatalf("Tried to use an env with relative path: %v", err)
	}
	return str
}

// LinkFile hard links the file path into the inside root.
func (fl *fileLinker) LinkFile(path, inName string, writeable bool) error {
	return fl.base(writeable).LinkInto(path, inName)
}

// clear resets the environment for a new execution.
func (fl *fileLinker) clear() error {
	rerr := fl.readBase.RemoveContents(".")
	werr := fl.writeBase.RemoveContents(".")
	if rerr != nil {
		return rerr
	}
	if werr != nil {
		return werr
	}
	return nil
}
