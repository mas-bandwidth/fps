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

Hits the same limit:

```
Apr 28 13:12:28 client-v07b client[11046]: inputs sent delta 99340, inputs processed delta 227622, player state delta 98809
Apr 28 13:12:29 client-v07b client[11046]: inputs sent delta 99396, inputs processed delta 227465, player state delta 98854
Apr 28 13:12:30 client-v07b client[11046]: inputs sent delta 99521, inputs processed delta 227486, player state delta 98989
```

This implementation is only able to reach 2k players.

The results are dissapointing. But there has to be a way to fix it... the CPU is barely being hit. 6% in the epoll example, and only 20% in the busy poll loop on the ring buffer.

So there is some contention issue perhaps with the ring buffer or player state map that is limiting throughput? I think it's most likely the player state map? If the ring buffer was being continually blocked writing to the player state map by the XDP threads, that would make a lot of sense.

What if I sharded the player state map into multiple maps per-CPU? 

Or could I use a different structure to communicate the player state back from userspace to XDP?

Maybe I could communicate the player state update back to the main XDP CPU on a userspace thread, so it can commit the updated player state without contention with the XDP program?

What if I even sent the player state packets back to the client using a regular socket with SO_REUSEPORT?
