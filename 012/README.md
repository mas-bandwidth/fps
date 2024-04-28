# 012

This version distributes player inputs received on CPUs [0,15] to worker threads running on CPUs [16,31].

This should hopefully improve throughput on Google Cloud.

The goal is to get 6k players per-VM with 32 CPUs.

# Results

Results are really bad. Much worse than the last implementation with perf buffer. So bad it feels like there is probably a bug somewhere?
