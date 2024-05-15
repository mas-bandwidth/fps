# 018

Run with XDP ring buffer again, this time regularly yielding with runtime.Gosched() and making sure the go workers are actually pinned to the right CPU.

The goal is to demonstrate that it's possible to use goroutines running on a single CPU to process player inputs off the ring buffer, and to show that in this golang code it is also possible to block on TCP request/response calls while yielding to other player simulation updates.

# Results

...