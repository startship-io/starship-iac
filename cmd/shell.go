// Copyright 2021 starship studio.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmd provides all CLI commands.
// NOTE: These are different from the commands that get run via pull request
// comments.
package cmd

import (
	"bufio"
	"os"
	"os/exec"
)

func Exec(command string, params []string) *exec.Cmd {
	cmd := exec.Command(command, params...)
	cmd.Stderr = os.Stderr
	return cmd
}

func ReadLog(filePath string, lineNumber int) ([]string, int) {
	file, _ := os.Open(filePath)
	fileScanner := bufio.NewScanner(file)
	lineCount := 1
	var lines []string
	for fileScanner.Scan() {
		if lineCount >= lineNumber {
			lines = append(lines, fileScanner.Text())
		}
		lineCount++
	}
	defer file.Close()
	return lines, lineCount - 1
}
