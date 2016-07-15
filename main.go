package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/chonthu/ssh"
	"github.com/fatih/color"
)

var (
	colors []func(...interface{}) string = []func(...interface{}) string{
		color.New(color.FgYellow).SprintFunc(),
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgGreen).SprintFunc(),
		color.New(color.FgBlue).SprintFunc(),
		color.New(color.FgMagenta).SprintFunc(),
	}
)

func main() {

	servers := parseServers(os.Args[1:])

	if len(servers) < 1 {
		fmt.Println("No servers passed")
		os.Exit(1)
	}

	jobs := make(chan int, len(servers))

	var wg sync.WaitGroup

	// Spawn example workers
	for _, s := range servers {
		wg.Add(1)
		go func(s string, jobs <-chan int) {
			defer wg.Done()
			Connect(s)
		}(s, jobs)
	}

	// Create example messages
	for i := 0; i < len(servers); i++ {
		jobs <- i
	}

	close(jobs)
	wg.Wait()
}

// Prases provided string into servers
func parseServers(servers []string) []string {
	var out []string

	for _, v := range servers {

		if 0 == strings.Index(v, "-") {
			continue
		}

		if strings.Contains(v, "[") {
			re, _ := regexp.Compile(`\[(\d+)-(\d+)\]`)
			res := re.FindAllStringSubmatch(v, -1)

			if len(res) == 0 {
				panic("Invalid server regex")
			}

			if len(res[0]) == 3 {
				min, _ := strconv.Atoi(res[0][1])
				max, _ := strconv.Atoi(res[0][2])
				for num := min; num <= max; num++ {
					out = append(out, strings.Replace(v, res[0][0], strconv.Itoa(num), -1))
				}
			} else if len(res[0]) == 2 {
				out = append(out, strings.Replace(v, res[0][0], res[0][1], -1))
			} else {
				panic("Invalid server grouping")
			}

			continue
		}

		out = append(out, v)
	}

	return out
}

func logFileShorcodes(name string) string {
	switch name {
	case "access_log":
		return "/var/log/httpd/access_log"
	case "error_log":
		return "/var/log/httpd/error_log"
	default:
		return name
	}
}

func execShorcodes(name string) string {
	switch name {
	case "varnish_url":
		return "varnishlog -g request | grep reqURL"
	case "varnish_hit":
		return "varnishlog -q \"VCL_call eq 'HIT'\" -d"
	case "varnish_miss":
		return "varnishlog -q \"VCL_call eq 'MISS'\" -d"
	case "varnish":
		return "varnishlog"
	default:
		return name
	}
}

func Connect(server string) {
	user := "root"
	// Is username passed?
	if strings.Contains(server, "@") {
		usernameSplit := strings.Split(server, "@")
		user = usernameSplit[0]
		server = usernameSplit[1]
	}

	fileToLog := "/var/log/httpd/error_log"
	cmdString := fmt.Sprintf("tail %v", fileToLog)
	// Is filename passed?
	if strings.Contains(server, "%") {
		fileSplit := strings.Split(server, "%")
		server = fileSplit[0]
		if strings.Contains(fileSplit[1], ":") {
			paramSplit := strings.Split(fileSplit[1], ":")
			cmdString = fmt.Sprintf(execShorcodes(paramSplit[0]), paramSplit[1:])
		} else {
			cmdString = execShorcodes(fileSplit[1])
		}

	} else if strings.Contains(server, ":") {
		fileSplit := strings.Split(server, ":")
		fileToLog = logFileShorcodes(fileSplit[1])
		server = fileSplit[0]
	}

	c := colors[rand.Intn(len(colors))]
	fmt.Printf("Connecting to %v as %v \n", c(server), user)

	// Create MakeConfig instance with remote username, server address and path to private key.
	s := &ssh.MakeConfig{
		User:   user,
		Server: server,
		// Optional key or Password without either we try to contact your agent SOCKET
		Key:  []string{"/.ssh/id_rsa", "/.ssh/id_dsa"},
		Port: "22",
	}

	// Call Run method with command you want to run on remote server.
	fmt.Println("Running: ", cmdString)
	channel, done, err := ssh.Stream(s, cmdString)
	if err != nil {
		fmt.Errorf("Stream failed: %s", err)
	}
	stillGoing := true
	for stillGoing {
		select {
		case <-done:
			stillGoing = false
		case line := <-channel:
			fmt.Printf("[%s] %s\n", c(server), line)
		}
	}
}
