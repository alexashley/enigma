# Enigma
An [Enigma](https://en.wikipedia.org/wiki/Enigma_machine) emulator written in Go.
## Status
Currently only supports stepless encoding (essentially a convoluted substitution cipher) on an M3. 
## Usage
````
e := loadConfig("config/M3.json")
e.code("super secret message")
```
## Resources
http://users.telenet.be/d.rijmenants/en/enigmatech.htm)

http://people.physik.hu-berlin.de/~palloks/js/enigma/enigma-u_v20_en.html

