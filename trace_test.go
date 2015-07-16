package main

import (
	"fmt"
	"log"
	"regexp"
	"testing"
)

func init() {
	setup = fmt.Sprintf(setup, prefix, formatLength)

	var err error
	filter, err = regexp.Compile(filterFlag)
	if err != nil {
		log.Fatal(err)
	}
}

var (
	testSrc1 = []byte(`package none

func testFunc1(a, b string) string {
	return a + b
}`)

	testSrc2 = []byte(`package none

func testFunc2(a, _ string) int {
	return 1
}`)

	testSrc3 = []byte(`package none

func testFunc3(a ...interface{}) bool {
	return true
}`)

	testSrc4 = []byte(`package none
	func testFunc4(a,   b) {return false}`)
)

func annotateTest(source []byte, t *testing.T) []byte {
	processed, err := annotate(source)
	if err != nil {
		fmt.Println(string(processed))
		t.Fatal(err)
	}
	t.Log(string(processed))
	return processed
}

// TODO: Use go/types to further check the output, since we don't execute the
//       new source.
//       I'll add that once go/types is moved into the main repo

// Check the syntax on a basic function
func TestBasic(t *testing.T) {
	annotateTest(testSrc1, t)
}

// Check that we don't fail on an un-named argument
func TestUnderscore(t *testing.T) {
	annotateTest(testSrc2, t)
}

// We should handle variadic just fine
func TestVariadic(t *testing.T) {
	annotateTest(testSrc3, t)
}

// Make sure we can handle improperly formatted source
func TestUnFmted(t *testing.T) {
	annotateTest(testSrc4, t)
}

// test output with return logging
func TestReturns(t *testing.T) {
	prev := showReturn
	defer func() {
		showReturn = prev
	}()
	showReturn = true
	annotateTest(testSrc1, t)
}

func TestTiming(t *testing.T) {
	prevR := showReturn
	prevT := timing
	defer func() {
		timing = prevT
		showReturn = prevR
	}()
	timing = true
	showReturn = true

	annotateTest(testSrc4, t)
}
