package main

import (
	"bufio"
	"fmt"
	"github.com/google/logger"
	"io"
	"math"
	"strings"
)

// DiffResult describes a comparison of two strings.
type DiffResult struct {
	// Whether the strings matched.
	Match bool
	// A textual description of the difference.
	Description string
}

// DiffArgs specifies rules
type DiffArgs struct {
	// Whether tokens representing floating point integers should be parsed as such and compared
	// using precisions RelativePrec and AbsolutePrec rather than as strings.
	ParseFloats bool
	// Indicates that floating-point tokens should be accepted if they are within relative error ≤ ε
	RelativePrec float64
	// Indicates that floating-point tokens should be accepted if they are within absolute error ≤ ε
	AbsolutePrec float64
	// Indicates that comparisons should be case-sensitive
	CaseSensitive bool
	// Indicates that changes in the amount of whitespace should be rejected (the default is that
	// any sequence of 1 or more whitespace characters are equivalent)
	SpaceSensitive bool
}

// Diff compares the a Reader against a "correct" reference Reader by tokenizing them.
func Diff(reference, output io.Reader, args DiffArgs) (*DiffResult, error) {
	ref := newPositionedScanner(bufio.NewReader(reference), args.SpaceSensitive)
	out := newPositionedScanner(bufio.NewReader(output), args.SpaceSensitive)
	for {
		refToken, err := ref.Scan()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		refStr := string(refToken.Token)
		outToken, err := out.Scan()
		if err == io.EOF {
			return &DiffResult{false, fmt.Sprintf("Expected more output (next reference token: %s at %v)", refStr, refToken.Pos)}, nil
		} else if err != nil {
			return nil, err
		}
		outStr := string(outToken.Token)
		match, desc := matchToken(refStr, outStr, args)
		if !match {
			return &DiffResult{false, fmt.Sprintf("%s (output %v, reference %v)", desc, outToken.Pos, refToken.Pos)}, nil
		}
	}
	outToken, err := out.Scan()
	if err != io.EOF {
		if err != nil {
			return nil, err
		}
		return &DiffResult{false, fmt.Sprintf("Too much output (next output token: %s at %v)", string(outToken.Token), outToken.Pos)}, nil
	}
	return &DiffResult{Match: true}, nil
}

func matchToken(ref, out string, args DiffArgs) (bool, string) {
	logger.Infof("Diff %s %s", ref, out)
	if args.ParseFloats {
		var refFloat, outFloat float64
		strVal := ""
		if scanned, _ := fmt.Sscanf(ref, "%f%s", &refFloat, &strVal); scanned == 1 {
			if scanned, _ := fmt.Sscanf(out, "%f%s", &outFloat, &strVal); scanned != 1 {
				return false, fmt.Sprintf("Reference was decimal, output was: %s", out)
			}
			logger.Infof("got decimals %f %f", refFloat, outFloat)
			diff := math.Abs(refFloat - outFloat)
			logger.Infof("float Diff %f", diff)
			if !(diff <= args.AbsolutePrec) &&
				!(diff <= args.RelativePrec*math.Abs(refFloat)) {
				return false, fmt.Sprintf(
					"Too large decimal difference. Reference: %s, output: %s, difference: %f (absolute difference > %f and relative difference > %f)",
					ref, out, diff, args.AbsolutePrec, args.RelativePrec)
			}
			return true, ""
		}
	}
	diff := !strings.EqualFold(ref, out)
	logger.Infof("Diff w/o case? %v", diff)
	if diff {
		return false, fmt.Sprintf("Output was %s, expected %s", out, ref)
	}
	if args.CaseSensitive && ref != out {
		return false, fmt.Sprintf("Output was %s, expected %s (difference in casing)", out, ref)
	}
	return true, ""
}

type Position struct {
	Line int
	Col  int
}

func (p *Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type positionedScanner struct {
	reader         *bufio.Reader
	spaceSensitive bool
	pos            Position
}

func newPositionedScanner(reader *bufio.Reader, spaceSensitive bool) positionedScanner {
	return positionedScanner{
		reader:         reader,
		spaceSensitive: spaceSensitive,
		pos: Position{
			Line: 1,
			Col:  1,
		},
	}
}

type posToken struct {
	Pos   Position
	Token []byte
}

func (sc *positionedScanner) Scan() (*posToken, error) {
	// Fast-forward through spaces
	if !sc.spaceSensitive {
		for {
			nextByte, err := sc.peekByte()
			if err != nil {
				return nil, err
			}
			if isSpace(nextByte) {
				if _, err := sc.eatByte(); err != nil {
					return nil, err
				}
			} else {
				break
			}
		}
	}
	token := &posToken{
		Pos:   sc.pos,
		Token: make([]byte, 0),
	}
	for {
		nextByte, err := sc.peekByte()
		if err == io.EOF {
			if len(token.Token) == 0 {
				return token, io.EOF
			} else {
				return token, nil
			}
		} else if err != nil {
			return nil, err
		}
		if isSpace(nextByte) {
			if !sc.spaceSensitive {
				if len(token.Token) == 0 {
					logger.Fatal("Peeked space after empty token!?")
				}
				return token, nil
			} else {
				// If we're space sensitive, single spaces are tokens, but spaces should never be
				// included with a token if we had non-space chars already
				if len(token.Token) == 0 {
					if _, err := sc.eatByte(); err != nil {
						return nil, err
					}
					token.Token = []byte{nextByte}
				}
				return token, nil
			}
		} else {
			if _, err := sc.eatByte(); err != nil {
				return nil, err
			}
			token.Token = append(token.Token, nextByte)
		}
	}
}

func (sc *positionedScanner) peekByte() (byte, error) {
	peek, err := sc.reader.Peek(1)
	if err != nil {
		return 0, err
	}
	return peek[0], nil
}

func (sc *positionedScanner) eatByte() (byte, error) {
	nextByte, err := sc.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	if nextByte == '\n' {
		sc.pos.Line++
		sc.pos.Col = 1
	} else {
		sc.pos.Col++
	}
	return nextByte, nil
}

func isSpace(b byte) bool {
	return (9 <= b && b <= 13) || b == 32
}
