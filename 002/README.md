# 002

Previously we have shown that we can actually receive and process 1M player inputs with 20 bare metal "player servers".

The next step is to actually step 1M player simulations forward on the player servers.

To do this we must:

1. load the most recent player state at time t by session id (random uint64 per-client assigned on connect)
2. step the player state forward with input and for the amount of time dt, eg. position += velocity * dt
3. store the updated player state post-simulation at the new time t += dt

We also need to send the player state down to the client that owns it 100 times per-second. This is important because the server *must* be authoritative over the player simulation. 

When the client receives the player state it rolls back in the past and applies that state and then invisibly resimulates back up to present time with stored local player inputs. 

This way the server correction can be applied RTT in the past, and the client can accept that update without being pulled back in time whenever the server corrects it.

To send player state we'll piggy back on the input packets. For each input packet we receive, we'll respond with the most recent player state for the client that sent the input packet, again looking it up via session id.

All of this points to bpf hash maps being a good fit for what we need to do. 

https://docs.kernel.org/bpf/map_hash.html

We can access bpf hash maps from inside the XDP program, and we can also read and write to them from the userspace server application. We can even use an LRU hashmap variant so we don't have to do any work to clean it up when players disconnect from the server.

Let's assume that player state is around 1200 bytes. This gives us a nice symmetric protocol between the client and the player server: around 1mbit/sec for inputs, and ~1mbit/sec is sent back down to the client for player state.

We're going to have 50k players per-player server, so we need around 1200 * 50000 = 65GB for store player state. 

Assume there is some overhead with the bpf hash map and conservatively we'll need 100GB for player state.

All signs are pointing towards player servers needing a good amount of memory for the perf buffers for processing inputs, as well as to store the player state. 

But consider, the sort of high powered bare metal machine that could drive a 100G NIC at line rate would mean that we probably already have 256GB of memory already. So this should be fine.

## Results:

There is some bottleneck and throughput drops from 50k players to ~5k players per-player server.

I believe this is the result of the google cloud NIC only having 16 receive queues max, and the cpus [0,15] getting thrashed between running the XDP program in kernel space, and running player simulations in userspace.

Alternatively, it could be that reading and updating player state in bpf hash maps is the bottleneck.
