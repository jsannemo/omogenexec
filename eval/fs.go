package eval

import (
	"fmt"
	"github.com/google/logger"
	omogenrunner "github.com/jsannemo/omogenexec/api"
	"io/ioutil"
	"os"
	"path"
)

const fsPath = "/var/lib/omogen/fs"

var readMounts = make(map[omogenrunner.LanguageGroup][]string)

func init() {
	readMounts[omogenrunner.LanguageGroup_CPP] = makeMounts("cpp")
	readMounts[omogenrunner.LanguageGroup_CSHARP] = makeMounts("csharp")
	readMounts[omogenrunner.LanguageGroup_GO] = makeMounts("go")
	readMounts[omogenrunner.LanguageGroup_JAVA] = makeMounts("java")
	readMounts[omogenrunner.LanguageGroup_PYTHON_3] = makeMounts("python3")
	readMounts[omogenrunner.LanguageGroup_RUBY] = makeMounts("ruby")
	readMounts[omogenrunner.LanguageGroup_RUST] = makeMounts("rust")
}

func makeMounts(tag string) []string {
	langPath := path.Join(fsPath, tag)
	entries, err := ioutil.ReadDir(langPath)
	if err != nil {
		logger.Fatalf("failed reading fs for %s: %v", tag, err)
	}
	var mounts []string
	for _, entry := range entries {
		fullPath := path.Join(langPath, entry.Name())
		if entry.IsDir() {
			mounts = append(mounts, fmt.Sprintf("%s:/%s", fullPath, entry.Name()))
		} else if entry.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(fullPath)
			if err != nil {
				logger.Fatalf("couldn't follow symlink")
			}
			newPath := path.Clean(path.Join(langPath, target))
			mounts = append(mounts, fmt.Sprintf("%s:/%s", newPath, entry.Name()))
		}
	}
	return mounts
}

func setLanguageSandbox(args *sandboxArgs, lang omogenrunner.LanguageGroup) {
	mounts, ok := readMounts[lang]
	if ok {
		args.SkipDefaultMounts = true
		args.ExtraReadPaths = append(args.ExtraReadPaths, mounts...)
	}
}
