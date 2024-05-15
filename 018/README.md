# 018

In this version the goals are:

1. Demonstrate that it's possible to use goroutines running on a single CPU to process player inputs from the ring buffer
2. Show that it's possible to block on IO calls and yielding to other player simulation updates on a single CPU

This time regularly yield with runtime.Gosched() at key places and make sure the go workers are actually pinned to the right CPU.

To mock the yield while waiting on blocking IO, implement a simple TCP server in world_database.go with a blocking TCP socket connection per-player. 

When the world database receives "ping" it returns the response "pong".

# Results

We can run multiple clients and their inputs get processed correctly. 

The IO block appears to be yielding to other player input processing functions.
