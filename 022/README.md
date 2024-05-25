# 022

In this version, we continue to prototype using the case study of a player raycasting from a -> b to shoot another player and apply damage.

The world is now made up of a number of different zones. Each zone has its own database to store object state for the last 1 second in a ring buffer and subscribes to player state in a volume defined as the union of one or more convex volumes.

To keep things simple, I'll setup a simple 2x2 grid world, with each zone being a 1km cubed grid cell.

Each player server has a complete mapping of zones, so it can look up potentially relevant zones for any location in constant time. For now I'm going to do this roughly, with a low resolution grid across the world that maps to the set of potential zones touched by each grid cell. Of course other spatial data structures in the future could be much better fits, especially if the world is large, but sparse.

Bringing it all together is a world server (previously index server), which owns the definition of the world and manages the sets of connected player servers and zone databases (previously world databases).

