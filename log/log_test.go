package log

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestFatal(t *testing.T) {
	var str bytes.Buffer
	ExitOnFatal = false
	origOut := Error.Writer()
	Error.SetOutput(&str)
	defer func() {
		ExitOnFatal = true
		Error.SetOutput(origOut)
	}()
	testStr := "test-string"

	Fatal(testStr)

	assert.True(t, strings.Contains(str.String(), testStr))
}

func TestWarnIfErr(t *testing.T) {
	var str bytes.Buffer
	origOut := Warn.Writer()
	Warn.SetOutput(&str)
	defer func() {
		Warn.SetOutput(origOut)
	}()
	testStr := "test-string"
	testDescr := "description"
	testError := errors.New(testStr)

	WarnIfErr(testDescr, testError)

	assert.True(t, strings.Contains(str.String(), testStr))
	assert.True(t, strings.Contains(str.String(), testDescr))
}

func TestErrIfErr(t *testing.T) {
	var str bytes.Buffer
	origOut := Error.Writer()
	Error.SetOutput(&str)
	defer func() {
		Error.SetOutput(origOut)
	}()
	testStr := "test-string"
	testDescr := "description"
	testError := errors.New(testStr)

	ErrIfErr(testDescr, testError)

	assert.True(t, strings.Contains(str.String(), testStr))
	assert.True(t, strings.Contains(str.String(), testDescr))
}
