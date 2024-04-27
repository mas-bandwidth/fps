# 012

This version distributes player inputs received on any XDP thread to 16 worker threads with session_id % 16

The XDP threads on Google Cloud are on CPUs [0,15] and the worker threads are on CPUs [16,31], so they shouldn't fight with each other.

This should hopefully improve throughput on Google Cloud.

The goal is to get 6k players per-machine with 32 CPUs.

# Results

...