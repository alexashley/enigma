package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Enigma stores the configuration of the machine.
// Create a struct
type Enigma struct {
	Name       string
	Plugboard  map[string]string
	Rotors     [3]Rotor //TODO: change this to a slice?
	Reflector  map[string]string
	Stepping   bool // increment left every keystroke, middle every
	DoubleStep bool
	Log        *log.Logger `json:omitempty`
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
// the right rotor advances every keypress
// the middle rotor advances every 26 turns of the right rotor
// the left rotor advances every 26 turns of the middle rotor
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

// code is the encode/decode function for the Enigma's encryption
func (e *Enigma) code(msg string) string {
	var result string
	for _, r := range msg {
		// step the rotors
		e.step()
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
	err = json.Unmarshal(data, &e)
	e.initLog("stdout", "")
	return &e
}

// Rotor is a data structure for representing an Enigma rotor.
type Rotor struct {
	Name    string
	Wiring  Wiring
	Roffset int // ringstellung
	Step    int
	Notch   string
}

func abs(i int) int {
	if i < 0 {
		return i * -1
	}
	return i
}

// value returns the result of a pass through the rotor
func (r *Rotor) value(c string, reflected bool) string {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// adjust for step value for right-hand (entrance) contacts
	c = string(alphabet[(int(c[0]-'A')+r.Step)%26])
	c = r.Wiring.get(c, reflected)
	// adjust for step value for left-hand (exit) contacts
	c = string(alphabet[(abs(int(c[0]-'A')-r.Step))%26])
	return c
}

// Wiring is a data structure for representing the rotor wiring using 2 maps
type Wiring struct {
	Rmap map[string]string // mapping from the right side of the rotor
	Lmap map[string]string // mapping from the left side of the rotor
}

// initWiring is a constructor for the Wiring data structure
func (w *Wiring) initWiring(mapping map[string]string) {
	w.Rmap = mapping
	w.Lmap = revMap(mapping)
}

// get: given a key, return a value from the Wiring data structure
// The left param indicates that the signal has been reflected
func (w *Wiring) get(key string, left bool) string {
	if left {
		return w.Lmap[key]
	}
	return w.Rmap[key]
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
	//v.setStepping(false)
	v.Log.Println("ENCODED MSG:\t" + v.code("AJK"))
}
