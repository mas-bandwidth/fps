# 022

We continue to work through the player raycasting a shot from a -> b to shoot another player.

Now we need to work out which world database (or databases?) should I send the raycast query to.

I propose that the world will be broken up into a bunch of zones. 

Each zone is defined as the union of one or more convex volumes and maps one-to-one to a world database instance.

Each player server shall have a complete mapping of zones at all times, so it can look up a location to the set of world databases that may be relevant to any query.

To keep things simple, I'll setup a simple 2x2 grid, with each cell being 1km cubed.

However, outside of the initial setup of the world in the index server, all code that interacts with these zones will do so in a generic manner, such that each cube is defined as the unions of one or more convex volumes, each volume defined as the set of points n planes.

This way I avoid overly specializing the system to the initial test case.
