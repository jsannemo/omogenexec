package main

import (
	"fmt"
	"github.com/google/logger"
	apipb "github.com/jsannemo/omogenexec/api"
	"github.com/jsannemo/omogenexec/util"
	"path"
	"path/filepath"
)

type Compilation struct {
	// This is unset if the Compilation failed.
	Program        *apipb.CompiledProgram
	CompilerErrors string
}

type CompileFunc func(program *apipb.Program, outputBase util.FileBase) (*Compilation, error)

func Compile(program *apipb.Program, outputPath string) (*Compilation, error) {
	langs := GetLanguages()
	lang, found := langs[program.Language]
	if !found {
		logger.Fatalf("Could not find submission language %v", program.Language)
	}
	fb := util.NewFileBase(outputPath)
	fb.GroupWritable = true
	fb.OwnerGid = util.OmogenexecGroupId()
	if err := fb.Mkdir("."); err != nil {
		return nil, fmt.Errorf("failed mkdir %s: %v", outputPath, err)
	}
	for _, file := range program.Sources {
		err := fb.WriteFile(file.Path, file.Contents)
		if err != nil {
			return nil, fmt.Errorf("failed writing %s: %v", file.Path, err)
		}
	}
	return lang.Compile(program, fb)
}

// NoCompile represents compilation that only copies some of the source files and uses the given
// Run command to execute the program.
func NoCompile(runCommand []string, include func(string) bool) CompileFunc {
	return func(program *apipb.Program, outputBase util.FileBase) (*Compilation, error) {
		var filteredPaths []string
		for _, file := range program.Sources {
			if include(file.Path) {
				filteredPaths = append(filteredPaths, file.Path)
			}
		}
		runCommand = substituteArgs(runCommand, filteredPaths)
		return &Compilation{
			Program: &apipb.CompiledProgram{
				ProgramRoot: outputBase.Path(),
				RunCommand:  runCommand,
			}}, nil
	}
}

func CppCompile(gppPath string) CompileFunc {
	return func(program *apipb.Program, outputBase util.FileBase) (*Compilation, error) {
		var filteredPaths []string
		for _, file := range program.Sources {
			if isCppFile(file.Path) {
				filteredPaths = append(filteredPaths, file.Path)
			}
		}
		sandboxArgs := sandboxForCompile(outputBase.Path())
		sandbox := newSandbox(0, sandboxArgs)
		err := sandbox.start()
		if err != nil {
			return nil, err
		}
		run, err := sandbox.Run(append([]string{gppPath}, substituteArgs(gppFlags, filteredPaths)...))
		sandbox.finish()
		if err != nil {
			return nil, fmt.Errorf("Sandbox failed: %v, %v", err, sandbox.sandboxErr.String())
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
	return ext == ".cc" || ext == ".cpp" || ext == ".h"
}

func sandboxForCompile(sourcePath string) sandboxArgs {
	return sandboxArgs{
		WorkingDirectory: sourcePath,
		InputPath:        "",
		OutputPath:       "",
		ErrorPath:        path.Join(sourcePath, "__compiler_errors"),
		ExtraReadPaths:   nil,
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
