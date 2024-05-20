# 019

Prototype sending shallow player state from the player server to the world database.

The world database stores this per-player state in one second ring buffer per-player.

At the same time, switch from a text protocal to binary over TCP.

# Results

With naive TCP I can send one ping/pong per-player update, and player state to the world database for around 1000 players across all threads.

Going above 1200 players on my macbook air m2 results in high system CPU use and the TCP server cannot keep up.

Adjusting so that everything is pinned to the same CPU, I still get the same results, perhaps with a bit fewer players, slightly less than 1000. 

The bottleneck seems to be system CPU usage processing the TCP stack in the kernel.

Next, I'll need to investigate faster ways to implement the TCP server for the world database.
