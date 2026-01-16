package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"strings"

	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

type Item struct {
	Name string
}

type information struct {
	Clusters []struct {
		Name    string `yaml:"name"`
		Cluster struct {
			Server string `yaml:"server"`
		} `yaml:"cluster"`
	} `yaml:"clusters"`
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func kubepath() string {

	return filepath.Join(UserHomeDir(), ".kube")
}

var configs = []string{}
var path = kubepath()

func loadConfigs() {
	entries, err := os.ReadDir(path + "/configs")
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		file, err := os.ReadFile(path + "/configs/" + entry.Name())
		if err != nil {
			log.Fatal(err)
		}
		var config information
		err = yaml.Unmarshal(file, &config)
		if err != nil {
			log.Fatal(err)
		}
		configs = append(configs, entry.Name()[strings.Index(entry.Name(), ".")+1:]+" IP: "+config.Clusters[0].Cluster.Server)

	}
}
func confirm(name string) {
	//fmt.Println("Navn:" + name)
	var source = path + "/configs/config." + name
	var dest = path + "/config"

	bytesRead, err := os.ReadFile(source)
	if err != nil {

	}
	os.Remove(dest)
	os.WriteFile(dest, bytesRead, 0644)
}

func main() {
	//fmt.Println("Hello World")

	app := tview.NewApplication()
	loadConfigs()
	configList := tview.NewList()
	configList.SetBorder(true).SetTitle("Configuration")

	infoData := tview.NewTextView()
	infoData.SetBorder(true).SetTitle("Information")

	refreshConfigs := func() {
		configList.Clear()
		if len(configs) == 0 {

		} else {
			for _, config := range configs {
				configList.AddItem(config, "", 0, func() {
					app.Stop()
					confirm(config)
				})
			}
		}

	}

	flex := tview.NewFlex().
		AddItem(configList, 0, 1, true).
		AddItem(infoData, 0, 1, false)

	refreshConfigs()
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
