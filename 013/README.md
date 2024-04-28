# 013

I think committing player state back to the bpf player state map is the bottleneck. 

In this version I run exactly the same setup as before, but disable the code that commits the player state back to the bpf map at the end of simulation.

If this is true, then the performance of this version should be much higher, and we need to find an different way to send player state packets back to the client...

# Results

...
