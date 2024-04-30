package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/cilium/ebpf"
)

const MaxCPUs = 16

func main() {

	if len(os.Args) != 2 {
		fmt.Printf( "\nusage: go run server <cpu_index>\n\n")
		os.Exit(0)
	}

	cpu, err :=	strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("\nerror: could not read cpu index\n\n")
		os.Exit(1)
	}

	if runtime.GOOS == "linux" {
		pid := os.Getpid()
		cmd := exec.Command("taskset", "-pc", fmt.Sprintf("%d", cpu), fmt.Sprintf("%d", pid))
		if err := cmd.Run(); err != nil {
			fmt.Printf("\nerror: could not pin process to cpu %d: %v\n\n", cpu, err)
			os.Exit(1)
		}	
	}

	runtime.GOMAXPROCS(1)

	fmt.Printf("golang server running on cpu %d/%d\n", cpu, MaxCPUs)

	input_buffer_outer, err := ebpf.LoadPinnedMap("/sys/fs/bpf/input_buffer_map", nil)
	if err != nil {
		fmt.Printf("\nerror: could not get input buffer map: %v\n\n", err)
		os.Exit(1)
	}
	defer input_buffer_outer.Close()

	var input_buffer_inner *ebpf.Map
	err := input_buffer_outer.Lookup(cpu, &input_buffer_inner)
	if err != nil {
		fmt.Printf("\nerror: could not lookup input buffer for cpu %d\n\n", cpu)
		os.Exit(1)
	}

	// todo: create ring buffer

	// todo: consume ring buffer events
}
