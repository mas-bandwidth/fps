# 019

Prototype sending shallow player state from the player server to the world database.

The world database stores this per-player state in one second ring buffer per-player.

At the same time, switch from a text protocal to binary over TCP.

# Results

With naive TCP I can send one ping/pong per-player update, and player state to the world database for around 1000 players across all threads.

Adjusting so that everything is pinned to the same CPU, I still get the same results.

I'm pretty sure this means the bottleneck is not the golang program, but the system CPU usage processing the TCP stack in the kernel.

Next, I'll need to investigate faster ways to implement the TCP server for the world database.

Some promising libraries to evaluate:

https://github.com/maurice2k/tcpserver
https://betterprogramming.pub/gain-the-new-fastest-go-tcp-framework-40ec111d40e6
https://gnet.host
