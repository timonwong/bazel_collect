package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"

	junit "github.com/joshdk/go-junit"
)

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName xml.Name `xml:"testsuites"`
	Suites  []JUnitTestSuite
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      uint64          `xml:"tests,attr"`
	Failures   uint64          `xml:"failures,attr"`
	Time       string          `xml:"time,attr"`
	Name       string          `xml:"name,attr"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []JUnitTestCase
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	SkipMessage *JUnitSkipMessage `xml:"skipped,omitempty"`
	Failure     *JUnitFailure     `xml:"failure,omitempty"`
}

// JUnitSkipMessage contains the reason why a testcase was skipped.
type JUnitSkipMessage struct {
	Message string `xml:"message,attr"`
}

// JUnitProperty represents a key/value pair used to define properties.
type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// JUnitFailure contains data related to a failed test.
type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

func MergeJunit(paths []string, output string) {
	suites, err := junit.IngestFiles(paths)
	if err != nil {
		log.Fatal(err)
	}
	var testSuite JUnitTestSuite
	var passed uint64
	var failed uint64
	for _, suite := range suites {
		for _, test := range suite.Tests {
			tc := JUnitTestCase{
				Name:      test.Name,
				Classname: test.Classname,
				Time:      test.Duration.String(),
			}
			if test.Status == junit.StatusPassed {
				passed = passed + 1
			}
			if test.Status == junit.StatusSkipped {
				tc.SkipMessage = &JUnitSkipMessage{
					Message: test.Message,
				}
				passed = passed + 1
			}
			if test.Status == junit.StatusFailed || test.Status == junit.StatusError {
				tc.Failure = &JUnitFailure{
					Message:  test.Message,
					Contents: test.Error.Error(),
				}
				failed = failed + 1
			}
			testSuite.TestCases = append(testSuite.TestCases, tc)
		}
	}
	testSuite.Tests = passed + failed
	testSuite.Failures = failed
	file, err := xml.MarshalIndent(testSuite, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(output, file, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
