package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Printf( "\nusage: go run server <cpu_index>\n\n")
		os.Exit(0)
	}

	cpu, err :=	strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf( "\nerror: could not read cpu index\n\n")
		os.Exit(1)
	}

	if runtime.GOOS == "linux" {
		pid := os.Getpid()
		cmd := exec.Command("taskset", "-pc", fmt.Sprintf("%d", cpu), fmt.Sprintf("%d", pid))
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}	
	}

	runtime.GOMAXPROCS(1)

	fmt.Printf("server running on cpu %d\n", cpu)

	// todo: get ring buffer outer map

	// todo: get ring buffer for cpu #

	// todo: consume ring buffer
}
