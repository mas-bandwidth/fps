# 011

I was able to get the ring buffer working by building latest libbpf and xdp-tools from source.

I setup a single ring buffer in this version, just to see it working. The ring buffer operates on a single CPU via polling.

The next version will have n worker threads that poll a ring buffer each.
