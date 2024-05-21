# 021

In this version we start to work through a player raycasting to on a server, hitting another player and applying damage to that player.

We cannot just send this damage player event to the world database, since the world database does not own the real player state, and cannot modify the player health.

So we must have a connection from each player server to other player servers, where we can send events to affect other players.

For example, "apply damage: [health to subtract]" -> player server that player is on.

But then the question becomes, how do we route from player server to the correct other player server?

The general idea moving forward is that now we have an "index server", which tracks the set of player servers active (and eventually also the definition of the world, world servers, and the world database servers assigned to it).

The player server connects to the index server on startup and is assigned a uint32 tag that identifies this player server globally in the system.

Every second the player server gets an updated list of all player servers active from the index server, so it can locally mirror the hash looking up player server addresses from tags.

Then, in the future when we implement a raycast, the raycast would go out to the world databases via an async call, and the response would include both a tag, and a session id that identifies both the player and the player server the player is on.

This way we can then send a damage player command to the correct player server if the raycast hits anything.
