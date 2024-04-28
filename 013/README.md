# 013

I run exactly the same setup as before, but disable the code that commits the player state back to the bpf map at the end of simulation.

I think this is the bottleneck. If this is true, then the performance of this version should be much higher, and we need to find an alternative way to send player state back to the client.

# Results

...
