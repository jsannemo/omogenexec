package eval

import (
	"fmt"
	"github.com/google/logger"
	apipb "github.com/jsannemo/omogenexec/api"
	"github.com/jsannemo/omogenexec/util"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var languages = make(map[apipb.LanguageGroup]*Language)

func InitLanguages() {
	gpp, err := initGpp()
	if err != nil {
		logger.Fatalf("Failure during C++17 initialization: %v", err)
	}
	languages[gpp.Info.Group] = gpp
	pypy, err := initPython()
	if err != nil {
		logger.Fatalf("Failure during PyPy3 initialization: %v", err)
	}
	languages[pypy.Info.Group] = pypy
	ruby, err := initRuby()
	if err != nil {
		logger.Fatalf("Failure during Ruby initialization: %v", err)
	}
	languages[ruby.Info.Group] = ruby
	rust, err := initRust()
	if err != nil {
		logger.Fatalf("Failure during Rust initialization: %v", err)
	}
	languages[rust.Info.Group] = rust
	java, err := initJava()
	if err != nil {
		logger.Fatalf("Failure during Java initialization: %v", err)
	}
	languages[java.Info.Group] = java
	csharp, err := initCsharp()
	if err != nil {
		logger.Fatalf("Failure during C# initialization: %v", err)
	}
	languages[csharp.Info.Group] = csharp
}

type Language struct {
	Info      *apipb.Language
	Container string
	compile   compileFunc
}

// GetLanguages returns all installed languages, mapped from language ID to the language itself.
func GetLanguages() map[apipb.LanguageGroup]*Language {
	return languages
}

const pypyPath = "/usr/bin/pypy3"

func initPython() (*Language, error) {
	logger.Infof("Checking for Python version")
	version, err := runCommandInSandbox([]string{pypyPath, "--version"}, apipb.LanguageGroup_PYTHON_3)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve version for python: %v", err)
	}
	logger.Infof("Using Python %s", version)
	return &Language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_PYTHON_3,
			Version:                version,
			CompilationDescription: nil,
			RunDescription:         fmt.Sprintf("pypy3 {files}"),
		},
		compile: noCompile([]string{pypyPath, "{files}"}, func(s string) bool {
			return filepath.Ext(s) == ".py"
		}, apipb.LanguageGroup_PYTHON_3),
	}, nil
}

const rubyPath = "/usr/bin/ruby"

func initRuby() (*Language, error) {
	logger.Infof("Checking Ruby version")
	version, err := runCommandInSandbox([]string{rubyPath, "--version"}, apipb.LanguageGroup_RUBY)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve version for python: %v", err)
	}
	logger.Infof("Using Ruby %s", version)
	return &Language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_RUBY,
			Version:                version,
			CompilationDescription: nil,
			RunDescription:         fmt.Sprintf("ruby {files}"),
		},
		compile: noCompile([]string{rubyPath, "{files}"}, func(s string) bool {
			return filepath.Ext(s) == ".rb"
		}, apipb.LanguageGroup_RUBY),
	}, nil
}

const gppPath = "/usr/bin/g++"

var gppFlags = []string{"-std=gnu++23", "-static", "-O2", "-o", "executable", "{files}"}
var isCppFile = hasExt([]string{".cpp", ".cc"})

func initGpp() (*Language, error) {
	logger.Infof("Checking for g++ version")
	version, err := runCommandInSandbox([]string{gppPath, "--version"}, apipb.LanguageGroup_CPP)
	if err != nil {
		return nil, fmt.Errorf("could not get g++ version: %v", err)
	}
	logger.Infof("Using g++ %s", version)
	return &Language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_CPP,
			Version:                version,
			CompilationDescription: []string{fmt.Sprintf("g++ %s", gppFlags)},
			RunDescription:         "./executable",
		},
		compile: simpleCompile(gppPath, gppFlags, []string{"./executable"}, isCppFile, apipb.LanguageGroup_CPP),
	}, nil
}

const rustPath = "/usr/bin/rustc"

var rustFlags = []string{"-O", "--crate-type", "bin", "--edition", "2021", "-o", "executable", "{files}"}
var isRustFile = hasExt([]string{".rs"})

func initRust() (*Language, error) {
	logger.Infof("Checking for rustc version")
	version, err := runCommandInSandbox([]string{rustPath, "--version"}, apipb.LanguageGroup_RUST)
	if err != nil {
		return nil, fmt.Errorf("could not get rustc version: %v", err)
	}
	logger.Infof("Using rustc %s", version)
	return &Language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_RUST,
			Version:                version,
			CompilationDescription: []string{fmt.Sprintf("rustc %s", rustFlags)},
			RunDescription:         "./executable",
		},
		compile: simpleCompile(rustPath, rustFlags, []string{"./executable"}, isRustFile, apipb.LanguageGroup_RUST),
	}, nil
}

var javacFlags = []string{"-d", ".", "{files}"}
var javaFlags = []string{"-cp", ".", "{main}"}

func initJava() (*Language, error) {
	logger.Infof("Checking for java version")
	version, err := runCommandInSandbox([]string{"/usr/bin/javac", "-version"}, apipb.LanguageGroup_JAVA)
	if err != nil {
		return nil, fmt.Errorf("could not get java version: %v", err)
	}
	logger.Infof("Using java %s", version)
	return &Language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_JAVA,
			Version:                version,
			CompilationDescription: []string{fmt.Sprintf("javac %s", javacFlags)},
			RunDescription:         fmt.Sprintf("java %s", javaFlags),
		},
		compile: javaCompile(javacFlags, javaFlags, hasExt([]string{".java"}), apipb.LanguageGroup_JAVA),
	}, nil
}

const monoPath = "/usr/bin/mono-csc"

var csharpFlags = []string{"-r:System.Numerics", "-out:executable", "{files}"}
var isCsharpFile = hasExt([]string{".cs"})

func initCsharp() (*Language, error) {
	logger.Infof("Checking for csharp version")
	version, err := runCommandInSandbox([]string{monoPath, "--version"}, apipb.LanguageGroup_CSHARP)
	if err != nil {
		return nil, fmt.Errorf("could not get csharp version: %v", err)
	}
	logger.Infof("Using csharp %s", version)
	return &Language{
		Info: &apipb.Language{
			Group:                  apipb.LanguageGroup_CSHARP,
			Version:                version,
			CompilationDescription: []string{fmt.Sprintf("mono-csc %s", csharpFlags)},
			RunDescription:         "./executable",
		},
		compile: simpleCompile(monoPath, csharpFlags, []string{"/usr/bin/mono", "./executable"}, isCsharpFile, apipb.LanguageGroup_CSHARP),
	}, nil
}

func hasExt(exts []string) func(string) bool {
	return func(search string) bool {
		ext := filepath.Ext(search)
		for _, pat := range exts {
			if pat == ext {
				return true
			}
		}
		return false
	}
}

func runCommandInSandbox(command []string, lang apipb.LanguageGroup) (string, error) {
	tmpdir, err := os.MkdirTemp("", "omogen")
	defer os.RemoveAll(tmpdir)
	base := util.NewFileBase(tmpdir)
	base.GroupWritable = true
	base.OwnerGid = util.OmogenexecGroupId()
	if err := base.FixOwners("."); err != nil {
		return "", err
	}
	if err := base.FixModeExec("."); err != nil {
		return "", err
	}
	if err := base.WriteFile("in", []byte{}); err != nil {
		return "", err
	}
	inputPath := path.Join(tmpdir, "in")
	outputPath := path.Join(tmpdir, "out")
	errorPath := path.Join(tmpdir, "err")
	args := sandboxArgs{
		InputPath:     inputPath,
		OutputPath:    outputPath,
		ErrorPath:     errorPath,
		TimeLimitMs:   10_000,
		MemoryLimitKb: 500_000,
	}
	setLanguageSandbox(&args, lang)
	sandbox := newSandbox(0, args)
	if err := sandbox.Start(); err != nil {
		return "", fmt.Errorf("couldn't start version sandbox: %v", err)
	}
	run, err := sandbox.Run(command)
	if err != nil {
		return "", fmt.Errorf("failed running %v", err)
	}
	sandbox.Finish()
	outputLine, _ := firstLine(outputPath)
	errorLine, _ := firstLine(errorPath)
	if run.Crashed() || run.TimedOut() {
		return "", fmt.Errorf("command crashed %v (output %s, %s)", command, outputLine, errorLine)
	}
	if err != nil {
		return "", err
	}
	if outputLine != "" {
		return outputLine, nil
	}
	if err != nil {
		return "", err
	}
	if errorLine != "" {
		return errorLine, nil
	}
	return "", fmt.Errorf("no output from command")
}

func firstLine(path string) (string, error) {
	output, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed reading command output: %v", err)
	}
	temp := strings.Split(string(output), "\n")
	return temp[0], nil
}
