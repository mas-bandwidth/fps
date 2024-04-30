package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"os/signal"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
)

const MaxCPUs = 16

// todo: signal handler CTRL-C etc

func main() {

	if len(os.Args) != 2 {
		fmt.Printf( "\nusage: go run worker <cpu_index>\n\n")
		os.Exit(0)
	}

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

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

	fmt.Printf("golang worker running on cpu %d/%d\n", cpu, MaxCPUs)

	input_buffer_outer, err := ebpf.LoadPinnedMap("/sys/fs/bpf/input_buffer_map", nil)
	if err != nil {
		fmt.Printf("\nerror: could not get input buffer map: %v\n\n", err)
		os.Exit(1)
	}
	defer input_buffer_outer.Close()

	var input_buffer_inner *ebpf.Map
	err = input_buffer_outer.Lookup(uint32(cpu), &input_buffer_inner)
	if err != nil {
		fmt.Printf("\nerror: could not lookup input buffer for cpu %d: %v\n\n", cpu, err)
		os.Exit(1)
	}

	input_buffer, err := ringbuf.NewReader(input_buffer_inner)

	go func() {
		for {
			record, err := input_buffer.Read()
			if err != nil {
				fmt.Printf("\nerror: failed to read from ring buffer: %v\n\n", err)
				os.Exit(1)
			}
			fmt.Printf("process event (%d bytes)\n")
			_ = record
		}
	}()

	<- termChan
}
