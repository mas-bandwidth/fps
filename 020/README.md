# 020

In this version we try to improve TCP server performance for the world database.

Split up into 250 players per-CPU, then run multiple player instances to simulate the player server (I'm travelling so I don't have access to my linux bare metal machines in my office).

Some promising libraries to evaluate:

https://github.com/maurice2k/tcpserver
https://betterprogramming.pub/gain-the-new-fastest-go-tcp-framework-40ec111d40e6
https://gnet.host

# Results

With maurice2k/tcpserver can still only do up to around 1000 players before it can't keep up.

With the bandwidth I'm sending, for 1000 players I have:

100 bytes * 100 * 1000 = 10,000,000 bytes per-second, = 80000000 bits per-second = 80 mbit/sec.

Which I should hope I should be able to do on my macbook air m2 over localhost?

Let's try gnet instead to see if it's faster...

Looks like gnet is built on https://github.com/tidwall/evio

Seems that the big win with TCP is to replace select with epoll on Linux.

https://en.wikipedia.org/wiki/Epoll

I don't think I can go much further with MacOS testing. Time to switch to Linux.