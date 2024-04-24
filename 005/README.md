# 005

Since I'm blocked on ring buffers with what looks like a libxdp bug, I'm going to move ahead with testing the performance of using bpf hash maps to store player state.

I'm going to spin up 16 threads, and pin each to CPUs [0,15], then I'm going to simulate 50k players evenly distributed across these CPUs (eg. 3125 players per-CPU).

Each player simulation step will read the player state from the kernel memory, then advance the player state forward via dt, and stors the updated player state back into the bpf hash map.

If the player maps are a bottleneck, that should show up here. It's possible they are, since each read and write to the map needs to be transferred between kernel and userspace memory, presumably via a syscall.

Initially, I'm going to keep the player simulation very light. It will just touch all 1200 bytes of the player state (read and write), before committing it back to the map.

I'm not sure if I'll need to expand the number of cores dedicated to player simulation or not. I have a 64 core thread ripper running Linux, and if I need to go up to 64 cores in order to distribute the load of 50k players @ 100HZ I should be able to.

# Results

I can do 345,541 player updates per-second on 16 cpus on my bare metal linux box.

This means I can do 21k player updates per-cpu on bare metal.

Given that each player does 100 updates per-second, for 50k players, I need 5,000,000 player updates per-second.

5,000,000 updates per-second would require 238 cpus.
