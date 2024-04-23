# 003

In this version I'm debugging why we're getting slow performance in google cloud for player input processing.

My first theory is that by attempting to do significant work on the same cores that are processing packets with XDP we cause the CPU to context switch and interrupt itself a lot between userspace player simulation and XDP packet processing.

What we really want is n CPUs dedicated to XDP, and m CPUs dedicated to player simulation, and for these CPUs to be different.

Since google cloud only uses 16 receive queues per-NIC, and these are the first 16 cores on the machine, the idea that we should try to have the first 16 cores dedicated to XDP only, and then use the next 16 cores exclusively for player simulation.

To do this we need a way to deliver inputs from the XDP program running on CPUs [0,15] -> CPUs [16,31]

Unfortunately, we can't do this with bpf perf buffers, they always deliver data to the same core that the XDP program processed the packet on.

But, if we use the newer bpf ring buffer and setup one ring buffer per-worker CPU we can distribute inputs to whatever CPUs we want.

More information on perf buffers vs. ring buffers here:

https://nakryiko.com/posts/bpf-ringbuf/#bpf-ringbuf-vs-bpf-perfbuf

## Results:

There seems to be a stack smash in libxdp when the ring buffer is created. Blocked until this is fixed.
