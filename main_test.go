package main

import (
	"os"
	"testing"
)

func TestKubePath(t *testing.T) {
	_, err := os.Stat(kubePath())
	if err != nil {
		t.Errorf("kubePath does not exist")
	}
	ConfigPathExists()
}

//func TestConfig(t *testing.T) {
//	ConfigPathExists()
//	confirm(filepath.FromSlash(kubepath() + "/configs/config"))
//}

func TestReadConfig(t *testing.T) {
	loadConfigs()
}
