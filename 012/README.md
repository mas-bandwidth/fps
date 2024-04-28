# 012

This version distributes player inputs received on CPUs [0,15] to worker threads running on CPUs [16,31].

This should hopefully improve throughput on Google Cloud.

The goal is to get 6k players per-VM with 32 CPUs.

# Results

I first tried sharding to worker threads with index = session_id % MAX_THREADS, but this tanked throughput down to only 10k players total -- probably due to contention (eg. each XDP thread fighting over the same right buffer)

I then adjusted so that each XDP CPU always delivered to the same ring buffer, and I get this result:

```
	Apr 28 12:45:32 client-8hqb client[11128]: inputs sent delta 99451, inputs processed delta 226808, player state delta 99351
```

Which is decent, in that player states are able to keep up returning back to client at 5k, but the input processing unable to keep up in the worker threads.

Something is not right, because the CPU is pegged at only ~6% usage, so I'm going to try changing the ring buffer from epoll to hammering the consume function directly to see if that helps.

And it tanks throughput even more. What.

```
Apr 28 13:02:04 client-nt4n client[11037]: inputs sent delta 99251, inputs processed delta 2480, player state delta 2480
Apr 28 13:02:05 client-nt4n client[11037]: inputs sent delta 99213, inputs processed delta 2474, player state delta 2473
Apr 28 13:02:06 client-nt4n client[11037]: inputs sent delta 99214, inputs processed delta 2489, player state delta 2489
Apr 28 13:02:07 client-nt4n client[11037]: inputs sent delta 99241, inputs processed delta 2476, player state delta 2485
Apr 28 13:02:08 client-nt4n client[11037]: inputs sent delta 99196, inputs processed delta 2509, player state delta 2485
Apr 28 13:02:09 client-nt4n client[11037]: inputs sent delta 99249, inputs processed delta 2485, player state delta 2477
Apr 28 13:02:10 client-nt4n client[11037]: inputs sent delta 99157, inputs processed delta 2477, player state delta 2482
```

I don't even.