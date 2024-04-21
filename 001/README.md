# 001

To build an FPS with a million players, we're first going to need a way to process player inputs at scale. 

It won't be possible to have all players simulated on one server, there's simply too much bandwidth and CPU cost for a single machine.

So let's create a new type of server. A "player server". Each player server handles player input processing and simulation, each with n players connected to them. 

I expect 50k players can be simulated per player server, so for 1M players, let's assume we can probably have 20 player servers.

Player servers take the player input + delta time (dt) and step the player state forward in time. Players are simulated forward only when input packets arrive from their client. There is no global tick on a player server. This is similar to how most first person shooters in the quake netcode model work. For example, Counterstrike, Titanfall and Apex Legends.

The assumptions made here are: 

1. the world is static
2. each player has a representation of this world in some form that allows for collision detection to be performed
3. players do not physically collide with each other (common in MMOs)

These assumptions together let us simulate player state forward completely independently of each other. This is how we can unlock 1M player performance.

Player inputs are sent at 100HZ, and each input is 100 bytes long. This is typically an overestimate, inputs are usually much smaller in FPS games. But let's be conservative.

Inputs are sent over UDP because they are time series data and are extremely latency sensitive. All inputs must arrive for simulation, or the client will see mispredictions, because the server state is authoritative. We rely on the client and server running the same client and server simulation and getting (approximately) the same result. This is known as client side prediction, or outside of game development, optimistic execution with rollback.

We cannot use TCP for this reliability, because head of line blocking would cause significant delays in input delivery under both latency and packet loss. Instead, we send the most recent 10 inputs in each input packet, thus we have 10X redundancy. Inputs are relatively small compared to player state (1000 bytes) so this strategy is acceptable, and if one input packet is dropped, the very next packet 1/100th of a second later contains the dropped input PLUS the next input we need to step the player forward.

This reliability is performed entirely in XDP, and inputs (excluding redundant ones) are queued up in a bpf perf buffer to be passed down to the userspace player server application for processing.

BPF perf buffers are per-CPU queues that are processed in userspace. This allows us to run player input processing wide across all cores in a machine (at least, the cores that are involved in packet processing).

Player input packets and the resulting inputs processed in userspace via the perf buffer are *always* processed on the same thread, because client packets all go to a single XDP thread for receive queue processing via receive queue hashing of source address and dest address. This is incredibly convenient, because now we don't need any synchronization primitives when modifying player state.

Results:

I'm able to run 1k clients on n1-standard-8 sending input packets at 100HZ, then scale up this up in a managed instance group (MIG) up to 1M players on google cloud.

I can run a player server on c3-standard-22 that tops out processing 50k players worth of inputs. 

Increasing CPU count on the player server instance doesn't allow more player inputs to be processed, so it's definitely IO bound. We're processing around 50gbps/sec, so this is totally understandable.

This confirms the assumption of 20 player servers required for 1M players. Thus, for 1M players, the cost of each player server with bare metal would be ~$16k USD per-month for a 100G NIC + bare metal in datapacket.com

Total player server cost per-month: $320k USD (conservative)
