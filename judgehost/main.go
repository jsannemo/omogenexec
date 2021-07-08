package main

import (
	"context"
	"flag"
	"github.com/google/logger"
	apipb "github.com/jsannemo/omogenexec/api"
	"github.com/jsannemo/omogenexec/judgehost/eval"
	"google.golang.org/grpc"
	"io/ioutil"
	"strconv"
	"sync"
)

var (
	log     = logger.Init("omogenexec-judgehost", true, false, ioutil.Discard)
	address = flag.String("listen_addr", "127.0.0.1:61811", "The Run server address to listen to in the format host:port")
)

type runServer struct {
}

func (s *runServer) GetLanguages(ctx context.Context, _ *apipb.GetLanguagesRequest) (*apipb.GetLanguagesResponse, error) {
	return nil, nil
}

func (s *runServer) Compile(ctx context.Context, _ *apipb.CompileRequest) (*apipb.CompileResponse, error) {
	return nil, nil
}

func (s *runServer) Evaluate(req *apipb.EvaluateRequest, stream apipb.RunService_EvaluateServer) error {
	return nil
}

func newServer() (*runServer, error) {
	s := &runServer{}
	return s, nil
}

// Register registers a new RunService with the given server.
func Register(grpcServer *grpc.Server) error {
	server, err := newServer()
	if err != nil {
		return err
	}
	apipb.RegisterRunServiceServer(grpcServer, server)
	return nil
}

// Following two source files:
// Copyright (c) 2010-2019 Kattis and all respective contributors
// License: https://github.com/Kattis/problemtools/blob/7f8a37902986558cf4a55211c60f1836ee3c2859/LICENSE
const validatorCc = `
#include <utility>
#include <string>
#include <cassert>
#include <cstring>
#include <cmath>
#include "validate.h"

using namespace std;

void check_case() {
    string line;
    /* Get test mode description from judge input file */
    assert(getline(judge_in, line));

    int value = -1;
    if (sscanf(line.c_str(), "fixed %d", &value) != 1) {
        if (sscanf(line.c_str(), "random %d", &value) == 1) {
            srandom(value);
            value = 1 + random() % 1000;
        } else if (sscanf(line.c_str(), "adaptive %d", &value) == 1) {
            srandom(value);
            value = -1;
        } else {
            assert(!"unknown input instructions");
        }
    }
    if (value == -1) {
        judge_message("I'm not committing to a value, will adaptively choose worst one\n");
    } else {
        judge_message("I'm thinking of %d\n", value);
    }

    int sol_lo = 1, sol_hi = 1000;
    int guesses = 0;
    for (int guesses = 0; guesses < 10; ++guesses) {
        int guess;
        if (!(author_out >> guess)) {
            wrong_answer("Guess %d: couldn't read an integer\n", guesses+1);
        }
        if (guess < 1 || guess > 1000) {
            wrong_answer("Guess %d is out of range: %d\n", guesses+1, guess);
        }
        judge_message("Guess %d is %d\n", guesses+1, guess);
        int diff;
        if (value == -1) {
            if (guess == sol_lo && sol_lo == sol_hi) {
                diff = 0;
            } else if (guess-1 - sol_lo > sol_hi - (guess+1)) {
                diff = -1;
            } else if (guess-1 - sol_lo < sol_hi - (guess+1)) {
                diff = 1;
            } else {
                diff = 2*(random() %2) - 1;
            }
        } else {
            diff = value - guess;
        }
        if (!diff) {
            cout << "correct\n";
            cout.flush();
            return;
        } else if (diff < 0) {
            cout << "lower\n";
            cout.flush();
            // Update the maximum possible hidden value.
            sol_hi = min(sol_hi, guess-1);
        } else {
            cout << "higher\n";
            cout.flush();
            // Update the minimum possible hidden value.
            sol_lo = max(sol_lo, guess+1);
        }
    }
    wrong_answer("Didn't get to correct answer in 10 guesses\n");

    return;
}

int main(int argc, char **argv) {
  init_io(argc, argv);

  check_case();

  /* Check for trailing output. */
  string trash;
  if (author_out >> trash) {
      wrong_answer("Trailing output\n");
  }

  /* Yay! */
  accept();
}
`

const validatorH = `
#pragma once

#include <sys/stat.h>
#include <cassert>
#include <cstdarg>
#include <cstdlib>
#include <iostream>
#include <fstream>
#include <sstream>

typedef void (*feedback_function)(const std::string &, ...);

const int EXITCODE_AC = 42;
const int EXITCODE_WA = 43;
const std::string FILENAME_AUTHOR_MESSAGE = "teammessage.txt";
const std::string FILENAME_JUDGE_MESSAGE = "judgemessage.txt";
const std::string FILENAME_JUDGE_ERROR = "judgeerror.txt";
const std::string FILENAME_SCORE = "score.txt";

#define USAGE "%s: judge_in judge_ans feedback_dir < author_out\n"

std::ifstream judge_in, judge_ans;
std::istream author_out(std::cin.rdbuf());

char *feedbackdir = NULL;

void vreport_feedback(const std::string &category,
                      const std::string &msg,
                      va_list pvar) {
    std::ostringstream fname;
    if (feedbackdir)
        fname << feedbackdir << '/';
    fname << category;
    FILE *f = fopen(fname.str().c_str(), "a");
    assert(f);
    vfprintf(f, msg.c_str(), pvar);
    fclose(f);
}

void report_feedback(const std::string &category, const std::string &msg, ...) {
    va_list pvar;
    va_start(pvar, msg);
    vreport_feedback(category, msg, pvar);
}

void author_message(const std::string &msg, ...) {
    va_list pvar;
    va_start(pvar, msg);
    vreport_feedback(FILENAME_AUTHOR_MESSAGE, msg, pvar);
}

void judge_message(const std::string &msg, ...) {
    va_list pvar;
    va_start(pvar, msg);
    vreport_feedback(FILENAME_JUDGE_MESSAGE, msg, pvar);
}

void wrong_answer(const std::string &msg, ...) {
    va_list pvar;
    va_start(pvar, msg);
    vreport_feedback(FILENAME_JUDGE_MESSAGE, msg, pvar);
    exit(EXITCODE_WA);
}

void judge_error(const std::string &msg, ...) {
    va_list pvar;
    va_start(pvar, msg);
    vreport_feedback(FILENAME_JUDGE_ERROR, msg, pvar);
    assert(0);
}

void accept() {
    exit(EXITCODE_AC);
}

void accept_with_score(double scorevalue) {
    report_feedback(FILENAME_SCORE, "%.9le", scorevalue);
    exit(EXITCODE_AC);
}


bool is_directory(const char *path) {
    struct stat entry;
    return stat(path, &entry) == 0 && S_ISDIR(entry.st_mode);
}

void init_io(int argc, char **argv) {
    if(argc < 4) {
        fprintf(stderr, USAGE, argv[0]);
        judge_error("Usage: %s judgein judgeans feedbackdir [opts] < userout", argv[0]);
    }

    // Set up feedbackdir first, as that allows us to produce feedback
    // files for errors in the other parameters.
    if (!is_directory(argv[3])) {
        judge_error("%s: %s is not a directory\n", argv[0], argv[3]);
    }
    feedbackdir = argv[3];

    judge_in.open(argv[1], std::ios_base::in);
    if (judge_in.fail()) {
        judge_error("%s: failed to open %s\n", argv[0], argv[1]);
    }

    judge_ans.open(argv[2], std::ios_base::in);
    if (judge_ans.fail()) {
        judge_error("%s: failed to open %s\n", argv[0], argv[2]);
    }

    author_out.rdbuf(std::cin.rdbuf());
}`

func main() {
	res, err := eval.Compile(&apipb.Program{
		Sources: []*apipb.SourceFile{
			{Path: "hello.cpp", Contents: []byte(`
#include <cstdio>
#include <cstring>

int main(void) {
    int lo = 0, hi = 1023;
    while (true) {
        int m = (lo+hi)/2;
        printf("%d\n", m);
		fflush(stdout);
        char res[1000];
        scanf("%s", res);
        if (!strcmp(res, "correct")) break;
        if (!strcmp(res, "lower")) hi = m-1;
        else lo = m+1;
    }
    return 0;
}
`)},
		},
		Language: apipb.LanguageGroup_CPP,
	}, "/var/lib/omogen/submissions/13123123/compile")
	if err != nil {
		logger.Fatalf("err: %v", err)
	}
	logger.Infof("res: %v", res)

	validator, err := eval.Compile(&apipb.Program{
		Sources: []*apipb.SourceFile{
			{Path: "validate.cc", Contents: []byte(validatorCc)},
			{Path: "validate.h", Contents: []byte(validatorH)},
		},
		Language: apipb.LanguageGroup_CPP,
	}, "/var/lib/omogen/submissions/13123123/validator")
	if err != nil {
		logger.Fatalf("err: %v", err)
	}
	if validator.Program == nil {
		logger.Fatalf("err: %v", validator.CompilerErrors)
	}
	logger.Infof("res: %v", validator)

	ch := make(chan *apipb.Result, 100)
	cases := []*apipb.TestCase{}
	for i := 0; i < 100; i++ {
		tc := apipb.TestCase{
			Name:       strconv.Itoa(i),
			InputPath:  "/var/lib/omogen/problems/helloworld/data/01.in",
			OutputPath: "/var/lib/omogen/problems/helloworld/data/01.ans",
		}
		cases = append(cases, &tc)
	}
	evaluator, err := eval.NewEvaluator("/var/lib/omogen/submissions/13123123", &apipb.EvaluationPlan{
		PlanType: apipb.EvaluationType_INTERACTIVE,
		Program:     res.Program,
		TimeLimitMs: 1000,
		MemLimitKb:  1000 * 1000,
		Validator: validator.Program,
		ValidatorTimeLimitMs: 60 * 1000,
		ValidatorMemLimitKb:  1000 * 1000,
		ScoringValidator:     false,
		RootGroup: &apipb.TestGroup{
			Cases:                cases,
			Groups:               nil,
			Name:                 "",
			AcceptScore:          0,
			RejectScore:          0,
			OutputValidatorFlags: nil,
			BreakOnFail:          false,
			ScoringMode:          0,
			VerdictMode:          apipb.VerdictMode_FIRST_ERROR,
			AcceptIfAnyAccepted:  false,
		},
	}, ch)
	if err != nil {
		logger.Fatalf("eval setup err: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for res := range ch {
			logger.Infof("result: %v", res)
		}
		wg.Done()
	}()
	err = evaluator.Evaluate()
	if err != nil {
		logger.Fatalf("eval err: %v", err)
	}
	wg.Wait()
}
