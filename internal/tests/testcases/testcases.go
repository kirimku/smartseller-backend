// Package testcases provides test data for integration tests
package testcases

// TestCase represents test case data for integration tests
type TestCase struct {
	Name     string
	Input    interface{}
	Expected interface{}
}

// 3PL test cases can be defined here
