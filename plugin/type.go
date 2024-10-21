package plugin

import (
	"encoding/xml"
)

type TestRun struct {
	XMLName    xml.Name    `xml:"test-run"`
	Total      int         `xml:"total,attr"`
	Passed     int         `xml:"passed,attr"`
	Failed     int         `xml:"failed,attr"`
	Result     string      `xml:"result,attr"`
	TestSuites []TestSuite `xml:"test-suite"`
}

type TestSuite struct {
	TestCases []TestCase `xml:"test-case"`
}

type TestCase struct {
	Result string `xml:"result,attr"`
}
