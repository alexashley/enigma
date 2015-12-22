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

// testHelper is the main testing template
// Tests encryption and decryption of message
func testHelper(t *testing.T, e *Enigma, msg string, encrypted string) {
	actual := e.code(msg)
	assert(t, encrypted, actual)
	e.reset()
	assert(t, e.code(encrypted), msg)
	e.reset()
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
	testHelper(t, M3, "A", "N")
	M3.setStepping(true)

}

func TestM3SteplessEncodeLong(t *testing.T) {
	M3.setStepping(false)
	original := "QWERTYUIOPASDFGHJKLZXCVBNM"
	encoded := "SMHIVZDRKLNQUBJEGOPYCXTFAW"
	testHelper(t, M3, original, encoded)
	M3.setStepping(true)
}

func TestM3StepEncodeRepeatedLetters(t *testing.T) {
	testHelper(t, M3, "AA", "FT")
}

func TestM3EncodeLong(t *testing.T) {
	original := "AQRAFDADFGBAK"
	encoded := "FIFMMESGOLQWM"
	testHelper(t, M3, original, encoded)
}

func TestInvalidInput(t *testing.T) {
	original := "1234kwisatz@hader2ach"
	encoded := "NNJJGXXWIMIZGTQ"
	assert(t, encoded, M3.code(original))
	M3.reset()
}
