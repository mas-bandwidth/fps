# 001

To build an FPS with a million players, we're first going to need a way to process player inputs at scale. 

It won't be possible to have all players simulated on one server, there's simply too much bandwidth and CPU cost for a single machine.

So let's create a new type of server. A "player server". Each player server handles player input processing and simulation, each with n players connected to them. 

Player servers take the player input + delta time (dt) and step the player state forward in time. Players are simulated forward only when input packets arrive from their client. There is no global tick on a player server. This is similar to how most first person shooters in the quake netcode model work. For example, Counterstrike, Titanfall and Apex Legends.

The assumptions made here are: 

1. the world is static
2. each player has a representation of this world in some form that allows for collision detection to be performed
3. players do not physically collide with each other (common in MMOs)

These assumptions together let us simulate player state forward completely independently of each other. This is how we can unlock 1M player performance.

Player inputs are sent at 100HZ, and each input is 100 bytes long. This is typically an overestimate, inputs are usually much smaller in FPS games. But let's be conservative.

Inputs are sent over UDP because they are time series data and are extremely latency sensitive. All inputs must arrive for simulation, or the client will see mispredictions, because the server state is authoritative. We rely on the client and server running the same simulation on the same inputs and delta times and getting (approximately) the same result. This is known as client side prediction, or outside of game development, optimistic execution with rollback.

We cannot use TCP for this reliability, because head of line blocking would cause significant delays in input delivery under both latency and packet loss. Instead, we send the most recent 10 inputs in each input packet, thus we have 10X redundancy. Inputs are relatively small so this strategy is acceptable, and if one input packet is dropped, the very next packet 1/100th of a second later contains the dropped input PLUS the next input we need to step the player forward. Perfect.

This reliability is performed entirely in XDP, and inputs (excluding redundant ones) are queued up in a bpf perf buffer to be passed down to the userspace player server application for processing.

BPF perf buffers are per-CPU queues that are processed in userspace. This allows us to run player input processing wide across all cores in a machine (at least, the cores that are involved in packet processing).

https://nakryiko.com/posts/bpf-ringbuf/#bpf-ringbuf-vs-bpf-perfbuf

Player input packets and the resulting inputs processed in userspace via the perf buffer are *always* processed on the same thread, because client packets all go to a single XDP thread for receive queue processing via receive queue hashing of source address and dest address. This is incredibly convenient, because now we don't need any synchronization primitives when modifying player state.

## Results:

I'm able to run 1k clients on n1-standard-8 sending input packets at 100HZ, then scale up this up in a managed instance group (MIG) up to 1M players on google cloud.

I can run a player server on c3-standard-44 modified so it has 22 cpus (avoid 2 cores-per CPU) and it tops out processing ~50k players worth of inputs. 

Increasing CPU count on the player server instance doesn't allow more player inputs to be processed, so it's definitely IO bound.

How much bandwidth is being sent? 

100 packets per-second, and each packet is around 1300 bytes (10 inputs @ 100 bytes + overhead).

Per-client this gives 100*1300 bytes per-second -> 130,000 bytes/sec, or around 130 kilobytes/sec.

Converting this to megabits, we see that each client sends just under 1mbit/sec for player inputs.

With each player sending 1 mbit/sec, 1M players are sending 1,000,000 mbit/sec -> 1,000gbit/sec -> 1tbit/sec.

That's a non-trivial amount of bandwidth. 

For 50k players we are sending 50,000 * 1mbit = 50,000 mbit/sec -> 50gbit/sec. No wonder google cloud is getting IO bound @ 50k players.

Being conservative, it seems that 20 player servers with 100gbit/nics are required for 1M players.

Assuming ~20k per-month per-player server, the total player server cost per-month is $400k USD

Verdict: *DEFINITELY POSSIBLE.*
