package eval

import (
	"fmt"
	apipb "github.com/jsannemo/omogenexec/api"
	"github.com/jsannemo/omogenexec/util"
	"path"
	"path/filepath"
)

// A Compilation is the result of compiling program sources.
type Compilation struct {
	// This is unset if and only if the Compilation failed.
	Program        *apipb.CompiledProgram
	CompilerErrors string
}

type compileFunc func(program *apipb.Program, outputBase util.FileBase) (*Compilation, error)

func Compile(program *apipb.Program, outputPath string) (*Compilation, error) {
	langs := GetLanguages()
	lang, found := langs[program.Language]
	if !found {
		return nil, fmt.Errorf("could not find language %v", program.Language)
	}
	fb := util.NewFileBase(outputPath)
	fb.GroupWritable = true
	fb.OwnerGid = util.OmogenexecGroupId()
	if err := fb.Mkdir("."); err != nil {
		return nil, fmt.Errorf("failed compile output mkdir %s: %v", outputPath, err)
	}
	for _, file := range program.Sources {
		err := fb.WriteFile(file.Path, file.Contents)
		if err != nil {
			return nil, fmt.Errorf("failed writing source %s: %v", file.Path, err)
		}
	}
	return lang.compile(program, fb)
}

// noCompile represents compilation that only copies some of the source files and uses the given
// run command to execute the program.
func noCompile(runCommand []string, include func(string) bool) compileFunc {
	return func(program *apipb.Program, outputBase util.FileBase) (*Compilation, error) {
		var filteredPaths []string
		for _, file := range program.Sources {
			if include(file.Path) {
				filteredPaths = append(filteredPaths, file.Path)
			}
		}
		if len(filteredPaths) == 0 {
			return &Compilation{CompilerErrors: "No valid source files found"}, nil
		}
		runCommand = substituteArgs(runCommand, filteredPaths)
		return &Compilation{
			Program: &apipb.CompiledProgram{
				ProgramRoot: outputBase.Path(),
				RunCommand:  runCommand,
			}}, nil
	}
}

func cppCompile(gppPath string) compileFunc {
	return func(program *apipb.Program, outputBase util.FileBase) (*Compilation, error) {
		var filteredPaths []string
		for _, file := range program.Sources {
			if isCppFile(file.Path) {
				filteredPaths = append(filteredPaths, file.Path)
			}
		}
		sandboxArgs := sandboxForCompile(outputBase.Path())
		sandbox := newSandbox(0, sandboxArgs)
		err := sandbox.Start()
		if err != nil {
			return nil, fmt.Errorf("couldn't start compilation sandbox: %v", err)
		}
		run, err := sandbox.Run(append([]string{gppPath}, substituteArgs(gppFlags, filteredPaths)...))
		sandbox.Finish()
		if err != nil {
			return nil, fmt.Errorf("sandbox failed: %v, %v", err, sandbox.sandboxErr.String())
		}
		stderr, err := outputBase.ReadFile("__compiler_errors")
		if err != nil {
			return nil, fmt.Errorf("could not read compiler errors: %v", err)
		}
		if !run.CrashedWith(0) {
			return &Compilation{
				CompilerErrors: string(stderr),
			}, nil
		}
		return &Compilation{
			Program: &apipb.CompiledProgram{
				ProgramRoot: outputBase.Path(),
				RunCommand:  []string{"./a.out"},
			}}, nil
	}
}

func isCppFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".cc" || ext == ".cpp"
}

func sandboxForCompile(sourcePath string) sandboxArgs {
	return sandboxArgs{
		WorkingDirectory: sourcePath,
		InputPath:        "",
		OutputPath:       "",
		ErrorPath:        path.Join(sourcePath, "__compiler_errors"),
		ExtraReadPaths:   []string{"/usr/include"},
		ExtraWritePaths:  []string{sourcePath},
		TimeLimitMs:      60 * 1000,
		MemoryLimitKb:    1000 * 1000,
	}
}

func substituteArgs(args []string, paths []string) []string {
	var newArgs []string
	for _, arg := range args {
		if arg == "{files}" {
			newArgs = append(newArgs, paths...)
		} else {
			newArgs = append(newArgs, arg)
		}
	}
	return newArgs
}
