package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"
)

func main() {
	err := runCli()
	if err != nil {
		log.Fatal(err)
	}
}

func runCli() error {
	timeoutPtr := flag.IntP("timeout", "t", 1, "timeout before restarting after program crashed (min: 1)")
	flag.Parse()

	if timeoutPtr == nil {
		return nil
	}

	index, err := findCommandIndex(os.Args[1:])
	if err != nil {
		return err
	}

	timeout := *timeoutPtr
	if timeout < 1 {
		timeout = 1
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	commandParts := os.Args[index+1:]
	cmd := commandParts[0]
	args := commandParts[1:]

	for {
		res := make(chan error, 1)
		go func() {
			log.Printf("Starting command: %s", strings.Join(commandParts, " "))
			command := exec.Command(cmd, args...)

			command.Stdout = os.Stdout
			command.Stderr = os.Stderr

			res <- command.Run()
		}()

		select {
		case err := <-res:
			if err != nil {
				log.Printf("Command exited with error: %v. Restarting...\n", err)
			} else {
				log.Println("Command completed succesffully. Restarting...")
			}

			time.Sleep(time.Duration(timeout) * time.Second)

		case sig := <-sigChan:
			log.Printf("Received signal: %v. Exiting...", sig)
			return nil
		}
	}
}

func findCommandIndex(args []string) (int, error) {
	found := false
	for index, arg := range args {
		if found {
			return index, nil
		}

		if arg != "--" {
			continue
		}

		found = true
	}

	return -1, errors.New("could not find command. Did you perhaps forget to seperate your command via --?")
}
