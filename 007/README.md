# 007

In this version I move the player state to live in user memory instead of kernel memory.

This should make player updates maybe twice as fast, because we don't have to copy down from kernel memory -> user memory, update the player, then upload user memory -> kernel memory, we only have to copy to kernel memory once.

This lets us send out the most recent player state in response to input packets sent by clients.

# Results

I can now do around 550k updates per-second with 16 cpus.

This means that we should be able to theoretically do 50k players with 144 cpus.
