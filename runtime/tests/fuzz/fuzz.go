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
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/tests/utils"
)

func runByteSample(data []byte) int {

	if !utf8.Valid(data) {
		return -1
	}

	return runStringSample(strings.TrimSpace(string(data)))
}

func runStringSample(code string) (rc int) {
	reproducer := fmt.Sprintf("runStringSample(%s)", strconv.QuoteToASCII(code))
	SetMessageToDumpOnUnexpectedExit(reproducer)
	defer SetMessageToDumpOnUnexpectedExit("")

	rc = 99999 // if this function returns normally, rc will change
	defer func() {
		if rc == 99999 { // sample returned abnormally, print a reproducer
			fmt.Println("\n\n\t" + reproducer + "\n\n")
		}
	}()

	program, err := parser2.ParseProgram(code)

	if err != nil {
		return 0
	}
	if program == nil || len(program.Declarations()) == 0 {
		return -1
	}

	checker, err := sema.NewChecker(
		program,
		utils.TestLocation,
		sema.WithAccessCheckMode(sema.AccessCheckModeNotSpecifiedUnrestricted),
	)
	if err != nil {
		return 0
	}

	err = checker.Check()
	if err != nil {
		return 0
	}

	var uuid uint64

	inter, err := interpreter.NewInterpreter(
		interpreter.ProgramFromChecker(checker),
		checker.Location,
		interpreter.WithUUIDHandler(func() (uint64, error) {
			defer func() { uuid++ }()
			return uuid, nil
		}),
	)
	if err != nil {
		return 0
	}

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
