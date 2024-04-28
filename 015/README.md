# 015

In this version we hook everything back up so that player state packets are returned to the client for each input packet received.

I expect this will slightly reduce throughput, but hopefully the ring buffer comes out ahead of the ~3k players that could be processed with the perf buffer.

Also, since I'm not seeing any throughput benefits from directly polling the ring buffer consume on the worker threads, I'm going to go back to using epoll so we free up some CPU.

Since it seems that the best result is to process packets on the CPU that receives them, we'll need some spare cycles on each CPU in order to do actual work for player simulation.

# Results

...
