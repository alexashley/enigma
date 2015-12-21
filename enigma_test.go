package main

import (
	"os"
	"testing"
)

var (
	M3 *Enigma = initEnigma("config/M3.json")
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func assert(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Error("Expected string " + expected + " but was " + actual)
	}
}

func initEnigma(config string) *Enigma {
	e := loadConfig(config)
	e.initLog("off", "")
	return e
}

func TestM3SteplessEncodeShort(t *testing.T) {
	M3.setStepping(false)
	assert(t, "N", M3.code("A"))
	M3.reset()
	assert(t, "A", M3.code("N"))
	M3.setStepping(true)
}

func TestM3SteplessEncodeLong(t *testing.T) {
	M3.setStepping(false)
	original := "QWERTYUIOPASDFGHJKLZXCVBNM"
	encoded := "SMHIVZDRKLNQUBJEGOPYCXTFAW"
	assert(t, encoded, M3.code(original))
	M3.reset()
	assert(t, original, M3.code(encoded))
	M3.setStepping(true)
}

func TestM3StepEncodeRepeatedLetters(t *testing.T) {
	assert(t, "FT", M3.code("AA"))
	M3.reset()
	assert(t, "AA", M3.code("FT"))
}
