package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Enigma stores the configuration of the machine.
type Enigma struct {
	Name          string
	Plugboard     Wiring            `json:"-"`
	Rotors        [3]Rotor          `json:"-"`
	Reflector     map[string]string `json:"-"`
	Stepping      bool
	DoubleStep    bool
	RotorBank     map[string]Rotor             // all the (available) rotors
	ReflectorBank map[string]map[string]string // all the reflectors
	Log           *log.Logger                  `json:"-"`
}

// initLog configures the log for an Enigma struct.
// The first argument is the destination of the log.
// stdout -> os.Stdout
// off -> ioutil.Discard
// "file" -> writes to given filename (given as second arg)
func (e *Enigma) initLog(logDest string, logFile string) {
	msg := ""
	flags := log.Ldate | log.Lmicroseconds | log.Lshortfile
	switch logDest {
	case "stdout":
		e.Log = log.New(os.Stdout, msg, flags)
	case "off":
		e.Log = log.New(ioutil.Discard, msg, flags)
	case "file":
		f, err := os.Create(logFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		e.Log = log.New(ioutil.Discard, msg, flags)
		e.Log.SetOutput(f)
	case "":
		e.Log = log.New(os.Stdout, msg, flags)
		e.Log.Println("Unrecognized log destination. Defaulting to stdout")
	}
}

// step handles the logic for rotor stepping.
// Note that the rotors advance *before* they see the signal.
// Below is the standard stepping sequence for an Enigma machine with 3 rotors:
// The right rotor advances every keypress.
// The middle rotor advances every 26 turns of the right rotor.
// The left rotor advances every 26 turns of the middle rotor.
func (e *Enigma) step() {
	if !e.Stepping {
		return
	}
	msg := "STEPPED ROTOR "
	e.Rotors[0].Step += 1
	e.Log.Println(msg + string(e.Rotors[0].Name))
	for i := 1; i < len(e.Rotors); i++ {
		if e.Rotors[i-1].Step >= 26 && (e.Rotors[i-1].Step%26) == 0 {
			e.Rotors[i].Step += 1
			e.Log.Println(msg + string(e.Rotors[i].Name))
		}
	}
}

// reset clears the internal state of the machine.
// Used mainly for testing
func (e *Enigma) reset() {
	for i := 0; i < len(e.Rotors); i++ {
		e.Rotors[i].Step = 0
	}
}

// setStepping enables/disables rotor stepping
func (e *Enigma) setStepping(status bool) {
	e.Stepping = status
}

func (e *Enigma) setRotorPosition(rotorName string, position string) {
	rotor := e.RotorBank[rotorName]
	switch position {
	case "right":
		e.Rotors[0] = rotor
	case "middle":
		e.Rotors[1] = rotor
	case "left":
		e.Rotors[2] = rotor
	}
}

func (e *Enigma) setReflector(reflectorName string) {
	reflector := e.ReflectorBank[reflectorName]
	e.Reflector = reflector
}

// isUpperCaseAscii checks if input is in the ASCII range A ... Z
func isUppercaseAscii(b byte) bool {
	return (b >= 'A' && b <= 'Z')
}

// validate returns a string that the Enigma can encode
// requirements: ASCII A-Z, no numbers or punctuation
// no doubt there is some DP soln for this that is much more efficient
func validate(s string) string {
	replace := " "
	s = strings.ToUpper(s)
	for i := 0; i < len(s); i++ {
		if !isUppercaseAscii(s[i]) {
			replace += string(s[i])
		}
	}
	for i := 0; i < len(replace); i++ {
		s = strings.Replace(s, string(replace[i]), "", -1)
	}
	return s
}

// code is the encode/decode function for the Enigma's encryption
// There are 4  main components to the Enigma encryption process
// Plugboard: operator can create a mapping between letters
// Static rotor: maps signal from plugboard wires to rotor contacts
// Rotors: the main encryption mechanism for the Enigma
// Reflector: redirects the signal back through the rotors
// The forward signal path for an M3:
// Plugboard -> Static Rotor -> R rotor -> M rotor -> L rotor
// Then the signal is hits the reflector and does the reverse journey
// L rotor -> M rotor -> R rotor -> static rotor -> plugboard
// msg: string to encode.
// chunkSize: length of output chunks, separated by spaces. -1 returns 1 chunk
func (e *Enigma) code(msg string, chunkSize int) string {
	var result string
	msg = validate(msg)
	for _, r := range msg {
		// step the rotors
		e.step()
		c := string(r)
		e.Log.Println("ENCODING:\t" + c)
		// plugboard mapping, if one exists
		if p := e.Plugboard.get(c, false); p != "" {
			e.Log.Println("PLUGBOARD:\t" + p)
			c = p
		}
		// forward signal path
		for i := 0; i < len(e.Rotors); i++ {
			rotor := e.Rotors[i]
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
		// plugboard return
		if p := e.Plugboard.get(c, true); p != "" {
			e.Log.Println("PLUGBOARD:\t" + p)
			c = p
		}
		result += c
	}
	var s string = ""
	if chunkSize != -1 {
		for i := 0; i < len(result); i++ {
			if i%chunkSize == 0 {
				s += " "
			}
			s += string(result[i])
		}
		result = s
	}
	return result
}

// saveConfig marshals the Enigma state into json and saves it to disk.
func (e *Enigma) saveConfig(filename string) {
	data, _ := json.MarshalIndent(e, "", "\t")
	ioutil.WriteFile(filename, data, 0644)
}

// loadConfig initializes an Enigma by loading a json configuration file.
func loadConfig(filename string) *Enigma {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var e Enigma
	if err = json.Unmarshal(data, &e); err != nil {
		panic(err)
	}
	e.initLog("stdout", "")
	return &e
}

// Rotor is a data structure for representing an Enigma rotor.
type Rotor struct {
	Name   string
	Wiring Wiring
	Step   int
	Notch  [2]string
}

func abs(i int) int {
	if i < 0 {
		return i * -1
	}
	return i
}

func makeRotor(name string, mapping string, notches string) Rotor {
	var r Rotor
	r.Name = name
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	w := make(map[string]string)
	for i := 0; i < len(alphabet); i++ {
		k := string(alphabet[i])
		v := string(mapping[i])
		w[k] = v
	}
	r.Wiring.initWiring(w)
	r.Notch[0] = string(notches[0])
	if len(notches) == 2 {
		r.Notch[1] = string(notches[1])
	}
	return r
}

// value returns the result of a pass through the rotor
func (r *Rotor) value(c string, reflected bool) string {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	offset := r.Step
	// adjust step value for entrance  contacts (right forward, left return)
	c = string(alphabet[(abs(int(c[0]-'A')+offset))%26])
	// map for rotor wiring core
	c = r.Wiring.get(c, reflected)
	// adjust step value for exit contacts (left forward, right return)
	c = string(alphabet[abs(26+int(c[0]-'A')-offset)%26])
	return c
}

// Wiring is a data structure for representing the rotor and plugboard wiring
type Wiring struct {
	// forward mapping: right side of rotor/forward signal through plugboard
	Fmap map[string]string
	// reverse mapping: left side of rotor/return signal through plugboard
	Rmap map[string]string
}

// initWiring is a constructor for the Wiring data structure
func (w *Wiring) initWiring(mapping map[string]string) {
	w.Fmap = mapping
	w.Rmap = revMap(mapping)
}

// get: given a key, return a value from the Wiring data structure
func (w *Wiring) get(key string, reverse bool) string {
	if reverse {
		return w.Rmap[key]
	}
	return w.Fmap[key]
}

// revMap reverses a map. k:v -> v:k
func revMap(m map[string]string) map[string]string {
	mRev := make(map[string]string)
	for k, v := range m {
		mRev[v] = k
	}
	return mRev
}

func main() {
	v := loadConfig("config/M3.json")
	//var v Enigma
	/*v.Name = "M3 Wehrmacht"
	v.Reflector = map[string]string{
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
	v.Stepping = true
	v.DoubleStep = true
	names := []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII"}
	maps := []string{
		"EKMFLGDQVZNTOWYHXUSPAIBRCJ",
		"AJDKSIRUXBLHWTMCQGZNPYFVOE",
		"BDFHJLCPRTXVZNYEIWGAKMUSQO",
		"ESOVPZJAYQUIRHXLNFTGKDCMWB",
		"VZBRGITYUPSDNHLXAWMJQOFECK",
		"JPGVOUMFYQBENHZRDKASXLICTW",
		"NZJHGRCXMYSWBOUFAIVLPEKQDT",
		"FKQHTLXOCBJSPDZRAMEWNIUYGV",
	}
	notches := []string{"Y", "M", "D", "R", "H", "HU", "HU", "HU"}
	v.RotorBank = make(map[string]Rotor)
	for i, r := range names {
		v.RotorBank[r] = makeRotor(r, maps[i], notches[i])
	}
	v.ReflectorBank = make(map[string]map[string]string)
	rNames := []string{"B", "C"}
	reflectors := []string{
		"YRUHQSLDPXNGOKMIEBFZCWVJAT",
		"FVPJIAOYEDRZXWGCTKUQSBNMHL",
	}
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i, r := range reflectors {
		mapping := make(map[string]string)
		for idx, c := range r {
			k := string(alphabet[idx])
			v := string(c)
			mapping[k] = v
		}
		v.ReflectorBank[rNames[i]] = mapping
	}*/
	v.setRotorPosition("I", "right")
	v.setRotorPosition("II", "middle")
	v.setRotorPosition("III", "left")
	v.setReflector("B")
	v.saveConfig("config/M3.json")
	//v.Log.Println(validate(msg))
	msg := "AQRAFDADFGBAK"
	v.Log.Println("ENCODED MSG:\t" + v.code(msg, 5))
}
