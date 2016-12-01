/*
Copyright 2015 All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeConfig(t *testing.T) {

}

func TestContainedIn(t *testing.T) {
	assert.False(t, ContainedIn("1", []string{"2", "3", "4"}))
	assert.True(t, ContainedIn("1", []string{"1", "2", "3", "4"}))
}

func TestIsDirectory(t *testing.T) {
	assert.False(t, IsDirectory("no_there"))
	dir, err := ioutil.TempDir("/tmp", "is_directory")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	assert.True(t, IsDirectory(dir))
}

func TestIsFile(t *testing.T) {
	assert.False(t, IsFile("no_there"))
	file, err := ioutil.TempFile("/tmp", "is_file")
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(file.Name())
	assert.True(t, IsFile(file.Name()))
}
