# 001

Let's see if we can process 1M player inputs in an FPS shooter style. 

Send inputs over UDP to the server. 100 byte inputs @ 100HZ.

10 inputs are included in each packet, giving 10X redundancy. If one input packet is lost, the very next one has the lost input, plus the next one we need.

Reliability for inputs is performed entirely in XDP, and inputs (excluding redundant ones) are queued up in a bpf perf buffer to be passed down to the userspace player server application for processing.

BPF perf buffers are per-CPU queues that are processed in userspace. This allows us to run player input processing wide across all cores in a machine (at least, the cores that are involved in packet processing).

https://nakryiko.com/posts/bpf-ringbuf/#bpf-ringbuf-vs-bpf-perfbuf

Player input packets and the resulting inputs processed in userspace via the perf buffer are *always* processed on the same thread, because client packets all go to a single XDP thread for receive queue processing via receive queue hashing of source address and dest address. This is incredibly convenient, because now we don't need any synchronization primitives when modifying player state.

## Results

I'm able to run 1000 clients per n1-standard-8 instance in google cloud. 

I can scale up this up in a managed instance group (MIG) up to any number of players I need and point the players at a player server.

I setup one player server on a c3-standard-44 VM, modified so it has 22 cpus (avoid 2 cores-per CPU) and it tops out processing ~50k players worth of inputs.

Increasing CPU count on the player server VM doesn't allow more player inputs to be processed, so it's definitely IO bound.

How much bandwidth is being sent? 

Each client sends 100 packets per-second, and each packet is around 1300 bytes (10 inputs @ 100 bytes + overhead).

Per-client this gives 100 * 1300 bytes per-second -> 130,000 bytes/sec, or around 130 kilobytes/sec.

Converting to megabits, each client sends just under 1mbit/sec for player inputs.

With each player sending 1 mbit/sec, 1M players are sending 1,000,000 mbit/sec -> 1,000 gbit/sec -> 1 tbit/sec.

That's a non-trivial amount of bandwidth. 

For 50k players we are sending 50,000 * 1mbit -> 50gbit/sec. No wonder google cloud is getting IO bound.

Being conservative, it seems that 20 player servers with 100gbit/nics are required for 1M players.

Assuming ~20k per-month per-player server, the total player server cost per-month is $400k USD

This sounds like a lot, but it's only 40c per-player, per-month.

Verdict: *DEFINITELY POSSIBLE.*
