package enigma

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	alphabet string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Enigma stores the configuration of the machine.
type Enigma struct {
	Name          string
	Plugboard     Wiring   `json:"-"`
	Rotors        [3]Rotor `json:"-"`
	Reflector     Rotor    `json:"-"`
	Stepping      bool
	DoubleStep    bool
	RotorBank     map[string]Rotor // all the (available) rotors
	ReflectorBank map[string]Rotor // all the reflectors
	Log           *log.Logger      `json:"-"`
}

func (e *Enigma) InitEnigma(left string, middle string, right string, reflector string) {
	e.SetRotorPosition(left, "left")
	e.SetRotorPosition(middle, "middle")
	e.SetRotorPosition(right, "right")
	e.SetReflector("B")
}

// initLog configures the log for an Enigma struct.
// The first argument is the destination of the log.
// stdout -> os.Stdoutr
// off -> ioutil.Discard
// "file" -> writes to given filename (given as second arg)
func (e *Enigma) InitLog(logDest string, logFile string) {
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
	if e.DoubleStep {
		dmsg := "DOUBLE STEP "
		window := string(alphabet[e.Rotors[0].Step%26])
		afterTurnoverLetters := ""
		var idx int
		var t string
		for i := 0; i < len(e.Rotors[0].Turnover); i++ {
			t = string(e.Rotors[0].Turnover[i])
			idx = strings.Index(alphabet, string(t)) + 1
			afterTurnoverLetters += string(alphabet[idx])
		}
		if strings.Index(window, afterTurnoverLetters) != -1 {
			e.Log.Println(dmsg + e.Rotors[1].Name)
			e.Rotors[1].Step += 1
		}
	}
}

// reset clears the internal state of the machine.
// Used mainly for testing
func (e *Enigma) Reset() {
	for i := 0; i < len(e.Rotors); i++ {
		e.Rotors[i].Step = 0
	}
}

// setStepping enables/disables rotor stepping
func (e *Enigma) SetStepping(status bool) {
	e.Stepping = status
}

func (e *Enigma) SetRotorPosition(rotorName string, position string) {
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

func (e *Enigma) SetReflector(reflectorName string) {
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
// chunkSize: length of output chunks, separated by spaces. -1 -> 1 chunk
func (e *Enigma) Code(msg string, chunkSize int) string {
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
			c = rotor.value(c, false, true)
			e.Log.Println("ROTOR " + rotor.Name + ":\t" + c)
		}
		// reflector
		c = e.Reflector.value(c, false, false)
		e.Log.Println("REFLECTOR:\t" + c)
		// reverse signal path
		for i := len(e.Rotors) - 1; i >= 0; i-- {
			c = e.Rotors[i].value(c, true, true)
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
func (e *Enigma) SaveConfig(filename string) {
	data, _ := json.MarshalIndent(e, "", "\t")
	ioutil.WriteFile(filename, data, 0644)
}

// loadConfig initializes an Enigma by loading a json configuration file.
func LoadConfig(filename string) *Enigma {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var e Enigma
	if err = json.Unmarshal(data, &e); err != nil {
		panic(err)
	}
	e.InitLog("stdout", "")
	return &e
}

// Rotor is a data structure for representing an Enigma rotor.
type Rotor struct {
	Name     string
	Wiring   string
	Step     int
	Turnover string
}

func abs(i int) int {
	if i < 0 {
		return i * -1
	}
	return i
}

// value returns the result of a pass through the rotor
func (r *Rotor) value(c string, reflected bool, rotated bool) string {
	offset := r.Step
	// adjust step value for entrance  contacts (right forward, left return)
	c = string(alphabet[(abs(int(c[0]-'A')+offset))%26])
	// rotor wiring core
	if !reflected {
		c = string(r.Wiring[strings.Index(alphabet, string(c))])
	} else {
		c = string(alphabet[strings.Index(r.Wiring, string(c))])
	}
	// adjust step value for exit contacts (left forward, right return)
	c = string(alphabet[abs(26+int(c[0]-'A')-offset)%26])
	return c
}

func (r *Rotor) setRing(letter string) {
	r.Step = strings.Index(alphabet, letter)
}

type Wiring struct {
	// forward mapping: forward signal through plugboard
	Fmap map[string]string
	// reverse mapping: return signal through plugboard
	Rmap map[string]string
}

// initWiring is a constructor for the Wiring data structure
func (w *Wiring) initWiring(mapping map[string]string) {
	w.Fmap = mapping
	w.Rmap = make(map[string]string) //revMap(mapping)
	for k, v := range w.Fmap {
		w.Rmap[v] = k
	}
}

// get: given a key, return a value from the Wiring data structure
func (w *Wiring) get(key string, reverse bool) string {
	if reverse {
		return w.Rmap[key]
	}
	return w.Fmap[key]
}
