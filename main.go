package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/rivo/tview"
)

type Item struct {
	Name string
}

type info struct {
	Name     string
	UserName string
	ip       string
	ping     bool
	path     string
	nodes    int
	pods     int
}

type configinformation struct {
	Clusters []struct {
		Name    string `yaml:"name"`
		Cluster struct {
			Server string `yaml:"server"`
		} `yaml:"cluster"`
	} `yaml:"clusters"`
}

func InfoDataDisplay(data info) string {
	var information = "Name: " + data.Name +
		"\n\nUser: " + data.UserName +
		"\nIP: " + data.ip +
		"\nPing: " + strconv.FormatBool(data.ping) + "\n"
	return information
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
var infos = []info{}

func loadConfigs() {
	entries, err := os.ReadDir(path + "/configs")
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		currInfo := info{}
		infos = append(infos, currInfo)
		configs = append(configs, entry.Name()[strings.Index(entry.Name(), ".")+1:])

	}
}
func confirm(name string) {

	var source = path + "/configs/config." + name
	var dest = path + "/config"

	bytesRead, err := os.ReadFile(source)
	if err != nil {

	}
	os.Remove(dest)
	os.WriteFile(dest, bytesRead, 0644)
}

func main() {

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

	infoData := tview.NewTextView()

	var hh info
	hh.ip = "127.0.0.1"
	hh.ping = true
	hh.path = path
	hh.nodes = runtime.NumCPU()
	hh.pods = runtime.NumCPU()
	hh.UserName = "per"

	configList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		infoData.SetText(InfoDataDisplay(infos[index]))
	})

	infoData.SetBorder(true).SetTitle("Info").SetTitleAlign(tview.AlignLeft)

	flex := tview.NewFlex().
		AddItem(configList, 0, 1, true).
		AddItem(infoData, 0, 1, false)

	refreshConfigs()
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
