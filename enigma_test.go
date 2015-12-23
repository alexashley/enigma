package enigma

import (
	"os"
	"testing"
)

var (
	M3 *Enigma = initM3("config/M3.json")
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// testHelper is the main testing template
// Tests encryption and decryption of message
func testHelper(t *testing.T, e *Enigma, msg string, encrypted string) {
	actual := e.Code(msg, -1)
	assert(t, encrypted, actual)
	e.Reset()
	assert(t, e.Code(encrypted, -1), msg)
	e.Reset()
}

func assert(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Error("Expected string " + expected + " but was " + actual)
	}
}

func initM3(config string) *Enigma {
	e := LoadConfig(config)
	e.SetRotorPosition("I", "right")
	e.SetRotorPosition("II", "middle")
	e.SetRotorPosition("III", "left")
	e.SetReflector("B")
	e.InitLog("off", "")
	return e
}

func TestM3SteplessEncodeShort(t *testing.T) {
	M3.SetStepping(false)
	testHelper(t, M3, "A", "N")
	M3.SetStepping(true)

}

func TestM3SteplessEncodeLong(t *testing.T) {
	M3.SetStepping(false)
	original := "QWERTYUIOPASDFGHJKLZXCVBNM"
	encoded := "SMHIVZDRKLNQUBJEGOPYCXTFAW"
	testHelper(t, M3, original, encoded)
	M3.SetStepping(true)
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
	assert(t, encoded, M3.Code(original, -1))
	M3.Reset()
}

func TestM3DoubleStep(t *testing.T) {
	original := "So long and thanks for all the fish"
	encoded := "XLNZBCSCQQPWWFRUEGOHNMLPUZIM"
	assert(t, encoded, M3.Code(original, -1))
	M3.Reset()
}
