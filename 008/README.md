# 008

Just so we don't fool ourselves, actually send the most recent player state from the player state bpf hash map down to the client in response to each input packet received.

This way we verify that not only can we get the player state up to the kernel, but we can send it out.

# Results

...