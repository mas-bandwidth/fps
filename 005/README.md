# 006

In this version I move the player state to live in user memory.

This should make player updates around twice as fast, because we don't have to copy down from kernel memory -> user memory, nad then upload user memory -> kernel memory per-update, we only have to copy into kernel memory once.

# Results

...