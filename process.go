package main

import (
    "os"
    "os/exec"
    "path/filepath"
)

type Process struct {
	Command string
	Output  chan string
}

func (self *Process) Run(processes ...Process) {
    currentDirectory, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
        Error.Println(err)
    }
    Info.Printf("Current directory: %s", currentDirectory)
    Info.Printf("Command: bash -c \"%s\"", self.Command)
    out, err := exec.Command("bash", "--login", "-c", self.Command).Output()
    if err != nil {
        Error.Println(err)
    }
    self.Output <- string(out)

    if len(processes) > 0 {
        process := processes[len(processes)-1]
        processes = processes[:len(processes)-1]
        Info.Println(process)
        process.Run()
    }
}

func Collect(c chan string) {
	for {
		msg := <-c
		Info.Printf("Output:\n%s", msg)
	}
}
