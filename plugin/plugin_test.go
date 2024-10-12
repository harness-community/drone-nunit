// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"encoding/xml"
	"errors"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestRunner structure to match your format
type testRunner struct {
	name  string
	input string
	want  []string
	err   error
}

// TestRunner structure to match your format
type testRunnerProcessTest struct {
	name     string
	filePath string
	want     bool
	err      error
}

// TestApplyXSLTTransformation structure to match your format
type testApplyXSLTTransformation struct {
	name           string
	inputContent   string
	expectedPath   string
	expectedOutput string // Expected output content directly
	err            error
}

// TestFindTestFiles tests the findTestFiles function with various cases
func TestFindTestFiles(t *testing.T) {
	tests := []testRunner{
		// Valid Path with NUnit XML files in ../pluginTest/validTestXML
		{
			name:  "validPathWithXML",
			input: "../pluginTest/validTestXML/*.xml", // Adjusted to your folder path
			want:  []string{"../pluginTest/validTestXML/testPassed.xml", "../pluginTest/validTestXML/testFailed.xml", "../pluginTest/validTestXML/empty.xml"},
			err:   nil,
		},
		// Invalid Path with PLUGIN_FAIL_IF_NO_RESULTS set to false
		{
			name:  "invalidPathNoResultsFalse",
			input: "invalid/path/*.xml",
			want:  nil,
			err:   nil,
		},
		// Empty Path
		{
			name:  "emptyPath",
			input: "",
			want:  nil,
			err:   errors.New("Test Report Path should not be empty"),
		},
		// Directory with no XML files
		{
			name:  "noXMLFiles",
			input: "../pluginTest/emptyTestFolder/*.xml", // Change to a folder with no XML files
			want:  nil,
			err:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files, err := findTestFiles(tc.input)

			// Sort both the expected and actual file paths to ensure order doesn't cause mismatches
			sort.Strings(files)
			sort.Strings(tc.want)

			// Compare the files found with the expected output
			if diff := cmp.Diff(files, tc.want); diff != "" {
				t.Errorf("findTestFiles() mismatch (-want +got):\n%s", diff)
			}

			// Compare the error if any
			if tc.err != nil && err != nil {
				if err.Error() != tc.err.Error() {
					t.Errorf("findTestFiles() expected error: %v, got: %v", tc.err, err)
				}
			} else if err != tc.err {
				t.Errorf("findTestFiles() expected error: %v, got: %v", tc.err, err)
			}
		})
	}
}

func TestProcessTestResults(t *testing.T) {
	tests := []testRunnerProcessTest{
		// 1. Valid NUnit XML, all tests passed
		{
			name:     "validXMLAllPassed",
			filePath: "../pluginTest/validTestXML/testPassed.xml",
			want:     false,
			err:      nil,
		},
		// 2. Valid NUnit XML, some tests failed
		{
			name:     "validXMLWithFailures",
			filePath: "../pluginTest/validTestXML/testFailed.xml",
			want:     true,
			err:      nil,
		},
		// 3. Invalid file path
		{
			name:     "invalidFilePath",
			filePath: "../pluginTest/invalidPath/nonexistent.xml",
			want:     false,
			err:      os.ErrNotExist,
		},
		// 4. Empty NUnit XML file
		{
			name:     "emptyFile",
			filePath: "../pluginTest/emptyFolder/empty.xml",
			want:     false,
			err:      os.ErrNotExist,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := processTestResults(tc.filePath)

			// Compare the result with the expected output
			if diff := cmp.Diff(result, tc.want); diff != "" {
				t.Errorf("processTestResults() mismatch (-want +got):\n%s", diff)
			}

			// Handle error comparison
			if tc.err != nil {
				// Check if it's a "file not found" error
				if os.IsNotExist(err) && os.IsNotExist(tc.err) {
					// Both errors indicate file not found, so this is considered a pass
					return
				}

				// Check for XML syntax errors
				if _, ok := err.(*xml.SyntaxError); ok {
					if _, ok2 := tc.err.(*xml.SyntaxError); ok2 {
						// Both errors are syntax errors, so this is considered a pass
						return
					}
				}

				// For other errors, compare error messages
				if err == nil || cmp.Diff(err.Error(), tc.err.Error()) != "" {
					t.Errorf("processTestResults() expected error: %v, got: %v", tc.err, err)
				}
			} else if err != nil {
				t.Errorf("processTestResults() expected no error, got: %v", err)
			}
		})
	}
}
