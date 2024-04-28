# 014

I want to confirm the NUMA hypothesis.

Use the ring buffer to deliver inputs to the same CPU, instead of trying to deliver to CPUs [16,31]

If the NUMA hypothesis is correct, I expect this will be the same speed or slightly faster than the perf buffer implementation (due to one less copy in XDP code).

The perf buffer implementation achieved 3k players per-16 CPUs. Let's see if the ring buffer implementation can do better.

# Results

Results seem to confirm the NUMA hypothesis. We can now do 4.5k players per-16 CPUs, which is much faster than the 2.5k we could do with the NUMA bottleneck.

```
Apr 28 16:27:00 client-ctng client[11220]: inputs sent delta 99720, inputs processed delta 452026, player state delta 0
Apr 28 16:27:01 client-ctng client[11220]: inputs sent delta 99658, inputs processed delta 453166, player state delta 0
Apr 28 16:27:02 client-ctng client[11220]: inputs sent delta 99496, inputs processed delta 452517, player state delta 0
Apr 28 16:27:03 client-ctng client[11220]: inputs sent delta 99722, inputs processed delta 452556, player state delta 0
Apr 28 16:27:04 client-ctng client[11220]: inputs sent delta 99947, inputs processed delta 452356, player state delta 0
Apr 28 16:27:05 client-ctng client[11220]: inputs sent delta 99624, inputs processed delta 452027, player state delta 0
Apr 28 16:27:06 client-ctng client[11220]: inputs sent delta 99735, inputs processed delta 452344, player state delta 0
Apr 28 16:27:07 client-ctng client[11220]: inputs sent delta 99764, inputs processed delta 453066, player state delta 0
```
