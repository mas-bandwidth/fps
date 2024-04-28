# 013

I think committing player state back to the bpf player state map is the bottleneck. 

In this version I run exactly the same setup as before, but disable the code that commits the player state back to the bpf map at the end of simulation.

If this is true, then the performance of this version should be much higher, and we need to find an different way to send player state packets back to the client...

# Results

The results are exactly the same:

```
Apr 28 14:51:23 client-gcqb client[11013]: inputs sent delta 99950, inputs processed delta 227367, player state delta 0
```

The player state map commit is not the bottleneck.

So what is?

My best guess is that we're going across NUMA boundaries with the ring buffer, and this is causing slower memory accesses for player state processing.

https://cloud.google.com/compute/docs/machine-resource

This wouldn't happen on normal 32 CPU bare metal machine.

I think perf testing on Google Cloud VMs past this point is counter productive, since the final solution can't run in google cloud anyway due to egress bandwidth costs.

What I really need to do is start testing on my linux bare metal machines with 10G NICs instead.
