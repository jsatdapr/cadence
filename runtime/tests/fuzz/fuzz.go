/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fuzz

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/cadence/runtime/parser2/lexer"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/tests/utils"
)

var FUZZSTATS = false

func init() {
	FUZZSTATS = os.Getenv("FUZZSTATS") == "1"
}

func sampleId(data []byte) string {
	x := sha1.New()
	x.Write(data)
	return fmt.Sprintf("%x", x.Sum([]byte{}))
}

func runByteSample(data []byte) int {
	reproducer := fmt.Sprintf("runByteSample(%#v)", data)
	return runStreamSample(sampleId(data), reproducer, lexer.Lex(strings.TrimSpace(string(data))))
}

func runStringSample(code string) (rc int) {
	reproducer := fmt.Sprintf("runStringSample(%s)", strconv.QuoteToASCII(code))
	return runStreamSample(sampleId([]byte(code)), reproducer, lexer.Lex(code))
}

func runStreamSample(sampleId string, reproducer string, stream lexer.TokenStream) (rc int) {
	SetMessageToDumpOnUnexpectedExit(reproducer)
	defer SetMessageToDumpOnUnexpectedExit("")

	currentState := ""
	runId := rand.Int31()
	state := func(s string) {
		if FUZZSTATS {
			currentState = s
			msg := fmt.Sprintf("\nSTAT %s %08x %s %s\n", sampleId, runId, currentState, "crashed")
			msg += fmt.Sprintf("CRASH %s %s\n", sampleId, reproducer)
			SetMessageToDumpOnUnexpectedExit(msg)
		}
	}

	rc = 99999 // if this function returns normally, rc will change
	defer func() {
		if rc == 99999 { // sample returned abnormally, print a reproducer
			fmt.Println("\n\n\t" + reproducer + "\n\n")
		}
		if FUZZSTATS {
			res := map[int]string{-1: "invalid", 0: "err", 1: "ok", 99999: "panic"}[rc]
			fmt.Printf("\nSTAT %s %08x %s %s\n", sampleId, runId, currentState, res)
			if res == "panic" {
				fmt.Printf("PANIC %s %s\n", sampleId, reproducer)
			}
		}
	}()

	state("parsing")
	program, err := parser2.ParseProgramFromTokenStream(stream)

	if err != nil {
		return 0
	}
	if program == nil || len(program.Declarations()) == 0 {
		return -1
	}

	state("newchecker")
	checker, err := sema.NewChecker(
		program,
		utils.TestLocation,
		sema.WithAccessCheckMode(sema.AccessCheckModeNotSpecifiedUnrestricted),
	)
	if err != nil {
		return 0
	}

	state("checking")
	err = checker.Check()
	if err != nil {
		return 0
	}

	var uuid uint64

	state("newinterpreter")
	inter, err := interpreter.NewInterpreter(
		interpreter.ProgramFromChecker(checker),
		checker.Location,
		interpreter.WithUUIDHandler(func() (uint64, error) {
			defer func() { uuid++ }()
			return uuid, nil
		}),
		interpreter.WithStorage(interpreter.NewInMemoryStorage()),
	)
	if err != nil {
		return 0
	}

	state("interpret")
	err = inter.Interpret()
	if err != nil {
		return 0
	}

	return 1
}

var MessageToDumpOnUnexpectedExit = []byte("Unexpected os.Exit()\n")
var MessageToDumpOnUnexpectedExit_len = len(MessageToDumpOnUnexpectedExit)

func SetMessageToDumpOnUnexpectedExit(msg string) {
	if msg == "" {
		MessageToDumpOnUnexpectedExit_len = 0
		return
	}
	s := `
Unexpected os.Exit()
v----------------------v
` + msg + `
^----------------------^
`
	MessageToDumpOnUnexpectedExit = []byte(s)
	MessageToDumpOnUnexpectedExit_len = len(MessageToDumpOnUnexpectedExit)
}
