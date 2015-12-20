package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Enigma struct {
	Name      string
	Plugboard map[string]string
	Rotors    [3]Rotor //TODO: change this to a slice?
	Reflector map[string]string
	Log       *log.Logger
}

// TODO: add log to file support
func (e *Enigma) initLog(logDest string, logFile string) {
	msg := ""
	flags := log.Ldate | log.Lmicroseconds | log.Lshortfile
	switch logDest {
	case "stdout":
		e.Log = log.New(os.Stdout, msg, flags)
	case "off":
		e.Log = log.New(ioutil.Discard, msg, flags)
	case "":
		e.Log = log.New(os.Stdout, msg, flags)
		e.Log.Println("Unrecognized log destination. Defaulting to stdout")
	}
}
func (e *Enigma) code(msg string) string {
	var result string
	for _, r := range msg {
		c := string(r)
		e.Log.Println("ENCODING:\t" + c)
		// plugboard mapping, if one exists
		if p := e.Plugboard[c]; p != "" {
			e.Log.Println("PLUGBOARD:\t" + p)
			c = p
		}
		// forward signal path
		for _, rotor := range e.Rotors {
			c = rotor.value(c, false)
			e.Log.Println("ROTOR " + rotor.Name + ":\t" + c)
		}
		// reflector
		c = e.Reflector[c]
		e.Log.Println("REFLECTOR:\t" + c)
		// reverse signal path
		for i := len(e.Rotors) - 1; i >= 0; i-- {
			c = e.Rotors[i].value(c, true)
			e.Log.Println("ROTOR " + e.Rotors[i].Name + ":\t" + c)
		}
		result += c

	}
	return result
}

// saves Enigma configuration to a JSON file
func (e *Enigma) saveConfig(fn string) {
	json, _ := json.MarshalIndent(e, "", "\t")
	ioutil.WriteFile(fn, json, 0644)
}

type Rotor struct {
	Name    string
	Wiring  Wiring
	Roffset int // ringstellung
	Step    int
	Notch   string
}

func (r *Rotor) value(c string, reflected bool) string {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// adjust for step value for right-hand (entrance) contacts
	c = string(alphabet[(int(c[0]-'A')+r.Step)%26])
	c = r.Wiring.get(c, reflected)
	// adjust for step value for left-hand (exit) contacts
	c = string(alphabet[(int(c[0]-'A')-r.Step)%26])
	return c
}

/*
* a data structure for representing the rotor wiring using 2 maps
* rMap: the signal has come from the right side rotor
* lMap: the signal has come from the left side rotor
 */
type Wiring struct {
	rMap map[string]string
	lMap map[string]string
}

// constructor for the Wiring data structure
func (w *Wiring) initWiring(mapping map[string]string) {
	w.rMap = mapping
	w.lMap = revMap(mapping)
}

// given a key, return a value from the Wiring data structure
// left: indicates that the signal has been reflected
func (w *Wiring) get(key string, left bool) string {
	if left {
		return w.lMap[key]
	}
	return w.rMap[key]
}

// reverses a map. k:v -> v:k
func revMap(m map[string]string) map[string]string {
	mRev := make(map[string]string)
	for k, v := range m {
		mRev[v] = k
	}
	return mRev
}
func main() {
	var q Enigma
	q.Name = "M3"
	q.Plugboard = map[string]string{}
	//var r1, r2 Rotor
	q.Rotors[0].Name = "I"
	q.Rotors[0].Wiring.initWiring(map[string]string{
		"A": "E",
		"B": "K",
		"C": "M",
		"D": "F",
		"E": "L",
		"F": "G",
		"G": "D",
		"H": "Q",
		"I": "V",
		"J": "Z",
		"K": "N",
		"L": "T",
		"M": "O",
		"N": "W",
		"O": "Y",
		"P": "H",
		"Q": "X",
		"R": "U",
		"S": "S",
		"T": "P",
		"U": "A",
		"V": "I",
		"W": "B",
		"X": "R",
		"Y": "C",
		"Z": "J",
	})
	q.Rotors[1].Name = "II"
	q.Rotors[1].Wiring.initWiring(map[string]string{
		"A": "A",
		"B": "J",
		"C": "D",
		"D": "K",
		"E": "S",
		"F": "I",
		"G": "R",
		"H": "U",
		"I": "X",
		"J": "B",
		"K": "L",
		"L": "H",
		"M": "W",
		"N": "T",
		"O": "M",
		"P": "C",
		"Q": "Q",
		"R": "G",
		"S": "Z",
		"T": "N",
		"U": "P",
		"V": "Y",
		"W": "F",
		"X": "V",
		"Y": "O",
		"Z": "E",
	})
	q.Rotors[2].Name = "III"
	q.Rotors[2].Wiring.initWiring(map[string]string{
		"A": "B",
		"B": "D",
		"C": "F",
		"D": "H",
		"E": "J",
		"F": "L",
		"G": "C",
		"H": "P",
		"I": "R",
		"J": "T",
		"K": "X",
		"L": "V",
		"M": "Z",
		"N": "N",
		"O": "Y",
		"P": "E",
		"Q": "I",
		"R": "W",
		"S": "G",
		"T": "A",
		"U": "K",
		"V": "M",
		"W": "U",
		"X": "S",
		"Y": "Q",
		"Z": "O",
	})
	q.Reflector = map[string]string{
		"A": "Y",
		"B": "R",
		"C": "U",
		"D": "H",
		"E": "Q",
		"F": "S",
		"G": "L",
		"H": "D",
		"I": "P",
		"J": "X",
		"K": "N",
		"L": "G",
		"M": "O",
		"N": "K",
		"O": "M",
		"P": "I",
		"Q": "E",
		"R": "B",
		"S": "F",
		"T": "Z",
		"U": "C",
		"V": "W",
		"W": "V",
		"X": "J",
		"Y": "A",
		"Z": "T",
	}
	q.initLog("stdout", "")
	q.saveConfig(q.Name)
	q.code("N")
}
