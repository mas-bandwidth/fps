# 008

Just so we don't fool ourselves, actually send the most recent player state from the player state bpf hash map down to the client in response to each input packet received.

This way we verify that not only can we get the player state up to the kernel, but we can also use it and send out.

Decrease player state to 1000 bytes, so it fits into 

# Results

We can now send 610000 player states per-second on 16 cpus.

This means we could have 380 players per-CPU, and 50k players would take 131 CPUs.
