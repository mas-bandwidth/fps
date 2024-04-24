# 006

In this version I move the player state to live in user memory.

This should make player updates maybe twice as fast, because we don't have to copy down from kernel memory -> user memory, nad then upload user memory -> kernel memory per-update, we only have to copy into kernel memory once.

# Results

I can now do around 550k updates per-second with 16 cpus.

This means that we should be able to theoretically do 50k players with 144 cpus.
