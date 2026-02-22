package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var configs = []string{}
var path = kubepath()
var config []ConfigInformation
var infos []info
var app = tview.NewApplication()
var configList = tview.NewList()
var infoData = tview.NewTextView()

var flex = tview.NewFlex().
	AddItem(configList, 0, 1, true).
	AddItem(infoData, 0, 1, false)

type info struct {
	Active bool
	Name   string
	User   string
	port   string
	ip     string
	ping   bool
	path   string
	nodes  int
	pods   int
	status string
	test   string
}

type ConfigInformation struct {
	Clusters []struct {
		Name    string `yaml:"name"`
		Cluster struct {
			Server string `yaml:"server"`
		} `yaml:"cluster"`
	} `yaml:"clusters"`

	Contexts []struct {
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
}

func InfoDataDisplay(data info) string {
	var statusIcon = "🔴"
	if data.ping {
		statusIcon = "🟢"
	}
	var information = "Name: " + data.Name +
		"\n\nUser:.. " + data.User +
		"\nIP:.... " + data.ip +
		"\nPort:.. " + data.port +
		"\nPing:.. " + strings.ToUpper(strconv.FormatBool(data.ping)) + statusIcon +
		"\nPath... " + data.path[strings.LastIndex(data.path, "/")+1:]
	if data.ping {
		information = information + "\nNodes.. " + strconv.Itoa(data.nodes) +
			"\nPods... " + strconv.Itoa(data.pods)
	}
	if data.status != "" {
		information = information + "\n\nStatus: " + data.status
	}
	if len(data.test) > 0 {
		information = information + "\n\n\nTests.. " + data.test
	}
	return information
}

func Testconnection(ip string, port string) bool {
	address := net.JoinHostPort(ip, port)
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		//log.Fatal(err)
		return false
	}
	defer conn.Close()
	return true
}

func Move(data ConfigInformation) info {
	var inn info
	inn.Name = data.Clusters[0].Name
	inn.User = data.Clusters[0].Cluster.Server
	inn.port = data.Clusters[0].Cluster.Server[strings.LastIndex(data.Clusters[0].Cluster.Server, ":")+1:]
	inn.ip = data.Clusters[0].Cluster.Server[strings.Index(data.Clusters[0].Cluster.Server, "/")+2 : strings.LastIndex(data.Clusters[0].Cluster.Server, ":")]
	inn.ping = false
	inn.path = data.Clusters[0].Cluster.Server
	inn.nodes = 0
	inn.pods = 0
	inn.status = "Getting info from cluster...."
	return inn
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return filepath.FromSlash(home)
	}
	return filepath.FromSlash(os.Getenv("HOME"))
}

func kubepath() string {
	return filepath.Join(UserHomeDir(), ".kube")
}

func loadConfigs() {
	entries, err := os.ReadDir(path + "/configs")
	if err != nil {
		log.Fatal(err)
	}
	for i, entry := range entries {

		file, err := os.ReadFile(path + "/configs/" + entry.Name())
		if err != nil {
			log.Fatal(err)
		}

		var newconfig ConfigInformation
		err = yaml.Unmarshal(file, &newconfig)
		if err != nil {
			log.Fatal(err)
		}
		config = append(config, newconfig)
		infos = append(infos, Move(newconfig))
		infos[i].path = path + "/configs/" + entry.Name()
		if is_current(file) {
			configs = append(configs, config[i].Clusters[0].Name+" - "+"ACTIVE")
		} else {
			configs = append(configs, config[i].Clusters[0].Name)
		}

	}
}

func GetInfo() {

	for i, _ := range infos {
		infos[i].ping = Testconnection(infos[i].ip, infos[i].port)
		if infos[i].ping {
			infos[i].status = ""
			kubeconfig, err := clientcmd.BuildConfigFromFlags("", infos[i].path)
			if err != nil {
				log.Fatal(err)
			}
			clientset, err := kubernetes.NewForConfig(kubeconfig)
			if err != nil {
				log.Fatal(err)
			}
			nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			numNodes := len(nodes.Items)
			infos[i].nodes = numNodes
			pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
			infos[i].pods = len(pods.Items)
		} else {
			infos[i].status = "Offline"
		}
	}
	app.QueueUpdateDraw(func() {
		refreshConfigs()
	})

}

func is_current(file []byte) bool {
	current, err := os.ReadFile(path + "/config")
	if err != nil {
		log.Fatal(err)
	}
	if bytes.Equal(file, current) {
		return true
	}
	return false
}

func confirm(name string) {
	fmt.Println("name: " + name)
	var source = name
	var dest = path + "/config"
	bytesRead, err := os.ReadFile(source)
	if err != nil {

	}
	os.Remove(dest)
	os.WriteFile(dest, bytesRead, 0644)
}

func refreshConfigs() {
	configList.Clear()
	if len(configs) == 0 {

	} else {
		for i, config := range configs {
			configList.AddItem(config, "", 0, func() {
				app.Stop()
				confirm(filepath.FromSlash(infos[i].path))
			})
		}
	}

}

func main() {

	loadConfigs()
	go GetInfo()

	configList.SetBorder(true).SetTitle("Configuration")

	configList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		infoData.SetText(InfoDataDisplay(infos[index]))
	})

	infoData.SetBorder(true).SetTitle("Info").SetTitleAlign(tview.AlignLeft)

	refreshConfigs()
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
