package main

import (
	"bufio"
	"fmt"
	"github.com/chonthu/rtail/easyssh"
	"io"
	"os"
	"sync"
)

func main() {

	servers := os.Args[1:]

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
			fmt.Println(s)
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

func Connect(server string) {

	fmt.Println("Connecting to " + server)

	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		User:   "root",
		Server: server,
		// Optional key or Password without either we try to contact your agent SOCKET
		Key:  "/.ssh/id_dsa",
		Port: "22",
	}

	// Call Run method with command you want to run on remote server.
	channel, done, err := Stream(ssh, "tail -f /var/log/httpd/error_log")
	if err != nil {
		fmt.Errorf("Stream failed: %s", err)
	}
	stillGoing := true
	for stillGoing {
		select {
		case <-done:
			stillGoing = false
		case line := <-channel:
			fmt.Printf("[%s] %s\n", server, line)
		}
	}
}

// Stream returns one channel that combines the stdout and stderr of the command
// as it is run on the remote machine, and another that sends true when the
// command is done. The sessions and channels will then be closed.
func Stream(ssh_conf *easyssh.MakeConfig, command string) (output chan string, done chan bool, err error) {
	// connect to remote host
	session, err := ssh_conf.Connect()
	if err != nil {
		return output, done, err
	}
	// connect to both outputs (they are of type io.Reader)
	outReader, err := session.StdoutPipe()
	if err != nil {
		return output, done, err
	}
	errReader, err := session.StderrPipe()
	if err != nil {
		return output, done, err
	}
	// combine outputs, create a line-by-line scanner
	outputReader := io.MultiReader(outReader, errReader)
	err = session.Start(command)
	scanner := bufio.NewScanner(outputReader)
	// continuously send the command's output over the channel
	outputChan := make(chan string)
	done = make(chan bool)
	go func(scanner *bufio.Scanner, out chan string, done chan bool) {
		defer close(outputChan)
		defer close(done)
		for scanner.Scan() {
			outputChan <- scanner.Text()
		}
		// close all of our open resources
		done <- true
		session.Close()
	}(scanner, outputChan, done)
	return outputChan, done, err
}
