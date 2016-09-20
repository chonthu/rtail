package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/chonthu/ssh"
	"github.com/fatih/color"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	colors = []func(...interface{}) string{
		color.New(color.FgYellow).SprintFunc(),
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgGreen).SprintFunc(),
		color.New(color.FgBlue).SprintFunc(),
		color.New(color.FgMagenta).SprintFunc(),
	}

	configPath = []string{
		".rtail.yml",
		os.Getenv("HOME") + "/.rtail.yml",
	}

	config     = new(Config)
	configFile = kingpin.Flag("config", "path to config, different from default").String()
	identity   = kingpin.Flag("indetityFile", "path to the identity file").Short('i').Strings()

	Version string
)

const exampleConfig = `---
aliases:
	access_log: /var/log/httpd/access_log
	error_log: /var/log/httpd/error_log
commands:
	varnish: varnishlog
	varnish_url: varnishlog -g request | grep reqURL
	varnish_hit: varnishlog -q \"VCL_call eq 'HIT'\" -d
	varnish_miss: varnishlog -q \"VCL_call eq 'MISS'\" -d
	varnish_security: varnishlog | grep security.vcl
hosts:
	- google.web1
`

// Server struct
type Server struct {
	user string
	host string
	cmd  string
}

type serverList []Server

func (i *serverList) Set(value string) error {

	// Is username passed? if not use root
	user := "root"
	if strings.Contains(value, "@") {
		usernameSplit := strings.Split(value, "@")
		user = usernameSplit[0]
		value = usernameSplit[1]
	}

	fileToLog := "/var/log/httpd/error_log"
	cmdString := fmt.Sprintf("tail -f %v", fileToLog)

	// Is filename passed?
	if strings.Contains(value, "%") {
		fileSplit := strings.Split(value, "%")
		value = fileSplit[0]
		if strings.Contains(fileSplit[1], ":") {
			paramSplit := strings.Split(fileSplit[1], ":")
			cmdString = fmt.Sprintf(execShorcodes(paramSplit[0]), paramSplit[1:])
		} else {
			cmdString = execShorcodes(fileSplit[1])
		}

	} else if strings.Contains(value, ":") {
		fileSplit := strings.Split(value, ":")
		fileToLog = logFileShorcodes(fileSplit[1])
		value = fileSplit[0]
		cmdString = fmt.Sprintf("tail -f %v", fileToLog)
	}

	*i = append(*i, Server{user, value, cmdString})
	return nil
}

func (i *serverList) String() string {
	return ""
}

func (i *serverList) IsCumulative() bool {
	return true
}

// ServerList is a list of Servers
func ServerList(s kingpin.Settings) (target *[]Server) {
	target = new([]Server)
	s.SetValue((*serverList)(target))
	return
}

// Config is a struct of our config options
type Config struct {
	Aliases  map[string]string
	Commands map[string]string
	Hosts    []string
}

func listHosts() []string {
	return config.Hosts
}

func initConfig(file *os.File) {
	b, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Invalid syntax in config file")
		os.Exit(1)
	}

	err = yaml.Unmarshal(b, &config)
	if err != nil {
		fmt.Println("Invalid syntax in config file")
		os.Exit(1)
	}
}

func main() {
	// Check if config is passed
	if *configFile != "" {
		configPath = append([]string{*configFile}, configPath...)
	}

	for _, v := range configPath {
		file, err := os.Open(v) // For read access.
		if err == nil {
			defer file.Close()
			initConfig(file)
			break
		}
	}

	// boostrap commandline cli
	servers := ServerList(kingpin.Arg("servers", "the servers to parse").Required().HintAction(listHosts))
	kingpin.Version(Version).Author("Nithin Meppurathu")
	kingpin.CommandLine.Help = "A log parser and command execution multiplexer"
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.CommandLine.VersionFlag.Short('v')
	kingpin.Parse()

	switch os.Args[1] {
	case "init":

		if _, err := os.Stat(os.Getenv("HOME") + "/.rtail"); os.IsNotExist(err) {
			fmt.Println("creating config file in home directory")
			f, err := os.Create(os.Getenv("HOME") + "/.rtail")
			if err != nil {
				fmt.Println(err)
			} else {
				f.Write([]byte(exampleConfig))
				f.Close()
			}
		} else {
			fmt.Println(colors[1](".rtail config file already exists in home directory"))
		}
		break
	default:
		srv, err := rangeSplitServers(*servers)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		jobs := make(chan int, len(srv))

		var wg sync.WaitGroup

		// Spawn example workers
		for _, s := range srv {
			wg.Add(1)
			go func(s Server, jobs <-chan int) {
				defer wg.Done()
				Connect(&s)
			}(s, jobs)
		}

		// Create example messages
		for i := 0; i < len(srv); i++ {
			jobs <- i
		}

		close(jobs)
		wg.Wait()
	}
}

// Prases provided server to check for expansions
func rangeSplitServers(servers []Server) ([]Server, error) {
	var out []Server

	for _, v := range servers {

		if 0 == strings.Index(v.host, "-") {
			continue
		}

		if strings.Contains(v.host, "[") {
			re, _ := regexp.Compile(`\[(\d+)-(\d+)\]`)
			res := re.FindAllStringSubmatch(v.host, -1)

			if len(res) == 0 {
				return out, fmt.Errorf("Invalid server regex")
			}

			if len(res[0]) == 3 {
				min, _ := strconv.Atoi(res[0][1])
				max, _ := strconv.Atoi(res[0][2])
				for num := min; num <= max; num++ {
					cp := v
					cp.host = strings.Replace(v.host, res[0][0], strconv.Itoa(num), -1)
					out = append(out, cp)
				}
			} else if len(res[0]) == 2 {
				cp := v
				cp.host = strings.Replace(v.host, res[0][0], res[0][1], -1)
				out = append(out, cp)
			} else {
				return out, fmt.Errorf("Invalid server grouping")
			}

			continue
		}

		out = append(out, v)
	}

	return out, nil
}

func logFileShorcodes(name string) string {
	if _, ok := config.Aliases[name]; ok {
		return config.Aliases[name]
	}

	return name
}

func execShorcodes(name string) string {
	if _, ok := config.Commands[name]; ok {
		return config.Commands[name]
	}
	return name
}

// Connect trying t connect to the server passed
func Connect(server *Server) {

	// Use a random color from the color list
	c := colors[rand.Intn(len(colors))]
	fmt.Printf("[%v] trying to connect as %v \n", c(server.host), server.user)

	keys := []string{os.Getenv("HOME") + "/.ssh/id_rsa", os.Getenv("HOME") + "/.ssh/id_dsa"}

	if len(*identity) > 0 {
		keys = *identity
	}

	// Create MakeConfig instance with remote username, server address and path to private key.
	s := &ssh.MakeConfig{
		User:   server.user,
		Server: server.host,
		// Optional key or Password without either we try to contact your agent SOCKET
		Key:  keys,
		Port: "22",
	}

	// Call Run method with command you want to run on remote server.
	fmt.Printf("[%v] runinng command: %v \n", c(server.host), server.cmd)
	channel, done, err := ssh.Stream(s, server.cmd)
	if err != nil {
		fmt.Println(fmt.Errorf("[%v] stream failed: %s", c(server.host), err))
		return
	}
	stillGoing := true
	for stillGoing {
		select {
		case <-done:
			stillGoing = false
		case line := <-channel:
			fmt.Printf("[%s] %s\n", c(server.host), line)
		}
	}
}
