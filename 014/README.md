# 014

I want to confirm the NUMA hypothesis.

Use the ring buffer to deliver inputs to the same CPU, instead of trying to deliver to CPUs [16,31]

If the NUMA hypothesis is correct, I expect this will be the same speed or slightly faster than the perf buffer implementation (due to one less copy in XDP code).

The perf buffer implementation achieved 3k players per-16 CTUs. Let's see if the ring buffer implementation can do better.

# Results

...
