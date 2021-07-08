package eval

import (
	"fmt"
	"github.com/google/logger"
	apipb "github.com/jsannemo/omogenexec/api"
	"github.com/jsannemo/omogenexec/util"
	"os/exec"
	"path/filepath"
)

var languages = make(map[apipb.LanguageGroup]*Language)

func init() {
	pypy, err := initPython("pypy3")
	if err != nil {
		logger.Fatalf("Failure during PyPy3 initialization: %v", err)
	}
	languages[pypy.Info.Group] = pypy
	gpp, err := initGpp()
	if err != nil {
		logger.Fatalf("Failure during C++17 initialization: %v", err)
	}
	languages[gpp.Info.Group] = gpp
}

type Language struct {
	Info    *apipb.Language
	compile compileFunc
}

// GetLanguages returns all installed languages, mapped from language ID to the language itself.
func GetLanguages() map[apipb.LanguageGroup]*Language {
	return languages
}

func initPython(executable string) (*Language, error) {
	logger.Infof("Checking for Python executable %s", executable)
	realPath, err := exec.LookPath(executable)
	if err != nil {
		return nil, fmt.Errorf("could not find python executable %s: %v", executable, err)
	}
	version, err := util.FirstLineFromCommand(realPath, []string{"--version"})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve version for python %s: %v", realPath, err)
	}
	logger.Infof("Using Python %s %s", realPath, version)
	return &Language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_PYTHON_3,
			Version:                version,
			CompilationDescription: nil,
			RunDescription:         fmt.Sprintf("%s {files}", executable),
		},
		compile: noCompile([]string{executable, "{files}", executable}, func(s string) bool {
			return filepath.Ext(s) == ".py"
		}),
	}, nil
}

var gppFlags = []string{"-std=gnu++17", "-static", "-O2", "{files}"}

func initGpp() (*Language, error) {
	logger.Infof("Checking for g++ executable")
	realPath, err := exec.LookPath("g++")
	if err != nil {
		return nil, fmt.Errorf("could not find g++: %v", err)
	}
	version, err := util.FirstLineFromCommand(realPath, []string{"--version"})
	if err != nil {
		return nil, fmt.Errorf("could not get g++ version: %v", err)
	}
	logger.Infof("Using g++ %s %s", realPath, version)
	return &Language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_CPP,
			Version:                version,
			CompilationDescription: []string{fmt.Sprintf("g++ %s", gppFlags)},
			RunDescription:         "./a.out",
		},
		compile: cppCompile(realPath),
	}, nil
}
