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

func main() {

	if len(os.Args) != 2 {
		fmt.Printf( "\nusage: go run worker <cpu_index>\n\n")
		os.Exit(0)
	}

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	cpu, err :=	strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("error: could not read cpu index\n")
		os.Exit(1)
	}

	if runtime.GOOS == "linux" {
		pid := os.Getpid()
		cmd := exec.Command("taskset", "-pc", fmt.Sprintf("%d", cpu), fmt.Sprintf("%d", pid))
		if err := cmd.Run(); err != nil {
			fmt.Printf("error: could not pin process to cpu %d: %v\n", cpu, err)
			os.Exit(1)
		}	
	}

	runtime.GOMAXPROCS(1)

	// get player state map for our CPU

	player_state_outer, err := ebpf.LoadPinnedMap("/sys/fs/bpf/player_state_map", nil)
	if err != nil {
		fmt.Printf("error: could not get player state map: %v\n", err)
		os.Exit(1)
	}
	defer player_state_outer.Close()

	fmt.Printf("map %s\n", player_state_outer.String())

	/*
	var player_state_map *ebpf.Map
	err = player_state_map.Lookup(uint32(cpu), &player_state_map)
	if err != nil {
		fmt.Printf("error: could not lookup player state map for cpu %d: %v\n", cpu, err)
		os.Exit(1)
	}
	*/

	// get input buffer map for our CPU

	input_buffer_outer, err := ebpf.LoadPinnedMap("/sys/fs/bpf/input_buffer_map", nil)
	if err != nil {
		fmt.Printf("error: could not get input buffer map: %v\n", err)
		os.Exit(1)
	}
	defer input_buffer_outer.Close()

	var input_buffer_inner *ebpf.Map
	err = input_buffer_outer.Lookup(uint32(cpu), &input_buffer_inner)
	if err != nil {
		fmt.Printf("error: could not lookup input buffer for cpu %d: %v\n", cpu, err)
		os.Exit(1)
	}

	// create input ring buffer

	input_buffer, err := ringbuf.NewReader(input_buffer_inner)

	// poll ring buffer to read inputs

	go func() {
		for {
			record, err := input_buffer.Read()
			if err != nil {
				fmt.Printf("error: failed to read from ring buffer: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("process event (%d bytes)\n", len(record.RawSample))
			_ = record
			// _ = player_state_map
		}
	}()

	<- termChan
}
