# Enigma
An [Enigma](https://en.wikipedia.org/wiki/Enigma_machine) emulator written in Go.
## Status
Currently only supports the M3

TO-DO:
- shiftable static rotor
- additional machine types

## Features
- Can be extended through the use of JSON configuration files (see config directory)
- Logging to file/stdout so the machine execution can be traced.
- Tests for each machine type 

## Usage
Example:
```go
func main() {
  // load basic M3 configuration
  e := enigma.LoadConfig("config/M3.json")
  // rotors: left middle right, then reflector
  e.InitEnigma("I", "II", "III", "B")
  // output: XLNZB CSCQQ PWWFR UEGOH NMLPU ZIM
  e.Log.Println(e.code("So long and thanks for all the fish!", 5)) 
}
```
Run the tests with `go test`
## Resources
http://users.telenet.be/d.rijmenants/en/enigmatech.htm

http://people.physik.hu-berlin.de/~palloks/js/enigma/enigma-u_v20_en.html

