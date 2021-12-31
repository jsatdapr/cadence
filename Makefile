#
# Cadence - The resource-oriented smart contract programming language
#
# Copyright 2019-2020 Dapper Labs, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

GOPATH ?= $(HOME)/go

# Ensure go bin path is in path (Especially for CI)
PATH := $(PATH):$(GOPATH)/bin

COVERPKGS := $(shell go list ./... | grep -v /cmd | grep -v /runtime/test | tr "\n" "," | sed 's/,*$$//')
GOFUZZBETA := $(shell go help testflag | grep -q fuzz && echo yes)
GOFUZZDVYU := $(shell go-fuzz-build -help 2>/dev/null && echo yes)

.PHONY: test-with-coverage
test-with-coverage: COVERAGE=-coverprofile=coverage.txt -covermode=atomic -coverpkg $(COVERPKGS)
test-with-coverage: test-with-race
test-with-race: RACE=-race
test-with-race: test

J ?= 8


.PHONY: test
test:
	# test all packages
	CGO_ENABLED=$(if $(RACE),1,0) \
	GO111MODULE=on go test -parallel $(J) $(RACE) $(COVERAGE) -test.count=1 ./...
	# remove coverage of empty functions from report
	touch coverage.txt && sed -i -e 's/^.* 0 0$$//' coverage.txt
	cd ./languageserver && make test

.PHONY: build
build:
	go build -o ./runtime/cmd/parse/parse ./runtime/cmd/parse
	GOARCH=wasm GOOS=js go build -o ./runtime/cmd/parse/parse.wasm ./runtime/cmd/parse
	go build -o ./runtime/cmd/check/check ./runtime/cmd/check
	go build -o ./runtime/cmd/main/main ./runtime/cmd/main
	cd ./languageserver && make build

.PHONY: lint-github-actions
lint-github-actions: build-linter
	tools/golangci-lint/golangci-lint run --out-format=github-actions -v ./...

.PHONY: lint
lint-%-buildtag: build-linter
	tools/golangci-lint/golangci-lint run -v --build-tags=$* ./...
lint: lint-default-buildtag
lint: lint-fuzzgen-buildtag
lint: lint-fuzzbuzz-buildtag
lint: $(if $(GOFUZZBETA),lint-gofuzzbeta-buildtag)

.PHONY: fix-lint
fix-lint: build-linter
	tools/golangci-lint/golangci-lint run -v --fix ./...

.PHONY: build-linter
build-linter: tools/golangci-lint/golangci-lint tools/maprangecheck/maprangecheck.so
include tools/maprangecheck/Makefile
include tools/golangci-lint/Makefile

.PHONY: check-headers
check-headers:
	@./check-headers.sh

.PHONY: generate
generate:
	go generate -x ./...

.PHONY: fuzz
fuzz: export FUZZTIMEOUT=2500
# try `make FUZZDUMP=1 J=1 fuzz` to see a stream of fuzz samples
fuzz: $(if $(GOFUZZDVYU),./runtime/tests/fuzz/FuzzRandomBytes-dvyukov)
fuzz: $(if $(GOFUZZBETA),./runtime/tests/fuzz/FuzzRandomBytes-gofuzzbeta)
fuzz: $(if $(GOFUZZBETA),./runtime/tests/fuzz/FuzzRandomStrings-gofuzzbeta)

.PHONY: fuzzstats
fuzzstats: fuzz
fuzzstats: J=1
fuzzstats: FUZZTIME=60s
fuzzstats: export FUZZSTATS=1
fuzzstats: export FUZZTIMEOUT=2500
fuzzstats: export FUZZTIMEOUT_generating=1600
fuzzstats: export FUZZTIMEOUT_parsing=400
fuzzstats: STATFILTER = | awk '/^(PANIC|CRASH|FUZZDUMP)/ { print } \
  /^STAT/ { total++; outcomes[$$4"-"$$5]++ ; sampleids[$$2]++ } END { \
  for (k in outcomes) { printf "%s %2.2f%%\n", k, 100.0*outcomes[k]/(1.0+total) } \
  for (k in sampleids) { if (sampleids[k] > 1) dupes++ } \
  printf "duplicate inputs %2.2f%% (%d out of %d total)\n", 100.0*dupes/(1.0+total), dupes+0, total+0 } \
' | tee $@.tmp && mv $@.tmp $@.fuzzstats

FUZZTIME ?= 5s
FUZZPCKG = github.com/onflow/cadence/$(dir $@)
FUZZFUNC = $(notdir $*)

%-gofuzzbeta:
	RUNUNTIL=$$(($$(date +%s) + $(subst s,,$(FUZZTIME)))); \
	while true ; do TIMELEFT=$$((RUNUNTIL - $$(date +%s))); \
	test $$TIMELEFT -gt 0 || break; \
	go test -run=NONE \
	  -tags=gofuzzbeta \
	  -test.parallel=$(J) \
	  -test.fuzztime $${TIMELEFT}s \
	  -fuzz=$(FUZZFUNC) $(FUZZPCKG) ; done $(STATFILTER)
%-dvyukov.zip:
	go-fuzz-build -o $@ \
	  -func $(FUZZFUNC) $(FUZZPCKG)
	unzip -p $@ metadata | sed -e 's|File":"\([^"]*\)"|\n\1: #\n$@: \1 #\n|g' | grep '$(shell pwd).*#$$' | sort -u \
       > .deps.$(subst /,_,$@).d
-include .deps.*.d
.PRECIOUS: %-dvyukov.zip
%-dvyukov: %-dvyukov.zip
	timeout --signal int --foreground --preserve-status $(FUZZTIME) \
	go-fuzz -testoutput -procs $(J) -bin $< \
	  -func $(FUZZFUNC) $(FUZZPCKG) $(STATFILTER)

.PHONY: check-tidy
check-tidy: generate
	go mod tidy
	cd languageserver; go mod tidy
	git diff --exit-code

.PHONY: release
release:
	@(VERSIONED_FILES="version.go \
	npm-packages/cadence-parser/package.json \
	npm-packages/cadence-docgen/package.json" \
	./bump-version.sh $(bump))
