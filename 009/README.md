# 009

If memory is the bottleneck, can we scale linearly from 16 CPUs up to 32 CPUs on a bare metal machine?

Increase the MAX_CPUS to 32 and see if we can still process 380 players per-CPU on a 32 core machine.

If we can process fewer players, the memory bottleneck is significant, and calculations should be done to verify that we can even theoretically transfer the player state for @ 50k players.

# Results

Increasing to 32 CPUs and the number of player updates increase from 610k to 650k per-second.

It's very likely hitting a memory bottleneck here. 

Some calculations are required to see if the memory transfer for 50k worth of player state is even possible.

Research is required into NUMA. It may be that there will need to be a bunch of CPUs, each with their own memory under NUMA, so that the NUMA nodes can quickly update their own local memory.

OR...

Alternatively, you could have a much larger number of player servers, each handling less bandwidth and CPU. 

No complicated NUMA would be required, and the system can scale up/down in increments smaller than 50k.

For example, seeing as with 16 cores we can do 600k player updates, this means we can conservatively have ~6000 players per-16 cores. 

For 6000 players, the bandwidth for input packets and player state will be approximately:

1200 * 6000 * 100 bytes per-second = 720,000,000 bytes per-second

720,000,000 * 8 = 5,760,000,000 bits per-second

Or ~5.8 gbps.

This can be handled on a 10G NIC, rather than a 100G NIC.

But now we would need 166 player servers, each with 10G NIC.

Each server will cost $2,160 per-month on datapacket.com with 32 CPUs. Assume we need 16 CPUs for XDP and 16 CPUs for player simulation.

<img width="1347" alt="image" src="https://github.com/mas-bandwidth/fps/assets/696656/9718bd1d-478a-433f-a023-a32d5186a452">

Cost per-month for players servers is now 2160 * 166 = $358,560 USD per-month.

Or ~36c per-player.

It actually got cheaper!

And now we can scale the system up and down in smaller increments of 6k players instead of 50k. Much better!
