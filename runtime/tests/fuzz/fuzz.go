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
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/cadence/runtime/parser2/lexer"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/tests/utils"
)

var FUZZSTATS = false
var FUZZDUMP = os.Getenv("FUZZDUMP") == "1"
var FUZZTIMEOUT = 0

func init() {
	FUZZSTATS = os.Getenv("FUZZSTATS") == "1"
	if ft, err := strconv.Atoi(os.Getenv("FUZZTIMEOUT")); err == nil && ft != 0 {
		FUZZTIMEOUT = ft
	}
}

func sampleId(data []byte) string {
	x := sha1.New()
	x.Write(data)
	return fmt.Sprintf("%x", x.Sum([]byte{}))
}

func outputId(code string) string {
	if len(code) == 0 {
		return "0"
	}
	return sampleId([]byte(code))
}

func runByteSample(data []byte) int {
	reproducer := fmt.Sprintf("runByteSample(%#v)", data)
	return runStreamSample(sampleId(data), reproducer, lexer.Lex(strings.TrimSpace(string(data))))
}

func runStringSample(code string) (rc int) {
	reproducer := fmt.Sprintf("runStringSample(%s)", strconv.QuoteToASCII(code))
	return runStreamSample(sampleId([]byte(code)), reproducer, lexer.Lex(code))
}

func runRandomTokenStreamSample(chunkSize int, data []byte) int {
	reproducer := fmt.Sprintf("runRandomTokenStreamSample(%d, %#v)", chunkSize, data)
	stream := &lexer.CodeGatheringTokenStream{
		Delegate: &lexer.AutoSpacingTokenStream{
			Delegate: &RandomTokenStream{Fuzzbits: NewFuzzbits(chunkSize, data)}}}
	return runStreamSample(sampleId(data), reproducer, stream)
}

func runSimpleRandomTokenStreamSample(chunkSize int, data []byte) int {
	reproducer := fmt.Sprintf("runSimpleRandomTokenStreamSample(%d, %#v)", chunkSize, data)
	stream := &lexer.CodeGatheringTokenStream{
		Delegate: &lexer.AutoSpacingTokenStream{
			Delegate: &SimpleRandomTokenStream{Fuzzbits: NewFuzzbits(chunkSize, data)}}}
	return runStreamSample(sampleId(data), reproducer, stream)
}

func runStreamSample(sampleId string, reproducer string, stream lexer.TokenStream) (rc int) {
	SetMessageToDumpOnUnexpectedExit(reproducer)
	defer SetMessageToDumpOnUnexpectedExit("")

	code := ""
	currentState := ""
	runId := rand.Int31()

	timeout := newTimeout()
	defer timeout.Cancel()

	state := func(s string) {
		timeout.Cancel()
		if FUZZSTATS {
			currentState = s
			msg := fmt.Sprintf("\nSTAT %s %08x %s %s %s\n", sampleId, runId, outputId(code), currentState, "crashed")
			msg += fmt.Sprintf("CRASH %s %s\n", sampleId, reproducer)
			SetMessageToDumpOnUnexpectedExit(msg)
		}
		timeout.Start(getFuzzTimeout(s), func() {
			if !FUZZSTATS {
				return
			}
			msg := fmt.Sprintf("\nSTAT %s %08x %s %s %s\n", sampleId, runId, outputId(code), currentState, "timeout")
			msg += fmt.Sprintf("TIMEOUT %s %s\n", sampleId, reproducer)
			fmt.Println(msg)
		})
	}

	rc = 99999 // if this function returns normally, rc will change
	defer func() {
		if rc == 99999 { // sample returned abnormally, print a reproducer
			fmt.Println("\n\n\t" + reproducer + "\n\n")
		}
		if FUZZSTATS {
			res := map[int]string{-1: "invalid", 0: "err", 1: "ok", 99999: "panic"}[rc]
			fmt.Printf("\nSTAT %s %08x %s %s %s\n", sampleId, runId, outputId(code), currentState, res)
			if res == "panic" {
				fmt.Printf("PANIC %s %s\n", sampleId, reproducer)
			}
		}
	}()

	state("generating")
	code, generatorPanicked := generate(stream)

	reproducer = fmt.Sprintf("runStringSample(%s) // %s", strconv.QuoteToASCII(code), reproducer)

	state("parsing")
	program, err := parser2.ParseProgram(code)

	if generatorPanicked != nil { // generator panicked, parse did not? a "false positive"
		state("generating")      // i.e. bug is in fuzzer generator, not in parser;
		panic(generatorPanicked) // rethrow original panic.
	}

	if err != nil {
		return 0
	}
	if program == nil || len(program.Declarations()) == 0 {
		return -1
	}

	if FUZZDUMP {
		fmt.Printf("FUZZDUMP %s // %s\n", reproducer, sampleId)
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

func getFuzzTimeout(stage string) int {
	if ms, err := strconv.Atoi(os.Getenv("FUZZTIMEOUT_" + stage)); err == nil {
		return ms
	}
	return FUZZTIMEOUT
}

func generate(stream lexer.TokenStream) (code string, err interface{}) {
	defer func() {
		code = stream.Input() // return the (potentially partially) generated code.
		if err = recover(); err != nil {
			// if there was a generation error, return it; panic later.
			stackbuf := make([]byte, 4096)
			stacklen := runtime.Stack(stackbuf, false)
			err = fmt.Sprintf("%v, %s", err, string(stackbuf[0:stacklen]))
		}
	}()
	_, _ = parser2.ParseProgramFromTokenStream(stream)
	return
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

/////////////////////////////////////////////////////////////////////////////

type Timeout struct {
	id *int32
}

const _TIMEDOUT = -1

func newTimeout() Timeout {
	id_ := int32(0)
	return Timeout{id: &id_}
}

func (t Timeout) Cancel() {
	was := atomic.LoadInt32(t.id)
	if was != _TIMEDOUT {
		atomic.CompareAndSwapInt32(t.id, was, 0)
	}
}

func (t Timeout) Start(ms int, callback func()) {
	if ms == 0 {
		return
	}
	if ms < 0 {
		panic(fmt.Errorf("invalid timeout: %d", ms))
	}
	timerId := rand.Int31()
	if !atomic.CompareAndSwapInt32(t.id, 0, timerId) {
		panic("failed to start timer")
	}
	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		if atomic.CompareAndSwapInt32(t.id, timerId, _TIMEDOUT) {
			callback()
			os.Exit(123)
		}
	}()
}

func (t Timeout) TimedOut() bool {
	return atomic.LoadInt32(t.id) <= _TIMEDOUT
}

func (t Timeout) Dispose() {
	t.Cancel()
}
