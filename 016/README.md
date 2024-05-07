# 016

In this version I'm going to try to poll the ring buffer in golang instead of C.

The reason for this is that in the future I'm going to need green threads (aka goroutines) so I can call async functions on other servers without blocking other player simulation udpates.

For example, a player firing a weapon would call out to an async method to raycast from A -> B and find the first object hit from the world server, and we don't want to block other player simulation updates while that async is in flight.

I need to make sure that the golang code runs _exclusively_ on the same CPU that the ring buffer is on.

I'll be following the principles outlined here: https://github.com/valyala/fasthttp#performance-optimization-tips-for-multi-core-systems

Basic idea is to run one go process per-CPU, set GOMAXPROCS to 1 and pin that process to run threads only on the CPU.

This will keep all ring buffer processing on the same CPU that processed the packet.

# Results

Throughput reduces to 14,000 players per-second, which is pathologically low. I believe something must be wrong. 

It's probably a good idea to test on bare metal moving forward instead of google cloud.