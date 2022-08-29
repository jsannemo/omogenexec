package eval

import (
	"fmt"
	"github.com/google/logger"
	apipb "github.com/jsannemo/omogenexec/api"
	"github.com/jsannemo/omogenexec/util"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	exitCodeAc = 42
	exitCodeWa = 43
)

func (e *Evaluator) GetResultForGroup(tcRes *apipb.Result, tg *apipb.TestGroup) *apipb.Result {
	updatedResult := *tcRes
	if !e.plan.ScoringValidator {
		if tcRes.Verdict == apipb.Verdict_ACCEPTED {
			updatedResult.Score = tg.AcceptScore
		} else {
			updatedResult.Score = tg.RejectScore
		}
	}
	return &updatedResult
}

type Evaluator struct {
	root                     string
	linker                   *fileLinker
	valLinker                *fileLinker
	graderLinker             *fileLinker
	plan                     *apipb.EvaluationPlan
	evalCache                map[string]*apipb.Result
	programSandbox           *sandboxWrapper
	evalSandbox              *sandboxWrapper
	graderSandbox            *sandboxWrapper
	resultChan               chan<- *apipb.Result
	validatorCommandTemplate []string
	graderCommandTemplate    []string
}

func NewEvaluator(root string, plan *apipb.EvaluationPlan, results chan<- *apipb.Result) (*Evaluator, error) {
	eval := &Evaluator{
		root:       root,
		plan:       plan,
		evalCache:  make(map[string]*apipb.Result),
		resultChan: results,
	}
	if err := eval.initProgram(); err != nil {
		return nil, fmt.Errorf("failed initializing program: %v", err)
	}
	if err := eval.initValidator(); err != nil {
		return nil, fmt.Errorf("failed initializing validator: %v", err)
	}
	if err := eval.initGrader(); err != nil {
		return nil, fmt.Errorf("failed initializing grader: %v", err)
	}
	return eval, nil
}

func (e *Evaluator) initProgram() error {
	fl, err := newFileLinker(filepath.Join(e.root, "env"))
	if err != nil {
		return fmt.Errorf("failed creating fileLinker: %v", err)
	}
	e.linker = fl
	args := sandboxArgs{
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
	}
	e.programSandbox = newSandbox(0, args)
	return nil
}

func (e *Evaluator) initValidator() error {
	if e.plan.Validator == nil {
		return nil
	}
	valfl, err := newFileLinker(filepath.Join(e.root, "valenv"))
	if err != nil {
		return fmt.Errorf("failed creating validator fileLinker: %v", err)
	}
	e.valLinker = valfl
	args := sandboxArgs{
		WorkingDirectory: e.plan.Validator.ProgramRoot,
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
		TimeLimitMs:   int(e.plan.ValidatorTimeLimitMs),
		MemoryLimitKb: int(e.plan.ValidatorMemLimitKb),
	}
	if e.plan.PlanType == apipb.EvaluationType_INTERACTIVE {
		args.OutputPath = e.linker.PathFor("input", false)
		args.InputPath = e.linker.PathFor("output", true)
	}
	e.evalSandbox = newSandbox(1, args)
	e.validatorCommandTemplate = append(e.validatorCommandTemplate, e.plan.Validator.RunCommand...)
	e.validatorCommandTemplate = append(e.validatorCommandTemplate,
		e.valLinker.PathFor("input", false),
		e.valLinker.PathFor("judge_answer", false),
		e.valLinker.PathFor(".", true),
	)
	return nil
}

func (e *Evaluator) graderCommand(groupFlags []string) []string {
	var flags []string
	flags = append(flags, e.graderCommandTemplate...)
	flags = append(flags, groupFlags...)
	return flags
}

func (e *Evaluator) initGrader() error {
	if e.plan.Grader == nil {
		return nil
	}
	graderfl, err := newFileLinker(filepath.Join(e.root, "graderenv"))
	if err != nil {
		return fmt.Errorf("failed creating grader fileLinker: %v", err)
	}
	e.graderLinker = graderfl
	args := sandboxArgs{
		WorkingDirectory: e.plan.Grader.ProgramRoot,
		InputPath:        e.graderLinker.PathFor("input", false),
		OutputPath:       e.graderLinker.PathFor("output", true),
		ErrorPath:        e.graderLinker.PathFor("error", true),
		ExtraReadPaths: []string{
			e.graderLinker.readBase.Path(),
			e.plan.Grader.ProgramRoot,
		},
		ExtraWritePaths: []string{
			e.graderLinker.writeBase.Path(),
		},
		TimeLimitMs:   int((60 * time.Second).Milliseconds()),
		MemoryLimitKb: 1000 * 1000, // 1000 MB = 1 GB
	}
	e.graderSandbox = newSandbox(2, args)
	e.graderCommandTemplate = e.plan.Grader.GetRunCommand()
	return nil
}

func (e *Evaluator) resetPermissions() error {
	cmd := exec.Command("/usr/bin/omogenexec-fixpermissions", "--path", filepath.Dir(e.root))
	return cmd.Run()
}

func (e *Evaluator) Evaluate() error {
	defer close(e.resultChan)
	logger.Infof("Starting evaluation in %s", e.root)
	if err := e.resetPermissions(); err != nil {
		return fmt.Errorf("could not reset permissions: %v", err)
	}
	defer e.resetPermissions()
	if err := e.programSandbox.Start(); err != nil {
		return fmt.Errorf("failed starting sandbox: %v", err)
	}
	defer e.programSandbox.Finish()
	if e.plan.Validator != nil {
		if err := e.evalSandbox.Start(); err != nil {
			return fmt.Errorf("failed starting sandbox: %v", err)
		}
		defer e.evalSandbox.Finish()
	}
	if e.plan.Grader != nil {
		if err := e.graderSandbox.Start(); err != nil {
			return fmt.Errorf("failed starting sandbox: %v", err)
		}
		defer e.graderSandbox.Finish()
	}
	_, err := e.evaluateGroup(e.plan.RootGroup)
	logger.Infof("Completed evaluation of %s", e.root)
	return err
}

type evalable struct {
	*apipb.TestGroup
	*apipb.TestCase
}

// evalableLess compares two evalable according to the order in which they should be evaluated.
func evalableLess(a *evalable, b *evalable) bool {
	var aName, bName string
	if a.TestGroup != nil {
		aName = a.TestGroup.Name
	} else {
		aName = a.TestCase.Name
	}
	if b.TestGroup != nil {
		bName = b.TestGroup.Name
	} else {
		bName = b.TestCase.Name
	}
	return aName < bName
}

// worseness orders, increasingly, verdicts by how they should be returned by the worst error verdict mode.
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
	default:
		panic(fmt.Sprintf("unknown verdict %v", v))
	}
}

func verdictToAbbreviation(v apipb.Verdict) string {
	switch v {
	case apipb.Verdict_ACCEPTED:
		return "AC"
	case apipb.Verdict_RUN_TIME_ERROR:
		return "RTE"
	case apipb.Verdict_TIME_LIMIT_EXCEEDED:
		return "TLE"
	case apipb.Verdict_WRONG_ANSWER:
		return "WA"
	default:
		panic(fmt.Sprintf("unknown verdict %v", v))
	}
}

func abbreviationToVerdict(v string) (apipb.Verdict, error) {
	switch v {
	case "AC":
		return apipb.Verdict_ACCEPTED, nil
	case "RTE":
		return apipb.Verdict_RUN_TIME_ERROR, nil
	case "TLE":
		return apipb.Verdict_TIME_LIMIT_EXCEEDED, nil
	case "WA":
		return apipb.Verdict_WRONG_ANSWER, nil
	default:
		return apipb.Verdict_VERDICT_UNSPECIFIED, fmt.Errorf("unknown abbreviation: %s", v)
	}
}

// mergeRes aggregates a set of subresults in a testgroup according to its aggregation rules.
func (e *Evaluator) mergeRes(results []*apipb.Result, tg *apipb.TestGroup) (*apipb.Result, error) {
	if tg.CustomGrading {
		result := &apipb.Result{
			Type:        apipb.ResultType_TEST_GROUP,
			TimeUsageMs: 0,
		}
		var graderInputs []byte
		for _, res := range results {
			graderInputs = append(graderInputs, []byte(fmt.Sprintf("%s %f\n", verdictToAbbreviation(res.Verdict), res.Score))...)
			if res.TimeUsageMs > result.TimeUsageMs {
				result.TimeUsageMs = res.TimeUsageMs
			}
		}
		if err := e.graderLinker.readBase.WriteFile("input", graderInputs); err != nil {
			return nil, err
		}
		run, err := e.graderSandbox.Run(e.graderCommand(tg.GraderFlags))
		if err != nil {
			return nil, fmt.Errorf("failed running grader: %v", err)
		}
		if run.TimedOut() {
			return nil, fmt.Errorf("custom grader timed out")
		}
		if run.Crashed() {
			return nil, fmt.Errorf("custom grader crashed")
		}
		dat, err := e.graderLinker.writeBase.ReadFile("output")
		if err != nil {
			return nil, fmt.Errorf("failed reading grader output: %v", err)
		}
		parts := strings.Split(string(dat), " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid grader output: %v", parts)
		}
		parts[1] = strings.TrimSpace(parts[1])
		score, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid grader output: %v", parts)
		}
		result.Score = score
		verdict, err := abbreviationToVerdict(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid grader output: %v", parts)
		}
		result.Verdict = verdict
		return result, nil
	} else {
		return defaultGrader(results, tg), nil
	}
}

func defaultGrader(results []*apipb.Result, tg *apipb.TestGroup) *apipb.Result {
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
	for _, res := range results {
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
		if res.TimeUsageMs > result.TimeUsageMs {
			result.TimeUsageMs = res.TimeUsageMs
		}
	}

	if tg.VerdictMode == apipb.VerdictMode_ALWAYS_ACCEPT || (anyAccepted && tg.AcceptIfAnyAccepted) {
		result.Verdict = apipb.Verdict_ACCEPTED
	}
	return result
}

func (e *Evaluator) evaluateGroup(tg *apipb.TestGroup) (*apipb.Result, error) {
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
			subres, err = e.evaluateGroup(group)
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			tc := eval.TestCase
			cacheKey := tc.InputPath + " " + tc.OutputPath + strings.Join(tg.OutputValidatorFlags, " ")
			if cached, found := e.evalCache[cacheKey]; found {
				subres = e.GetResultForGroup(cached, tg)
			} else {
				subres, err = e.evaluateCase(tc, tg)
				if err != nil {
					return nil, fmt.Errorf("failed on case %s: %v", eval.TestCase.Name, err)
				}
				e.evalCache[cacheKey] = subres
			}
		}
		res = append(res, subres)
		if subres.Verdict != apipb.Verdict_ACCEPTED && tg.BreakOnFail {
			break
		}
	}
	// By happy coincidence, sample < secret in sort order, so sample is always the first result for the root group.
	if tg.IgnoreSample {
		res = res[1:]
	}
	groupRes, err := e.mergeRes(res, tg)
	if err != nil {
		return nil, err
	}
	e.resultChan <- groupRes
	return groupRes, nil
}

func (e *Evaluator) evaluateInteractive(tc *apipb.TestCase, tg *apipb.TestGroup) (*apipb.Result, error) {
	programInput := e.linker.PathFor("input", false)
	programOutput := e.linker.PathFor("output", true)
	if err := e.valLinker.LinkFile(tc.InputPath, "input", false); err != nil {
		return nil, err
	}
	if err := e.valLinker.LinkFile(tc.OutputPath, "judge_answer", false); err != nil {
		return nil, err
	}

	if err := syscall.Mkfifo(programInput, 0660); err != nil {
		return nil, fmt.Errorf("failed making interactive pipe: %v", err)
	}
	if err := syscall.Mkfifo(programOutput, 0660); err != nil {
		return nil, fmt.Errorf("failed making interactive pipe: %v", err)
	}

	e.linker.readBase.GroupWritable = true
	if err := e.linker.readBase.FixMode("input"); err != nil {
		return nil, fmt.Errorf("failed fixing mode for interactive input: %v", err)
	}
	if err := e.linker.readBase.FixOwners("input"); err != nil {
		return nil, fmt.Errorf("failed fixing owners for interactive input: %v", err)
	}
	e.linker.readBase.GroupWritable = false

	if err := e.linker.writeBase.FixMode("output"); err != nil {
		return nil, fmt.Errorf("failed fixing mode for interactive output: %v", err)
	}
	if err := e.linker.writeBase.FixOwners("output"); err != nil {
		return nil, fmt.Errorf("failed fixing owners for interactive output: %v", err)

	}
	inWrite, err := os.OpenFile(programInput, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeNamedPipe)
	inRead, err := os.OpenFile(programInput, os.O_CREATE, os.ModeNamedPipe)
	outWrite, err := os.OpenFile(programOutput, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeNamedPipe)
	outRead, err := os.OpenFile(programOutput, os.O_CREATE, os.ModeNamedPipe)

	var programRun, validatorRun *execResult
	var programErr, validatorErr error
	var validatorFirst bool
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		programRun, programErr = e.programSandbox.Run(e.plan.Program.RunCommand)
		outWrite.Close()
		wg.Done()
	}()
	go func() {
		validatorRun, validatorErr = e.evalSandbox.Run(e.validatorCommand(tg.OutputValidatorFlags))
		inWrite.Close()
		if programRun == nil {
			validatorFirst = true
			if !validatorRun.CrashedWith(exitCodeAc) {
				outRead.Close()
			}
		}
		wg.Done()
	}()
	wg.Wait()
	inRead.Close()
	outRead.Close()
	if programErr != nil {
		return nil, fmt.Errorf("program run failed: %v", programErr)
	}
	if validatorErr != nil {
		return nil, fmt.Errorf("validator run failed: %v", validatorErr)
	}

	val, err := e.validatorOutputFromExit(validatorRun)
	if err != nil {
		return nil, err
	}

	if err := e.linker.Clear(); err != nil {
		return nil, fmt.Errorf("failed clearing program environment: %v", err)
	}
	if err := e.linker.Clear(); err != nil {
		return nil, fmt.Errorf("failed clearing program environment: %v", err)
	}
	if err := e.valLinker.Clear(); err != nil {
		return nil, fmt.Errorf("failed clearing validator environment: %v", err)
	}

	res := &apipb.Result{
		Score:       tg.RejectScore,
		TimeUsageMs: programRun.TimeUsageMs,
	}
	if programRun.TimedOut() {
		res.Verdict = apipb.Verdict_TIME_LIMIT_EXCEEDED
	} else if programRun.Crashed() && programRun.Signal != int(syscall.SIGPIPE) && (!validatorFirst || val.Accepted) {
		res.Verdict = apipb.Verdict_RUN_TIME_ERROR
	} else {
		res.Message = val.JudgeMessage

		if e.plan.ScoringValidator && val.HasScore {
			res.Score = val.Score
		} else if val.Accepted {
			res.Score = tg.AcceptScore
		} else {
			res.Score = tg.RejectScore
		}

		if val.Accepted {
			res.Verdict = apipb.Verdict_ACCEPTED
		} else {
			res.Verdict = apipb.Verdict_WRONG_ANSWER
		}
	}
	return res, nil
}

func (e *Evaluator) evaluateCase(tc *apipb.TestCase, tg *apipb.TestGroup) (*apipb.Result, error) {
	if e.plan.PlanType == apipb.EvaluationType_INTERACTIVE {
		return e.evaluateInteractive(tc, tg)
	}
	outPath := e.linker.PathFor("output", true)
	res := &apipb.Result{
		Type: apipb.ResultType_TEST_CASE,
	}
	tcPath := filepath.Join(e.root, fmt.Sprintf("case-%s", tc.Name))
	exit, err := e.runSubmission(tcPath, tc.InputPath)
	if err != nil {
		return res, fmt.Errorf("sandbox fail: %v, logs %v", err, e.programSandbox.logs())
	}
	if exit.Crashed() {
		res.Verdict = apipb.Verdict_RUN_TIME_ERROR
	} else if exit.TimedOut() {
		res.Verdict = apipb.Verdict_TIME_LIMIT_EXCEEDED
	} else {
		ac := false
		if e.evalSandbox != nil {
			valOutput, err := e.runValidator(tg.OutputValidatorFlags, tc.InputPath, outPath, tc.OutputPath)
			if err != nil {
				return res, fmt.Errorf("failed validator run: %v", err)
			}
			ac = valOutput.Accepted
			res.Message = valOutput.JudgeMessage
			if e.plan.ScoringValidator && valOutput.HasScore {
				res.Score = valOutput.Score
			} else if ac {
				res.Score = tg.AcceptScore
			} else {
				res.Score = tg.RejectScore
			}
		} else {
			diff, err := diffOutput(tc.OutputPath, outPath, tg.OutputValidatorFlags)
			if err != nil {
				return res, fmt.Errorf("default validator failed: %v", err)
			}
			ac = diff.Match
			res.Message = diff.Description
			if ac {
				res.Score = tg.AcceptScore
			} else {
				res.Score = tg.RejectScore
			}
		}

		if ac {
			res.Verdict = apipb.Verdict_ACCEPTED
		} else {
			res.Verdict = apipb.Verdict_WRONG_ANSWER
		}
	}
	res.TimeUsageMs = exit.TimeUsageMs
	if err := e.linker.Clear(); err != nil {
		return nil, fmt.Errorf("failed clearing program env: %v", err)
	}
	if e.valLinker != nil {
		if err := e.valLinker.Clear(); err != nil {
			return nil, fmt.Errorf("failed clearing validator env: %v", err)
		}
	}
	e.resultChan <- res
	logger.Infof("finished test case %s: %v", tc.Name, res)
	return res, nil
}

func (e *Evaluator) runSubmission(tcPath, inputPath string) (*execResult, error) {
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
	return e.programSandbox.Run(e.plan.Program.RunCommand)
}

type ValidatorOutput struct {
	Accepted     bool
	HasScore     bool
	Score        float64
	JudgeMessage string
}

const (
	judgeMessageFile = "judgemessage.txt"
	scoreFile        = "score.txt"
)

func (e *Evaluator) runValidator(groupFlags []string, inpath, teampath, anspath string) (*ValidatorOutput, error) {
	if err := e.valLinker.LinkFile(inpath, "input", false); err != nil {
		return nil, err
	}
	if err := e.valLinker.LinkFile(teampath, "team_output", false); err != nil {
		return nil, err
	}
	if err := e.valLinker.LinkFile(anspath, "judge_answer", false); err != nil {
		return nil, err
	}

	exit, err := e.evalSandbox.Run(e.validatorCommand(groupFlags))
	if err != nil {
		return nil, err
	}
	return e.validatorOutputFromExit(exit)
}

func (e *Evaluator) validatorOutputFromExit(exit *execResult) (*ValidatorOutput, error) {
	output := &ValidatorOutput{}
	if exit.TimedOut() {
		return nil, fmt.Errorf("output validator timed out")
	}
	if exit.CrashedWith(exitCodeAc) {
		output.Accepted = true
	} else if exit.CrashedWith(exitCodeWa) {
		output.Accepted = false
	} else {
		// Crash was abnormal
		dat, err := e.valLinker.writeBase.ReadFile("error")
		if err != nil {
			return nil, fmt.Errorf("could not read crashed output validator errors: %v", err)
		}
		dat2, err := e.valLinker.writeBase.ReadFile("output")
		if err != nil {
			return nil, fmt.Errorf("could not read crashed output validator output: %v", err)
		}
		return nil, fmt.Errorf("output validator crashed (err: %s, output: %s)", string(dat), string(dat2))
	}
	judgeMessage, err := e.valLinker.writeBase.ReadFile(judgeMessageFile)
	if err == nil {
		output.JudgeMessage = string(judgeMessage)
		logger.Infof("output validator message: %s", output.JudgeMessage)
	}
	if e.plan.ScoringValidator && output.Accepted {
		scoreBytes, err := e.valLinker.writeBase.ReadFile(scoreFile)
		scoreStr := strings.TrimSpace(string(scoreBytes))
		if err != nil {
			output.HasScore = false
			logger.Infof("no score.txt from AC output validator?")
		} else {
			score, err := strconv.ParseFloat(scoreStr, 64)
			if err != nil {
				return nil, fmt.Errorf("could not parse score %s from scoring validator %v", scoreStr, err)
			}
			output.Score = score
			output.HasScore = true
		}
	}
	return output, nil
}

func (e *Evaluator) validatorCommand(groupFlags []string) []string {
	var flags []string
	flags = append(flags, e.validatorCommandTemplate...)
	flags = append(flags, groupFlags...)
	return flags
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
			argIdx += 1
		} else if arg == "space_change_sensitive" {
			diffArgs.SpaceSensitive = true
			argIdx += 1
		} else if argIdx+1 < len(args) && (arg == "float_tolerance" || arg == "float_relative_tolerance" || arg == "float_absolute_tolerance") {
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
			diffArgs.ParseFloats = true
			argIdx += 2
		} else {
			logger.Warningf("unknown default output validator flags: %v", args)
		}
	}
	return Diff(refFile, outFile, diffArgs)
}
