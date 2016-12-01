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
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// DecodeFile decodes the file
func DecodeFile(path string, data interface{}) error {
	// step: read in the file contents
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	format := strings.TrimPrefix(filepath.Ext(path), ".")

	return DecodeConfig(bytes.NewReader(content), format, data)
}

// EncodeConfig marshals the item
func EncodeConfig(data interface{}, format string) ([]byte, error) {
	var content []byte
	var err error

	switch format {
	case "json":
		content, err = json.Marshal(data)
	case "yml":
		fallthrough
	case "yaml":
		content, err = yaml.Marshal(data)
	default:
		return []byte(""), fmt.Errorf("unsupported file format: %s", format)
	}
	if err != nil {
		return []byte(""), err
	}

	return content, nil
}

// DecodeConfig unmarshal's the configuration file
func DecodeConfig(reader io.Reader, format string, data interface{}) error {
	// step: read in the content
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		err = json.Unmarshal(content, data)
	case "yml":
		fallthrough
	case "yaml":
		err = yaml.Unmarshal(content, data)
	default:
		return fmt.Errorf("unsupported file format: %s", format)
	}

	if err != nil {
		return err
	}

	return nil
}

func FindFilesInDirectory(paths []string, glob string) ([]string, error) {
	var list []string

	for _, d := range paths {
		if !IsDirectory(d) {
			return list, fmt.Errorf("the path %s is not a directory", d)
		}

		files, err := FindFiles(d, glob)
		if err != nil {
			return list, err
		}

		list = append(list, files...)
	}

	return list, nil
}

// FindFiles retrieves a bunch of files from a directory
func FindFiles(path, glob string) ([]string, error) {
	var list []string
	files, err := filepath.Glob(fmt.Sprintf("%s/%s", path, glob))
	if err != nil {
		return list, err
	}
	for _, j := range files {
		if !IsFile(j) {
			continue
		}

		list = append(list, j)
	}

	return list, err
}

// ContainedIn checks if a value in a list of a strings
func ContainedIn(value string, list []string) bool {
	for _, x := range list {
		if x == value {
			return true
		}
	}

	return false
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	return stat.IsDir()
}

// IsFile checks if the path is a file
func IsFile(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !stat.IsDir()
}
