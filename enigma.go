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
	Stepping  bool
	Log       *log.Logger `json:omitempty`
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
func (e *Enigma) step() {
	if e.Stepping {
		e.Rotors[0].Step += 1
	}
}

func (e *Enigma) reset() {
	for i := 0; i < len(e.Rotors); i++ {
		e.Rotors[i].Step = 0
	}
}

func (e *Enigma) setStepping(status bool) {
	e.Stepping = status
}

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

// saves Enigma configuration to a JSON file
func (e *Enigma) saveConfig(filename string) {
	data, _ := json.MarshalIndent(e, "", "\t")
	ioutil.WriteFile(filename, data, 0644)
}

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
	Rmap map[string]string
	Lmap map[string]string
}

// constructor for the Wiring data structure
func (w *Wiring) initWiring(mapping map[string]string) {
	w.Rmap = mapping
	w.Lmap = revMap(mapping)
}

// given a key, return a value from the Wiring data structure
// left: indicates that the signal has been reflected
func (w *Wiring) get(key string, left bool) string {
	if left {
		return w.Lmap[key]
	}
	return w.Rmap[key]
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
	v := loadConfig("config/M3.json")
	v.Log.Println("ENCODED MSG:\t" + v.code("AA"))
}
