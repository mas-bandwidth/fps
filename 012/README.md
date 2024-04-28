# 012

This version distributes player inputs received on CPUs [0,15] to worker threads running on CPUs [16,31].

This should hopefully improve throughput on Google Cloud.

The goal is to get 6k players per-VM with 32 CPUs.

# Results

Results are really bad. Much worse than the last implementation with perf buffer. So bad it feels like there is probably a bug somewhere?

If it's not a bug, then possibly the issue is contention between the read access on the player state map on the XDP threads vs. the write access on the player simulation worker threads?

I still want an option that separates XDP and worker threads for later ideas (especially, player simulation blocking like goroutines when it calls out to world servers...).

So I'm going to need to work out how to minimize contention in this setup.