# 009

If memory is the bottleneck, can we scale linearly from 16 CPUs up to 32 CPUs on a bare metal machine?

Increase the MAX_CPUS to 32 and see if we can still process 380 players per-CPU on a 32 core machine.

If we can process fewer players, the memory bottleneck is significant, and calculations should be done to verify that we can even theoretically transfer the player state for @ 50k players.

# Results

...
