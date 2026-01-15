package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rivo/tview"
)

type Item struct {
	Name string
}

var configs = []string{}
var path = "/mnt/c/Users/chris/.kube"

func loadConfigs() {
	entries, err := os.ReadDir(path + "/configs")
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Println(entry.Name())
		configs = append(configs, entry.Name()[strings.Index(entry.Name(), ".")+1:])

	}
}
func confirm(name string) {
	fmt.Println("Navn:" + name)
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
		AddItem(configList, 0, 1, true)

	refreshConfigs()
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
