package main

import "testing"

func initEnigma(config string) *Enigma {
	e := loadConfig(config)
	e.initLog("off", "")
	return e
}

func TestSteplessEncode(t *testing.T) {
	e := initEnigma("config/M3.json")
	if e.code("A") != "N" {
		t.Fail()
	}
	e = initEnigma("config/M3.json")
	if e.code("N") != "A" {
		t.Fail()
	}
}
