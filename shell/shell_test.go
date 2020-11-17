// Copyright 2020 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package shell

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMenderShellExecShell(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Errorf("cant get current user: %s", err.Error())
		return
	}
	uid, err := strconv.ParseUint(currentUser.Uid, 10, 32)
	if err != nil {
		t.Errorf("cant get current uid: %s", err.Error())
		return
	}

	gid, err := strconv.ParseUint(currentUser.Gid, 10, 32)
	if err != nil {
		t.Errorf("cant get current gid: %s", err.Error())
		return
	}

	//command does not exist
	pid, pseudoTTY, cmd, err := ExecuteShell(uint32(uid), uint32(gid), "thatissomethingthatdoesnotexecute", "xterm-256color", 24, 80)
	assert.Error(t, err)
	assert.Equal(t, pid, -1)
	assert.Nil(t, pseudoTTY)
	assert.Nil(t, cmd)

	pid, pseudoTTY, cmd, err = ExecuteShell(uint32(uid), uint32(gid), "/bin/sh", "xterm-256color", 24, 80)
	assert.Nil(t, err)
	assert.NotZero(t, pid)
	assert.NotNil(t, pseudoTTY)

	t.Logf("started shell, pid: %d", pid)

	p, err := os.FindProcess(pid)
	t.Logf("FindProcess p: %v err: %v", p, err)
	assert.Nil(t, err)
	assert.NotNil(t, p)
	p.Signal(syscall.SIGHUP)
	time.Sleep(time.Second)
	pseudoTTY.Close()
	p.Signal(syscall.SIGTERM)
	time.Sleep(time.Second)
	err = p.Signal(syscall.SIGKILL)
	time.Sleep(time.Second)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case err := <-done:
		if err != nil {
		}
	}
	time.Sleep(time.Second)

	p, err = os.FindProcess(pid)
	assert.NotNil(t, p)
	err = p.Signal(syscall.Signal(0))
	if err == nil {
		t.Logf("process is still running after kill -9")
	}
}