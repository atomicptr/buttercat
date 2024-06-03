package cli

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

func Run() error {
	timeout := flag.IntP("timeout", "t", 1, "timeout before restarting after program crashed")
	flag.Parse()

	if timeout == nil {
		return nil
	}

	index, err := findCommandIndex(os.Args[1:])
	if err != nil {
		return err
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
			res <- command.Run()
		}()

		select {
		case err := <-res:
			if err != nil {
				log.Printf("Command exited with error: %v. Restarting...\n", err)
			} else {
				log.Println("Command completed succesffully. Restarting...")
			}

			time.Sleep(time.Duration(*timeout) * time.Second)
		case sig := <-sigChan:
			log.Printf("Received signal: %v. Exiting...", sig)
			return nil
		}
	}
}

func findCommandIndex(args []string) (int, error) {
	for index, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return index, nil
	}

	return -1, errors.New("could not find command")
}
