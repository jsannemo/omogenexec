package main

import (
	"fmt"
	"github.com/google/logger"
	apipb "github.com/jsannemo/omogenexec/api"
	"github.com/jsannemo/omogenexec/util"
	"os/exec"
	"path/filepath"
)

var languages = make(map[apipb.LanguageGroup]*language)

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

type language struct {
	// This is suitable for inclusion in URLs, and can be displayed externally.
	Info    *apipb.Language
	Compile CompileFunc
}

// GetLanguages returns all installed languages, mapped from language ID to the language itself.
func GetLanguages() map[apipb.LanguageGroup]*language {
	return languages
}

func initPython(executable string) (*language, error) {
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
	return &language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_PYTHON_3,
			Version:                version,
			CompilationDescription: nil,
			RunDescription:         fmt.Sprintf("%s {files}", executable),
		},
		Compile: NoCompile([]string{executable, "{files}", executable}, func(s string) bool {
			return filepath.Ext(s) == ".py"
		}),
	}, nil
}

var gppFlags = []string{"-std=gnu++17", "-static", "-O2", "{files}"}

func initGpp() (*language, error) {
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
	return &language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_CPP,
			Version:                version,
			CompilationDescription: []string{fmt.Sprintf("g++ %s", gppFlags)},
			RunDescription:         "./a.out",
		},
		Compile: CppCompile(realPath),
	}, nil
}
