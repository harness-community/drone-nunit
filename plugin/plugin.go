package plugin

import (
	"context"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/wamuir/go-xslt"
)

// Args provides plugin execution arguments.
type Args struct {
	// Level defines the plugin log level.
	Level                      string `envconfig:"PLUGIN_LOG_LEVEL"`
	PluginTestReportPath       string `envconfig:"PLUGIN_TEST_REPORT_PATH"`
	PluginFailIfNoResults      bool   `envconfig:"PLUGIN_FAIL_IF_NO_RESULTS"`
	PluginFailedTestsFailBuild bool   `envconfig:"PLUGIN_FAILED_TESTS_FAIL_BUILD"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {

	logger := logrus.
		WithField("PLUGIN_TEST_REPORT_PATH", args.PluginTestReportPath).
		WithField("PLUGIN_FAIL_IF_NO_RESULTS", args.PluginFailIfNoResults).
		WithField("PLUGIN_FAILED_TESTS_FAIL_BUILD", args.PluginFailedTestsFailBuild)

	logger.Info("Starting plugin execution")

	// Find the test report files based on the PLUGIN_TEST_REPORT_PATH
	files, err := findTestFiles(args.PluginTestReportPath)
	if err != nil {
		logger.WithError(err).Error("Failed to find test files")
		return errors.New("failed to find test files")
	}

	// Log the number of test files found
	logger.Infof("Found %d test report file(s)", len(files))

	// If no files are found and PLUGIN_FAIL_IF_NO_RESULTS is true, return an error
	if len(files) == 0 {
		if args.PluginFailIfNoResults {
			errMsg := "no test results found, failing the build as PLUGIN_FAIL_IF_NO_RESULTS is set to true"
			logger.Error(errMsg)
			return errors.New(errMsg)
		} else {
			logger.Warn("No test results found, but failing the build is not configured.")
		}
	}

	// Flag to track if any test failed
	var testFailed bool

	// Process the test result files
	for _, file := range files {
		failed, err := processTestResults(file)
		if err != nil {
			logger.WithError(err).Errorf("Failed to process test result file %s", file)
			continue
		}

		if failed {
			logger.Warnf("Test results indicate failure in file: %s", file)
			testFailed = true
		}

		conversionErr := applyXSLTTransformation(file, logger)

		if conversionErr != nil {
			errMsg := "Build failed, XSLT transformation from NUnit to JUnit failed"
			logger.Error(conversionErr)
			return errors.New(errMsg)
		}
	}

	// After processing and transforming all files, check if any test failed
	if testFailed && args.PluginFailedTestsFailBuild {
		errMsg := "tests failed, failing the build as PLUGIN_FAILED_TESTS_FAIL_BUILD is set to true"
		logger.Error(errMsg)
		return errors.New(errMsg)
	}

	logger.Info("Plugin execution completed successfully")
	return nil
}

// findTestFiles locates the test result files based on the path pattern provided.
func findTestFiles(reportPath string) ([]string, error) {

	if len(reportPath) == 0 {
		errMsg := "Test Report Path should not be empty"
		return nil, errors.New(errMsg)
	}

	files, err := filepath.Glob(reportPath)

	if err != nil {
		return nil, err
	}
	return files, nil
}

// processTestResults parses the NUnit XML file to check if any tests failed.
func processTestResults(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logrus.Errorf("Failed to open test result file %s: %v", filePath, err)
		return false, errors.New("failed to open test result file")
	}
	defer file.Close()

	var testRun TestRun
	if err := xml.NewDecoder(file).Decode(&testRun); err != nil {
		logrus.Errorf("Failed to parse test result XML file %s: %v", filePath, err)
		return false, errors.New("failed to parse test result XML")
	}

	// If there are failed tests, return true
	return testRun.Failed > 0, nil
}

// applyXSLTTransformation applies the XSLT transformation to convert NUnit XML to JUnit XML.
func applyXSLTTransformation(filePath string, logger *logrus.Entry) error {
	// Define the path to the XSLT file (bundled within your Docker image)
	xslFilePath := "/docs/conversionStyleSheet.xsl"

	// Load the NUnit XML input
	input, err := os.ReadFile(filePath)
	if err != nil {
		logger.Errorf("Failed to read NUnit XML file %s: %v", filePath, err)
		return errors.New("failed to read NUnit XML file")
	}
	// Load the XSLT content
	xsltContent, err := os.ReadFile(xslFilePath)
	if err != nil {
		logger.Errorf("Failed to read XSLT file %s: %v", xslFilePath, err)
		return errors.New("failed to read XSLT file")
	}

	// Create a new stylesheet
	xs, err := xslt.NewStylesheet(xsltContent)
	if err != nil {
		logger.Errorf("Failed to create stylesheet: %v", err)
		return errors.New("failed to create stylesheet")
	}

	// Apply the XSLT transformation
	transformed, err := xs.Transform(input)
	if err != nil {
		logger.Errorf("Failed to apply XSLT transformation to file %s: %v", filePath, err)
		return errors.New("failed to apply XSLT transformation to file")
	}
	defer xs.Close()

	// Write the transformed content back to the same file
	junitFilePath := filePath
	err = os.WriteFile(junitFilePath, transformed, 0644)
	if err != nil {
		logger.Errorf("Failed to write transformed JUnit XML to file %s: %v", junitFilePath, err)
		return errors.New("failed to write transformed JUnit XML to file")
	}

	logger.Debugf("Successfully wrote transformed file to: %s", filePath)

	return nil
}
