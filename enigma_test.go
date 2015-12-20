package main

import "testing"

func initEnigma(config string) *Enigma {
	e := loadConfig(config)
	e.initLog("off", "")
	return e
}

func TestM3SteplessEncode(t *testing.T) {
	e := initEnigma("config/M3.json")
	e.setStepping(false)
	if e.code("A") != "N" {
		t.Fail()
	}
	e.reset()
	if e.code("N") != "A" {
		t.Fail()
	}
}

func TestM3StepEncode(t *testing.T) {
	e := initEnigma("config/M3.json")
	if e.code("AA") != "FT" {
		t.Fail()
	}
	e.reset()
	if e.code("FT") != "AA" {
		t.Fail()
	}
}
