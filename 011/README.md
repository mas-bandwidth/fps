# 010

Let's bring it all together.

6000 players per-player server running on a c3-highcpu-44, reduced to 22 CPUs by disabling hyperthreading.

Still blocked on the ring buffer, so a perf buffer forwards player inputs from XDP to userspace for now.

Player state lives in userspace and is copied to a bpf hash map after each player simulation step. 

Each player input request is replied to to with the most recent player state in the map for the player belonging to that client.

Success is defined as being able to process 3000 players on this server machine (half of target), since only 16 cores are used on the google cloud machine to process packets.

# Results

Success. I can scale up to 3k players on google cloud. Going to 4k and some player states don't get sent back to the client... so we're right at the limit.

If we can hit 3k players on a VM with only 16 CPUs, we can certainly hit 6k players on bare metal with 32 CPUs and hit the expected price: 36c per-player per-month.
