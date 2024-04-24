# 008

Just so we don't fool ourselves, actually send the most recent player state from the player state bpf hash map down to the client in response to each input packet received.

This way we verify that not only can we get the player state up to the kernel, but we can also use it and send out.

Decrease player state to 1000 bytes, so it fits into the input packet to be send as a response.

# Results

We can now send 610000 player states per-second on 16 cpus.

This gives us ~380 players per-CPU, so 50k players would take 131 CPUs.

1200 bytes down to 1000 bytes gave a clear speed up, this seems to indicate we are primarily bound by memory speeds.

To verify that we can really hit the 50k number, it's quite possible that memory will become the bottleneck. Perhaps NUMA will be required to hit 50k.
