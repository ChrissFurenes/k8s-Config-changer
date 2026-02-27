package main

import (
	"os"
	"testing"
)

func TestKubePath(t *testing.T) {
	_, err := os.Stat(kubepath())
	if err != nil {
		t.Errorf("kubepath does not exist")
	}
}

//func TestConfig(t *testing.T) {
//	ConfigPathExists()
//	confirm(filepath.FromSlash(kubepath() + "/configs/config"))
//}

func TestReadConfig(t *testing.T) {
	loadConfigs()
}
