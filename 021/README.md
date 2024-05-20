# 021

In this version we implement method for the player server to apply damage to another player.

We cannot just send this event to the world database, since the world database does not own the real player state, only a history of shallow state per-player that is read only.

So we must have a connection from each player server to other player servers, where we can send events.

For example, "apply damage: [health to subtract]".

The question should be, how do we route from player server to the correct other player server?

To keep it general, maybe we could keep a map of session id -> player server the are on, and broadcast from each player server to each other player server?

eg. broadcast out player join [session id] and player leave [session id]?

And then following this we can track where each player is across player servers, and automatically route to the correct player server instance.

Hmmmm. But then when a new player server instance connects, it must be told -- on join, the entire state of all players connected to all other player server instances.

And when a player server shuts down, it must broadcast to other player server instances that it is leaving the mesh.

Thinking about this some more I really don't want to implement a service mesh or peer-to-peer thing between the player servers.

I think I will cop it out and have an "index server" which is tracking things globally, and then player servers can each connect to it, rather than O(n^2), and get updated on player servers connecting, disconnecting, and on players joining and leaving.

# Results

...
