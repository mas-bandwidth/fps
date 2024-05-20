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

Notes on further TCP scalability in C: 

https://jackiedinh8.medium.com/1m-tcp-connections-in-c-511da0b1a283

The good thing is that there is an existence proof this is possible. We can have 10k TCP clients per-server and it is an expected, fairly standard benchmark that we should be able to meet.

My guess is that the best way to get there is to use evio with TCP in golang, and switch to epoll and away from one goroutine polling sockets per-connection.

Moving forward, I feel confident that this can be achieved. So I will focus more effort now on implementing functionality in the world server instead, and switch back to scalability over TCP when I have access to Linux machines again.
