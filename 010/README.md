# 010

Let's bring it all together.

6000 players per-player server running on a c3-highcpu-44, reduced to 22 CPUs by disabling hyperthreading. We're only going to use 16 CPUs.

Still blocked on the ring buffer, so a perf buffer forwards player inputs from XDP to userspace.

Player state lives in userspace and is copied to a bpf hash map after each player simulation step. 

Each player input request is replied to to with the most recent player state in the map for the client's player.

Success is defined as being able to process 6000 players on this server machine, with each client getting player state packets sent back at 100HZ.

# Results

...
