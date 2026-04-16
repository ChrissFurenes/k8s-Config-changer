package main

import (
	"bytes"
	"context"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	//"github.com/ChrissFurenes/k8s-Config-changer/cmd"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var configs []string
var path = kubePath()
var newFolderPath = "/"
var PrevFolder = []string{"config"}

var config []ConfigInformation
var infos []Info
var app = tview.NewApplication()
var configList = tview.NewList().ShowSecondaryText(false)
var infoData = tview.NewTextView()

// var commandList = tview.NewTextView().SetText("[F5] Refresh | [F10] Settings | [F2] Open").SetTextAlign(tview.AlignCenter)
var commandList = tview.NewTextView().SetText("[F5] Refresh").SetTextAlign(tview.AlignCenter)

var grid = tview.NewGrid().
	SetRows(-1, 25).
	SetColumns(-1, -1).
	SetBorders(false).
	AddItem(configList, 0, 0, 5, 1, 0, 0, true).
	AddItem(infoData, 0, 1, 5, 1, 0, 0, false).
	AddItem(commandList, 5, 0, 1, 2, 1, 0, false)

type Info struct {
	Active        bool
	Name          string
	User          string
	port          string
	ip            string
	ping          bool
	path          string
	nodes         int
	pods          int
	status        string
	test          string
	folder        bool
	filesInFolder int
	prevFolder    string
	isBack        bool
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

func InfoDataDisplay(data Info) string {
	var information = ""
	if !data.folder && !data.isBack {
		var statusIcon = "🔴"
		var color = "[red]"
		if data.ping {
			statusIcon = "🟢"
			color = "[green]"
		}
		information = "Name:.. " + data.Name +
			"\n\nUser:.. " + data.User +
			"\nIP:.... " + data.ip +
			"\nPort:.. " + data.port +
			"\nPing:.. " + color + strings.ToUpper(strconv.FormatBool(data.ping)) + "[::-] [white]" + statusIcon +
			"\nPath:.. " + data.path[strings.LastIndex(data.path, "/")+1:]
		if data.ping {
			information = information + "\nNodes:. " + strconv.Itoa(data.nodes) +
				"\nPods:.. " + strconv.Itoa(data.pods)
		}
		if data.status != "" {
			information = information + "\n\nStatus: " + data.status
		}
		if len(data.test) > 0 {
			information = information + "\n\n\nTests:. " + data.test
		}
		return information

	}

	information = ReadFolderInfo(path + "/configs" + newFolderPath + data.path)
	return information
}

func TestConnection(ip string, port string) bool {
	address := net.JoinHostPort(ip, port)
	conn, err := net.DialTimeout("tcp", address, 2000*time.Millisecond)
	if err != nil {
		return false
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)
	return true
}

func Move(data ConfigInformation) Info {
	var inn Info
	inn.Active = false
	inn.Name = data.Clusters[0].Name
	inn.User = data.Contexts[0].Context.User
	inn.port = data.Clusters[0].Cluster.Server[strings.LastIndex(data.Clusters[0].Cluster.Server, ":")+1:]
	inn.ip = data.Clusters[0].Cluster.Server[strings.Index(data.Clusters[0].Cluster.Server, "/")+2 : strings.LastIndex(data.Clusters[0].Cluster.Server, ":")]
	inn.ping = false
	inn.path = ""
	inn.nodes = 0
	inn.pods = 0
	inn.status = "[yellow]Getting info from cluster....[::-]"
	inn.test = ""
	inn.folder = false
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

func kubePath() string {

	return filepath.Join(UserHomeDir(), ".kube")
}

func loadConfigs() {
	entries, err := os.ReadDir(path + "/configs" + newFolderPath)
	if err != nil {
		log.Fatal(err)
	}

	var num = 0
	if len(newFolderPath) > 1 {
		var backInfo Info
		backInfo.isBack = true
		backInfo.prevFolder = newFolderPath[strings.LastIndex(newFolderPath[0:len(newFolderPath)-1], filepath.FromSlash("/"))+1 : len(newFolderPath)-1]
		backInfo.folder = false
		configs = append(configs, " << Back to folder: "+PrevFolder[len(PrevFolder)-1])

		infos = append(infos, backInfo)
		num++
	}
	if len(entries) == 0 {
		return
	}
	for _, entry := range entries {
		var newConfig ConfigInformation
		if entry.IsDir() {
			configs = append(configs, "📁 "+entry.Name())
			var newFolder Info
			newFolder.folder = true
			newFolder.path = entry.Name()
			infos = append(infos, newFolder)
			config = append(config, ConfigInformation{})
			num++
			continue
		}
		file, err := os.ReadFile(path + "/configs" + newFolderPath + entry.Name())
		if err != nil {
			log.Fatal(err)
		}

		err = yaml.Unmarshal(file, &newConfig)
		if err != nil {
			log.Fatal(err)
		}
		config = append(config, newConfig)
		infos = append(infos, Move(newConfig))
		infos[num].path = filepath.FromSlash(path + "/configs" + newFolderPath + entry.Name())

		if !infos[0].isBack {
			if IsCurrent(file) {
				configs = append(configs, "☸  "+config[num].Clusters[0].Name+" - "+"[green]ACTIVE[::-]")
				infos[num].Active = true
			} else {
				configs = append(configs, "☸  "+config[num].Clusters[0].Name)
			}
		} else {
			if IsCurrent(file) {
				configs = append(configs, "☸  "+config[num-1].Clusters[0].Name+" - "+"[green]ACTIVE[::-]")
				infos[num].Active = true
			} else {
				configs = append(configs, "☸  "+config[num-1].Clusters[0].Name)
			}
		}
		num++
	}

}

func GetInfo() {
	var inn = infos
	var folder = newFolderPath
	for i := range inn {
		if !(inn[i].folder || inn[i].isBack) {
			inn[i].ping = TestConnection(inn[i].ip, inn[i].port)
			if inn[i].ping {
				inn[i].status = ""
				kubeconfig, err := clientcmd.BuildConfigFromFlags("", inn[i].path)
				if err != nil {
					log.Fatal(err)
				}
				clientSet, err := kubernetes.NewForConfig(kubeconfig)
				if err != nil {
					log.Fatal(err)
				}
				nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					log.Fatal(err)
				}
				numNodes := len(nodes.Items)
				inn[i].nodes = numNodes
				pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					log.Fatal(err)
				}
				inn[i].pods = len(pods.Items)
			} else {
				inn[i].status = "[red]Offline[::-]"
			}
		}
		if folder == newFolderPath {
			infos = inn
			app.QueueUpdateDraw(func() {
				cu := configList.GetCurrentItem()
				refreshConfigs()
				configList.SetCurrentItem(cu)
			})
		} else {
			break
		}
	}

}
func ReadFolderInfo(folderPath string) string {
	files, err := os.ReadDir(filepath.FromSlash(folderPath))
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
	}
	return "This is " + filepath.FromSlash(folderPath)
}

func IsCurrent(file []byte) bool {
	current, err := os.ReadFile(filepath.FromSlash(path + "/config"))
	if err != nil {
		log.Fatal(err)
	}
	if bytes.Equal(file, current) {
		return true
	}
	return false
}

func confirm(name string) {
	var source = filepath.FromSlash(name)
	var dest = filepath.FromSlash(path + "/config")
	bytesRead, err := os.ReadFile(source)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove(dest)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(dest, bytesRead, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func refreshConfigs() {
	var pos = configList.GetCurrentItem()
	configList.Clear()

	if len(configs) == 0 {
		return
	} else {

		for i, configEntry := range configs {
			configList.AddItem(configEntry, "", 0, func() {
				if !(infos[i].folder || infos[i].isBack) {
					app.Stop()
					confirm(filepath.FromSlash(infos[i].path))
				} else if infos[i].folder && !infos[i].isBack {
					newFolderPath = filepath.FromSlash(newFolderPath + infos[i].path + "/")
					if len(infos[0].prevFolder) != 0 {
						PrevFolder = append(PrevFolder, infos[0].prevFolder)
					}

				} else if infos[0].isBack {
					newFolderPath = newFolderPath[0 : strings.LastIndex(newFolderPath[0:len(newFolderPath)-1], filepath.FromSlash("/"))+1]
					if len(PrevFolder) > 1 {
						PrevFolder = PrevFolder[:len(PrevFolder)-1]
					}

				}

				configs = nil
				infos = nil
				config = nil
				loadConfigs()
				go GetInfo()
				refreshConfigs()
				if configList.GetItemCount() > 0 && len(newFolderPath) > 1 {
					configList.SetCurrentItem(1)
				} else {
					configList.SetCurrentItem(0)
				}
				//if !infos[0].isBack || (infos[0].isBack && !infos[0].folder) {
				//	configList.SetCurrentItem(0)
				//}
			})
		}
	}
	configList.SetCurrentItem(pos)
}

func ConfigPathExists() {
	_, err := os.Stat(filepath.FromSlash(path))
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(filepath.FromSlash(path + "/configs"))
	if err != nil {
		errors := os.MkdirAll(filepath.FromSlash(path+"/configs"), 0755)
		if errors != nil {
			log.Fatal(errors)
		}
		bytesRead, err := os.ReadFile(filepath.FromSlash(path + "/config"))
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(filepath.FromSlash(path+"/configs/config"), bytesRead, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	ConfigPathExists()
	loadConfigs()
	go GetInfo()

	configList.SetBorder(true).SetTitle("Configuration")
	infoData.SetDynamicColors(true)
	configList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		infoData.SetText(InfoDataDisplay(infos[index]))
	})

	infoData.SetBorder(true).SetTitle("Info").SetTitleAlign(tview.AlignCenter)

	refreshConfigs()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF5 {
			for i := range infos {
				infos[i].status = "[yellow]Getting info from cluster....[::-]"
			}
			go GetInfo()
			refreshConfigs()
		} else if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyEsc {
			if len(newFolderPath) > 1 {
				newFolderPath = newFolderPath[0 : strings.LastIndex(newFolderPath[0:len(newFolderPath)-1], filepath.FromSlash("/"))+1]
				if len(PrevFolder) > 1 {
					PrevFolder = PrevFolder[:len(PrevFolder)-1]
				}
				configs = nil
				infos = nil
				config = nil
				loadConfigs()
				go GetInfo()
				refreshConfigs()
				if configList.GetItemCount() > 0 && len(newFolderPath) > 1 {
					configList.SetCurrentItem(1)
				} else {
					configList.SetCurrentItem(0)
				}
			}
		}
		return event
	})
	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
