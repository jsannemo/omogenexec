package runner

import (
	"strings"
	"testing"
)

type testCase struct {
	reference string
	output    string
	match     bool
}

func TestDiff(t *testing.T) {
	args := DiffArgs{
		ParseFloats:    false,
		CaseSensitive:  false,
		SpaceSensitive: false,
	}
	cases := []testCase{
		{
			reference: "hello world!",
			output:    "Hello World!",
			match:     true,
		},
		{
			reference: "   hello     world!    ",
			output:    "Hello World!",
			match:     true,
		},
		{
			reference: "test",
			output:    " test ",
			match:     true,
		},
		{
			reference: "te st",
			output:    "test",
			match:     false,
		},
		{
			reference: "te\nst",
			output:    "te st",
			match:     true,
		},
		{
			reference: "åÄö",
			output:    "åäÖ",
			match:     true,
		},
		{
			reference: "1.0",
			output:    "1.00",
			match:     false,
		},
	}
	for _, tc := range cases {
		runTest(tc, args, t)
	}
}

func TestDiff_CaseSensitive(t *testing.T) {
	args := DiffArgs{
		ParseFloats:    false,
		CaseSensitive:  true,
		SpaceSensitive: false,
	}
	cases := []testCase{
		{
			reference: "HallÅ!",
			output:    "HallÅ!",
			match:     true,
		},
		{
			reference: "hello world!",
			output:    "Hello World!",
			match:     false,
		},
		{
			reference: "Hallå!",
			output:    "HallÅ!",
			match:     false,
		},
	}
	for _, tc := range cases {
		runTest(tc, args, t)
	}
}

func TestDiff_Floats(t *testing.T) {
	args := DiffArgs{
		ParseFloats:    true,
		RelativePrec:   0.1,
		AbsolutePrec:   1,
		CaseSensitive:  false,
		SpaceSensitive: false,
	}
	cases := []testCase{
		{
			reference: "1.0",
			output:    "1.00",
			match:     true,
		},
		{
			reference: "2.0",
			output:    "1.00",
			match:     true,
		},
		{
			reference: "2.0000000001",
			output:    "1.00",
			match:     false,
		},
		{
			reference: "101",
			output:    "100",
			match:     true,
		},
		{
			reference: "0e-100",
			output:    "1.0",
			match:     true,
		},
		{
			reference: "0x1.fffffffffffffp+1023",
			output:    "1.7976931348623158e+308",
			match:     true,
		},
	}
	for _, tc := range cases {
		runTest(tc, args, t)
	}
}

func TestDiff_SpaceSensitive(t *testing.T) {
	args := DiffArgs{
		ParseFloats:    false,
		CaseSensitive:  false,
		SpaceSensitive: true,
	}
	cases := []testCase{
		{
			reference: "   hello     world!    ",
			output:    "Hello World!",
			match:     false,
		},
		{
			reference: "   hello     world!    ",
			output:    "   Hello     World!    ",
			match:     true,
		},
		{
			reference: " test",
			output:    "test",
			match:     false,
		},
		{
			reference: "test ",
			output:    "test",
			match:     false,
		},
		{
			reference: "test",
			output:    "test ",
			match:     false,
		},
		{
			reference: "test",
			output:    " test",
			match:     false,
		},
		{
			reference: "te st",
			output:    "te  st",
			match:     false,
		},
		{
			reference: "te st",
			output:    "te\nst",
			match:     false,
		},
	}
	for _, tc := range cases {
		runTest(tc, args, t)
	}
}

func runTest(test testCase, args DiffArgs, t *testing.T) {
	diff, err := Diff(strings.NewReader(test.reference), strings.NewReader(test.output), args)
	if err != nil {
		t.Errorf("Got unexpected error from Diff: %v", err)
	}
	if diff.Match != test.match {
		t.Errorf("Expected match to be %v, was %v.\n%s\n%s", test.match, diff.Match, test.reference, test.output)
	}
}
