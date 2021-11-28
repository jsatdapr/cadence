//go:build linux && amd64 && !race
// +build linux,amd64,!race

/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2021 Dapper Labs, Inc.
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
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

func init() {
	replaceSignalHandler(unix.SIGABRT, reflect.ValueOf(OnUnexpectedExit).Pointer())

	buf := unix.Utsname{}
	err := unix.Uname(&buf)
	if err != nil || strings.Contains(string(buf.Version[:]), "Microsoft") {
		return // don't bother trying this on WSL1
	}
	replaceSignalHandler(unix.SIGSYS, reflect.ValueOf(OnUnexpectedExit).Pointer())

	installBpfExitTrap()
}

//go:nosplit
func OnUnexpectedExit() {
	if MessageToDumpOnUnexpectedExit_len != 0 {
		_, _, _ = unix.RawSyscall6(unix.SYS_WRITE, 2,
			uintptr(unsafe.Pointer(&MessageToDumpOnUnexpectedExit[0])),
			uintptr(MessageToDumpOnUnexpectedExit_len), 0, 0, 0,
		)
	}
	_, _, _ = unix.RawSyscall6(unix.SYS_EXIT_GROUP, 123, 0, 0, 0, 0, 0)
	*(*int)(nil) = 0 // unreachable
}

//go:nosplit
func replaceSignalHandler(sig unix.Signal, newHandler uintptr) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var sa [4]uintptr // get the old settings
	_, _, e := unix.RawSyscall6(unix.SYS_RT_SIGACTION, uintptr(sig),
		0, uintptr(unsafe.Pointer(&sa[0])), 8, 0, 0)
	if e != 0 {
		panic(e)
	}

	sa[0] = newHandler // update with new setting
	_, _, e = unix.RawSyscall6(unix.SYS_RT_SIGACTION, uintptr(sig),
		uintptr(unsafe.Pointer(&sa[0])), 0, 8, 0, 0)
	if e != 0 {
		panic(e)
	}
}

/* this comment is a workaround for https://github.com/dvyukov/go-fuzz/issues/301 */

//go:nosplit
func installBpfExitTrap() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// before adding seccomp filter, have to "SET_NO_NEW_PRIVS"
	// https://www.kernel.org/doc/html/v5.13/userspace-api/no_new_privs.html
	_, _, errno := unix.RawSyscall6(unix.SYS_PRCTL, unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0, 0)
	if errno != 0 {
		panic(errno)
	}

	const (
		// https://github.com/torvalds/linux/blob/v5.13/include/uapi/linux/audit.h#L435
		AUDIT_ARCH_X86_64 = 0xc000003e

		// https://github.com/torvalds/linux/blob/v5.13/include/uapi/linux/seccomp.h#L14
		SECCOMP_SET_MODE_FILTER   = 1
		SECCOMP_FILTER_FLAG_TSYNC = 1
		SECCOMP_RET_KILL_THREAD   = 0x00000000
		SECCOMP_RET_TRAP          = 0x00030000
		SECCOMP_RET_ALLOW         = 0x7fff0000

		// https://github.com/torvalds/linux/blob/v5.13/include/uapi/linux/bpf_common.h#L7
		LDABSW = 0x20
		RETURN = 0x06
		JMPKEQ = 0x15

		// https://github.com/torvalds/linux/blob/v5.13/include/uapi/linux/seccomp.h#L61
		offsetof_seccomp_data_nr   = 0
		offsetof_seccomp_data_arch = 4
		offsetof_seccomp_data_arg0 = 16
	)

	// https://github.com/torvalds/linux/blob/v5.13/include/uapi/linux/filter.h#L24
	type sock_filter struct {
		code uint16
		jt   uint8
		jf   uint8
		k    uint32
	}
	type sock_fprog struct {
		len uint16
		_   [6]byte
		ops *sock_filter
	}

	if false { // satisfy golint-ci: it says that code,jt,jf,k are unused.
		println(sock_filter{code: 0, jt: 1, jf: 2, k: 3}) // fake them used
	}

	filter := []sock_filter{
		// if arch != expected arch; kill thread.
		{code: LDABSW, k: offsetof_seccomp_data_arch},
		{code: JMPKEQ, k: AUDIT_ARCH_X86_64, jt: 1, jf: 0},
		{code: RETURN, k: SECCOMP_RET_KILL_THREAD},
		// if not SYS_EXIT_GROUP, allow
		{code: LDABSW, k: offsetof_seccomp_data_nr},
		{code: JMPKEQ, k: unix.SYS_EXIT_GROUP, jt: 1, jf: 0},
		{code: RETURN, k: SECCOMP_RET_ALLOW},
		// if Exit(0), allow!
		{code: LDABSW, k: offsetof_seccomp_data_arg0},
		{code: JMPKEQ, k: 0, jt: 0, jf: 1},
		{code: RETURN, k: SECCOMP_RET_ALLOW},
		// if Exit(123), allow!
		{code: JMPKEQ, k: 123, jt: 0, jf: 1},
		{code: RETURN, k: SECCOMP_RET_ALLOW},
		// otherwise: unexpected os.Exit; raise SIGSYS
		{code: RETURN, k: SECCOMP_RET_TRAP},
	}

	tid, _, errno := unix.RawSyscall(unix.SYS_SECCOMP,
		SECCOMP_SET_MODE_FILTER,
		SECCOMP_FILTER_FLAG_TSYNC,
		uintptr(unsafe.Pointer(&sock_fprog{
			len: uint16(len(filter)),
			ops: &filter[0],
		})),
	)

	if errno != 0 {
		panic(errno)
	} else if tid != 0 {
		panic(fmt.Errorf("couldn't synchronize filter to TID %d, errno %s", tid, errno.Error()))
	}
}
