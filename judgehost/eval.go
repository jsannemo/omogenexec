package main

import (
	"fmt"
	apipb "github.com/jsannemo/omogenexec/api"
	"github.com/jsannemo/omogenexec/util"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
)

type Evaluator struct {
	root           string
	linker         *fileLinker
	valLinker      *fileLinker
	plan           *apipb.EvaluationPlan
	evalCache      map[string]*apipb.Result
	programSandbox *Sandbox
	evalSandbox    *Sandbox
	resultChan     chan<- *apipb.Result
}

func NewEvaluator(root string, plan *apipb.EvaluationPlan, results chan<- *apipb.Result) (*Evaluator, error) {
	fl, err := NewFileLinker(filepath.Join(root, "env"))
	if err != nil {
		return nil, fmt.Errorf("failed creating fileLinker: %v", err)
	}
	eval := &Evaluator{
		root:       root,
		linker:     fl,
		plan:       plan,
		evalCache:  make(map[string]*apipb.Result),
		resultChan: results,
	}
	if plan.Validator != nil {
		valfl, err := NewFileLinker(filepath.Join(root, "valenv"))
		if err != nil {
			return nil, fmt.Errorf("failed creating validator fileLinker: %v", err)
		}
		eval.valLinker = valfl
	}
	return eval, nil
}

func (e *Evaluator) resetPermissions() error {
	cmd := exec.Command("/usr/bin/omogenexec-fixpermissions", "--path", filepath.Dir(e.root))
	return cmd.Run()
}

func (e *Evaluator) Evaluate() error {
	if err := e.resetPermissions(); err != nil {
		return fmt.Errorf("could not reset permissions: %v", err)
	}
	defer close(e.resultChan)
	outPath := e.linker.PathFor("output", true)
	e.programSandbox = newSandbox(0, sandboxArgs{
		WorkingDirectory: e.plan.Program.ProgramRoot,
		InputPath:        e.linker.PathFor("input", false),
		OutputPath:       e.linker.PathFor("output", true),
		ErrorPath:        e.linker.PathFor("error", true),
		ExtraReadPaths: []string{
			e.plan.Program.ProgramRoot,
		},
		ExtraWritePaths: nil,
		TimeLimitMs:     int(e.plan.TimeLimitMs),
		MemoryLimitKb:   int(e.plan.MemLimitKb),
	})
	if err := e.programSandbox.start(); err != nil {
		return fmt.Errorf("failed starting sandbox: %v", err)
	}
	if e.plan.Validator != nil {
		e.evalSandbox = newSandbox(1, sandboxArgs{
			WorkingDirectory: e.plan.Program.ProgramRoot,
			InputPath:        e.valLinker.PathFor("team_output", false),
			OutputPath:       e.valLinker.PathFor("output", true),
			ErrorPath:        e.valLinker.PathFor("error", true),
			ExtraReadPaths: []string{
				e.valLinker.readBase.Path(),
				e.plan.Validator.ProgramRoot,
			},
			ExtraWritePaths: []string{
				e.valLinker.writeBase.Path(),
			},
			// TODO: make this configurable
			TimeLimitMs:   60000,
			MemoryLimitKb: 1000 * 1000,
		})
	}
	_, err := e.evaluateGroup(e.plan.RootGroup, outPath)
	return err
}

type evalable struct {
	*apipb.TestGroup
	*apipb.TestCase
}

func evalableLess(a *evalable, b *evalable) bool {
	var aName, bName string
	if a.TestGroup != nil {
		aName = a.TestGroup.Name
	} else {
		aName = b.TestGroup.Name
	}
	if b.TestGroup != nil {
		bName = b.TestGroup.Name
	} else {
		bName = b.TestGroup.Name
	}
	return aName < bName
}

func worseness(v apipb.Verdict) int {
	switch v {
	case apipb.Verdict_ACCEPTED:
		return 0
	case apipb.Verdict_RUN_TIME_ERROR:
		return 1
	case apipb.Verdict_TIME_LIMIT_EXCEEDED:
		return 2
	case apipb.Verdict_WRONG_ANSWER:
		return 3
	}
	return -1
}

func mergeRes(res []*apipb.Result, tg *apipb.TestGroup) *apipb.Result {
	result := &apipb.Result{
		Type:        apipb.ResultType_TEST_GROUP,
		Verdict:     apipb.Verdict_ACCEPTED,
		Score:       0,
		TimeUsageMs: 0,
	}
	if tg.ScoringMode == apipb.ScoringMode_MIN {
		result.Score = math.Inf(1)
	} else if tg.ScoringMode == apipb.ScoringMode_MAX {
		result.Score = math.Inf(-1)
	}
	anyAccepted := false
	for _, res := range res {
		if res.Verdict == apipb.Verdict_ACCEPTED {
			anyAccepted = true
		} else if tg.VerdictMode == apipb.VerdictMode_WORST_ERROR && worseness(res.Verdict) > worseness(result.Verdict) {
			result.Verdict = res.Verdict
		} else if tg.VerdictMode == apipb.VerdictMode_FIRST_ERROR && result.Verdict == apipb.Verdict_ACCEPTED {
			result.Verdict = res.Verdict
		}

		if tg.ScoringMode == apipb.ScoringMode_SUM || tg.ScoringMode == apipb.ScoringMode_AVG {
			result.Score += res.Score
		} else if tg.ScoringMode == apipb.ScoringMode_MIN {
			result.Score = math.Min(result.Score, res.Score)
		} else if tg.ScoringMode == apipb.ScoringMode_MAX {
			result.Score = math.Max(result.Score, res.Score)
		}
	}

	if tg.VerdictMode == apipb.VerdictMode_ALWAYS_ACCEPT || (anyAccepted && tg.AcceptIfAnyAccepted) {
		result.Verdict = apipb.Verdict_ACCEPTED
	}
	return result
}

func (e *Evaluator) evaluateGroup(tg *apipb.TestGroup, outPath string) (*apipb.Result, error) {
	var evalables []evalable = nil
	for _, group := range tg.Groups {
		evalables = append(evalables, evalable{TestGroup: group})
	}
	for _, testcase := range tg.Cases {
		evalables = append(evalables, evalable{TestCase: testcase})
	}
	sort.Slice(evalables, func(i, j int) bool {
		return evalableLess(&evalables[i], &evalables[j])
	})

	var res []*apipb.Result
	for _, eval := range evalables {
		var subres *apipb.Result
		if group := eval.TestGroup; group != nil {
			var err error
			subres, err = e.evaluateGroup(group, outPath)
			if err != nil {
				return nil, err
			}
			res = append(res, subres)
		} else {
			var err error
			subres, err = e.evaluateCase(eval.TestCase, tg.OutputValidatorFlags, outPath)
			if err != nil {
				return nil, err
			}
			res = append(res, subres)
		}
		if subres.Verdict != apipb.Verdict_ACCEPTED && tg.BreakOnFail {
			break
		}
	}
	groupRes := mergeRes(res, tg)
	e.resultChan <- groupRes
	return groupRes, nil
}

func (e *Evaluator) evaluateCase(tc *apipb.TestCase, validatorFlags []string, outPath string) (*apipb.Result, error) {
	// TODO: implement evaluation cache
	res := &apipb.Result{
		Type: apipb.ResultType_TEST_CASE,
	}
	tcPath := filepath.Join(e.root, tc.Name)
	exit, err := e.runSubmission(tcPath, tc.InputPath)
	if err != nil {
		return res, err
	}
	if exit.Crashed() {
		res.Verdict = apipb.Verdict_RUN_TIME_ERROR
	} else if exit.TimedOut() {
		res.Verdict = apipb.Verdict_TIME_LIMIT_EXCEEDED
	} else {
		wa := false
		if e.evalSandbox != nil {
			wa, err = e.runValidator(validatorFlags, tc.InputPath, outPath, tc.OutputPath)
			if err != nil {
				return res, err
			}
		} else {
			diff, err := diffOutput(tc.OutputPath, outPath, validatorFlags)
			if err != nil {
				return res, err
			}
			wa = !diff.Match
		}
		if wa {
			res.Verdict = apipb.Verdict_WRONG_ANSWER
		} else {
			res.Verdict = apipb.Verdict_ACCEPTED
		}
	}
	res.TimeUsageMs = int32(exit.TimeUsageMs)
	e.resultChan <- res
	return res, nil
}

func (e *Evaluator) runSubmission(tcPath, inputPath string) (*ExecResult, error) {
	fb := util.NewFileBase(tcPath)
	fb.OwnerGid = util.OmogenexecGroupId()
	fb.GroupWritable = true
	if err := os.MkdirAll(tcPath, 0755); err != nil {
		return nil, err
	}
	if err := e.linker.LinkFile(inputPath, "input", false); err != nil {
		return nil, err
	}
	if err := fb.WriteFile("output", []byte{}); err != nil {
		return nil, err
	}
	if err := e.linker.LinkFile(tcPath+"/output", "output", true); err != nil {
		return nil, err
	}
	if err := fb.WriteFile("error", []byte{}); err != nil {
		return nil, err
	}
	if err := e.linker.LinkFile(tcPath+"/error", "error", true); err != nil {
		return nil, err
	}
	defer e.resetPermissions()
	return e.programSandbox.Run(e.plan.Program.RunCommand)
}

func (e *Evaluator) runValidator(groupFlags []string, inpath, teampath, anspath string) (bool, error) {
	if err := e.valLinker.LinkFile(inpath, "input", false); err != nil {
		return false, err
	}
	if err := e.valLinker.LinkFile(teampath, "team_output", false); err != nil {
		return false, err
	}
	if err := e.valLinker.LinkFile(anspath, "judge_answer", false); err != nil {
		return false, err
	}
	if err := e.valLinker.LinkFile(anspath, "judge_answer", false); err != nil {
		return false, err
	}

	exit, err := e.evalSandbox.Run(append(e.plan.Validator.RunCommand, append([]string{
		e.valLinker.PathFor("input", false),
		e.valLinker.PathFor("judge_answer", false),
		e.valLinker.PathFor("feedback_dir", true),
	}, groupFlags...)...))
	if err != nil {
		return false, err
	}
	if err := e.resetPermissions(); err != nil {
		return false, err
	}

	if exit.TimedOut() {
		return false, fmt.Errorf("output validator timed out")
	}
	if exit.CrashedWith(42) {
		return false, nil
	}
	if exit.CrashedWith(43) {
		return true, nil
	}
	// Crash was abnormal
	dat, err := ioutil.ReadFile(e.valLinker.PathFor("error", true))
	if err != nil {
		return false, fmt.Errorf("could not read output validator errors: %v", err)
	}
	dat2, err := ioutil.ReadFile(e.valLinker.PathFor("output", true))
	if err != nil {
		return false, fmt.Errorf("could not read output validator output: %v", err)
	}
	return false, fmt.Errorf("output validator Crashed: %v", string(dat)+" "+string(dat2))
}

func diffOutput(refPath, outPath string, args []string) (*DiffResult, error) {
	refFile, err := os.Open(refPath)
	if err != nil {
		return nil, err
	}
	outFile, err := os.Open(outPath)
	if err != nil {
		return nil, err
	}
	diffArgs := DiffArgs{}
	argIdx := 0
	for argIdx < len(args) {
		arg := args[argIdx]
		if arg == "case_sensitive" {
			diffArgs.CaseSensitive = true
		} else if arg == "space_change_sensitive" {
			diffArgs.SpaceSensitive = true
		} else if argIdx+1 < len(args) {
			if arg == "float_tolerance" || arg == "float_relative_tolerance" || arg == "float_absolute_tolerance" {
				tolerance, err := strconv.ParseFloat(args[argIdx+1], 64)
				if err != nil {
					return nil, err
				}
				if arg != "float_absolute_tolerance" {
					diffArgs.RelativePrec = tolerance
				}
				if arg != "float_relative_tolerance" {
					diffArgs.AbsolutePrec = tolerance
				}
			}
		}
	}
	return Diff(refFile, outFile, diffArgs)
}
