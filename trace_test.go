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
	func testFunc4(a,   b int) {return false}`)

	testSrc5 = []byte(`package none
func testFunc5(a, b int) {
	func(a int) {
		a = b
	}(a)
}`)
)

func annotateTest(source []byte, t *testing.T) []byte {
	processed, err := annotate("test.go", source)
	if err != nil {
		fmt.Println(string(processed))
		t.Fatal(err)
	}
	t.Log(string(processed))
	return processed
}

func returnsOn() func() {
	prev := showReturn
	showReturn = true
	return func() {
		showReturn = prev
	}
}

func timingOn() func() {
	prevT := timing
	timing = true
	prevR := showReturn
	showReturn = true
	return func() {
		timing = prevT
		showReturn = prevR
	}
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
	defer returnsOn()()
	annotateTest(testSrc1, t)
}

func TestTiming(t *testing.T) {
	defer timingOn()()
	annotateTest(testSrc4, t)
}

func TestEmbedded(t *testing.T) {
	defer timingOn()()
	annotateTest(testSrc5, t)
}
