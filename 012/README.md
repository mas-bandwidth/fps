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

Something is not right, because the CPU is pegged at only 5% usage, so I'm going to try increasing the ring buffer to wake up every 1000 entries instead of 1 and see if that helps.

