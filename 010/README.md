# 010

Let's bring it all together.

6000 players per-player server implemented via 6 n1-standard-8 machines in google cloud, with 1k players each.

One player server in google cloud with c3-highcpu-44, reduced to 22 CPUs by disabling hyperthreading. We're only going to use 16 CPUs on this.

Still blocked on the ring buffer, so aperf buffer forwards player inputs from XDP to userspace.

Player state lives in userspace and is copied to a bpf hash map after each player simulation step. 

Each player input request is replied to to with the most recent player state in the map for that client in the hash map.

Success is defined as being able to process 6000 players on this server machine, with each client getting player state packets sent back at 100HZ.

# Results

...