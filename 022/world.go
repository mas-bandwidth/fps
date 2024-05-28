package main

import (
    "fmt"
)

const Meter = int64(1000000)
const Kilometer = 1000 * Meter
const Centimeter = Meter / 100
const Millimeter = Meter / 1000
const Micrometer = Meter / 1000000

// ---------------------------------------------------------

type Vector struct {
    x int64
    y int64
    z int64
}

func (value *Vector) Write(data []byte, index *int) {
    WriteInt64(data, index, value.x)
    WriteInt64(data, index, value.y)
    WriteInt64(data, index, value.z)
}

func (value *Vector) Read(data []byte, index *int) bool {
    if !ReadInt64(data, index, &value.x) {
        return false
    }
    if !ReadInt64(data, index, &value.y) {
        return false
    }
    if !ReadInt64(data, index, &value.z) {
        return false
    }
    return true
}

// ---------------------------------------------------------

func Dot(x int64, y int64, z int64, nx int64, ny int64, nz int64) int64 {
    dx := ( x * nx ) / Meter
    dy := ( y * ny ) / Meter
    dz := ( z * nz ) / Meter
    return dx + dy + dz
}

// ---------------------------------------------------------

type Plane struct {
    normal Vector
    d      int64
}

func (value *Plane) Write(data []byte, index *int) {
    value.normal.Write(data, index)
    WriteInt64(data, index, value.d)    
}

func (value *Plane) Read(data []byte, index *int) bool {
    if !value.normal.Read(data, index) {
        return false
    }
    if !ReadInt64(data, index, &value.d) {
        return false
    }
    return true
}

// ---------------------------------------------------------

type AABB struct {
    min Vector
    max Vector
}

func (value *AABB) Write(data []byte, index *int) {
    value.min.Write(data, index)
    value.max.Write(data, index)
}

func (value *AABB) Read(data []byte, index *int) bool {
    if !value.min.Read(data, index) {
        return false
    }
    if !value.max.Read(data, index) {
        return false
    }
    return true
}

// ---------------------------------------------------------

type Volume struct {
    bounds AABB
    planes []Plane
}

func (volume *Volume) Inside(x int64, y int64, z int64) bool {
    for i := range volume.planes {
        if Dot(x, y, z, volume.planes[i].normal.x, volume.planes[i].normal.y, volume.planes[i].normal.z) < volume.planes[i].d {
            return false
        }
    }
    return true
}

func (value *Volume) Write(data []byte, index *int) {
    value.bounds.Write(data, index)
    numPlanes := len(value.planes)
    WriteInt(data, index, numPlanes)
    for i := 0; i < numPlanes; i++ {
        value.planes[i].Write(data, index)
    }
}

func (value *Volume) Read(data []byte, index *int) bool {
    if !value.bounds.Read(data, index) {
        return false
    }
    var numPlanes int
    if !ReadInt(data, index, &numPlanes) {
        return false
    }
    value.planes = make([]Plane, numPlanes)
    for i := 0; i < numPlanes; i++ {
        if !value.planes[i].Read(data, index) {
            return false
        }
    }
    return true
}

// ---------------------------------------------------------

type Zone struct {
    id      uint32
    origin  Vector
    bounds  AABB
    volumes []Volume
}

func (zone *Zone) Inside(x int64, y int64, z int64) bool {
    for i := range zone.volumes {
        if zone.volumes[i].Inside(x, y, z) {
            return true
        }
    }
    return false
}

func (value *Zone) Write(data []byte, index *int) {
    WriteUint32(data, index, value.id)
    value.origin.Write(data, index)
    value.bounds.Write(data, index)
    numVolumes := len(value.volumes)
    WriteInt(data, index, numVolumes)
    for i := 0; i < numVolumes; i++ {
        value.volumes[i].Write(data, index)
    }
}

func (value *Zone) Read(data []byte, index *int) bool {
    if !ReadUint32(data, index, &value.id) {
        return false
    }
    if !value.origin.Read(data, index) {
        return false
    }
    if !value.bounds.Read(data, index) {
        return false
    }
    var numVolumes int
    if !ReadInt(data, index, &numVolumes) {
        return false
    }
    value.volumes = make([]Volume, numVolumes)
    for i := 0; i < numVolumes; i++ {
        if !value.volumes[i].Read(data, index) {
            return false
        }
    }
    return true
}

// ---------------------------------------------------------

type World struct {
    bounds  AABB
    zones   []Zone
    zoneMap map[uint32]*Zone
}

func (world *World) Fixup() {
    world.zoneMap = make(map[uint32]*Zone, len(world.zones))
    for i := range world.zones {
        world.zoneMap[world.zones[i].id] = &world.zones[i]
    }
}

func (world *World) Print() {

    fmt.Printf("world bounds are (%d,%d,%d) -> (%d,%d,%d)\n", 
        world.bounds.min.x,
        world.bounds.min.y,
        world.bounds.min.z,
        world.bounds.max.x,
        world.bounds.max.y,
        world.bounds.max.z,
    )

    fmt.Printf("world has %d zones:\n", len(world.zones))

    for i := range world.zones {
        fmt.Printf(" + 0x%08x: (%d,%d,%d) -> (%d,%d,%d)\n",
            world.zones[i].id,
            world.zones[i].bounds.min.x,
            world.zones[i].bounds.min.y,
            world.zones[i].bounds.min.z,
            world.zones[i].bounds.max.x,
            world.zones[i].bounds.max.y,
            world.zones[i].bounds.max.z,
        )
    }
}

func (value *World) Write(data []byte, index *int) {
    value.bounds.Write(data, index)
    numZones := len(value.zones)
    WriteInt(data, index, numZones)
    for i := 0; i < numZones; i++ {
        value.zones[i].Write(data, index)
    }
}

func (value *World) Read(data []byte, index *int) bool {
    if !value.bounds.Read(data, index) {
        return false
    }
    var numZones int
    if !ReadInt(data, index, &numZones) {
        return false
    }
    value.zones = make([]Zone, numZones)
    for i := 0; i < numZones; i++ {
        if !value.zones[i].Read(data, index) {
            return false
        }
    }
    value.Fixup()
    return true
}

// ---------------------------------------------------------

func generateWorld_Grid(i int64, j int64, k int64, cellSize int64) *World {

    fmt.Printf("generating grid world: %dx%dx%d\n", i, j, k)
    
    world := World{}

    world.bounds.max.x = i * int64(cellSize)
    world.bounds.max.y = j * int64(cellSize)
    world.bounds.max.z = k * int64(cellSize)

    numZones := i * j * k

    world.zones = make([]Zone, numZones)

    index := 0

    for y := int64(0); y < j; y++ {

        for z := int64(0); z < k; z++ {

            for x := int64(0); x < i; x++ {

                world.zones[index].id = uint32(index) + 1

                world.zones[index].bounds.min.x = x * int64(cellSize)
                world.zones[index].bounds.min.y = y * int64(cellSize)
                world.zones[index].bounds.min.z = z * int64(cellSize)

                world.zones[index].bounds.max.x = (x+1) * int64(cellSize)
                world.zones[index].bounds.max.y = (y+1) * int64(cellSize)
                world.zones[index].bounds.max.z = (z+1) * int64(cellSize)

                world.zones[index].origin.x = ( world.zones[index].bounds.min.x + world.zones[index].bounds.max.x ) / 2
                world.zones[index].origin.y = ( world.zones[index].bounds.min.y + world.zones[index].bounds.max.y ) / 2
                world.zones[index].origin.z = ( world.zones[index].bounds.min.z + world.zones[index].bounds.max.z ) / 2

                world.zones[index].volumes = make([]Volume, 1)

                world.zones[index].volumes[0].bounds = world.zones[index].bounds

                world.zones[index].volumes[0].planes = make([]Plane, 6)

                // left plane

                world.zones[index].volumes[0].planes[0].normal.x = Meter
                world.zones[index].volumes[0].planes[0].normal.y = 0
                world.zones[index].volumes[0].planes[0].normal.z = 0
                world.zones[index].volumes[0].planes[0].d = x * Meter

                // right plane

                world.zones[index].volumes[0].planes[1].normal.x = -Meter
                world.zones[index].volumes[0].planes[1].normal.y = 0
                world.zones[index].volumes[0].planes[1].normal.z = 0
                world.zones[index].volumes[0].planes[1].d = x * Meter + int64(cellSize)

                // bottom plane

                world.zones[index].volumes[0].planes[2].normal.x = 0
                world.zones[index].volumes[0].planes[2].normal.y = Meter
                world.zones[index].volumes[0].planes[2].normal.z = 0
                world.zones[index].volumes[0].planes[2].d = y * Meter

                // top plane

                world.zones[index].volumes[0].planes[3].normal.x = 0
                world.zones[index].volumes[0].planes[3].normal.y = -Meter
                world.zones[index].volumes[0].planes[3].normal.z = 0
                world.zones[index].volumes[0].planes[3].d = y * Meter + int64(cellSize)

                // front plane

                world.zones[index].volumes[0].planes[4].normal.x = 0
                world.zones[index].volumes[0].planes[4].normal.y = 0
                world.zones[index].volumes[0].planes[4].normal.z = Meter
                world.zones[index].volumes[0].planes[4].d = z * Meter

                // back plane

                world.zones[index].volumes[0].planes[5].normal.x = 0
                world.zones[index].volumes[0].planes[5].normal.y = 0
                world.zones[index].volumes[0].planes[5].normal.z = -Meter
                world.zones[index].volumes[0].planes[5].d = z * Meter + int64(cellSize)

                index++
            }

        }

    }
    
    return &world
}

// ---------------------------------------------------------

type WorldGridCell struct {
    zones []*Zone
}

type WorldGrid struct {
    i        int32
    j        int32
    k        int32
    cellSize int64
    bounds   AABB
    cells    [][][]WorldGridCell
}

func createWorldGrid(world *World, cellSize int64) *WorldGrid {
    
    dx := world.bounds.max.x - world.bounds.min.x
    dy := world.bounds.max.y - world.bounds.min.y
    dz := world.bounds.max.z - world.bounds.min.z

    cx := dx / int64(cellSize)
    cy := dy / int64(cellSize)
    cz := dz / int64(cellSize)

    if dx % int64(cellSize) != 0 {
        cx++
    }

    if dy % int64(cellSize) != 0 {
        cy++
    }

    if dz % int64(cellSize) != 0 {
        cz++
    }

    cellCount := cx * cy * cz

    fmt.Printf("world grid has %d cells\n", cellCount)

    grid := &WorldGrid{}

    grid.cells = make([][][]WorldGridCell, cz)

    numZones := len(world.zones)

    inside := make([]bool, numZones)

    for k := 0; k < int(cz); k++ {
        z := world.bounds.min.z + int64(cellSize) * int64(k)
        grid.cells[k] = make([][]WorldGridCell, cy)
        for j := 0; j < int(cy); j++ {
            y := world.bounds.min.y + int64(cellSize) * int64(j)
            grid.cells[k][j] = make([]WorldGridCell, cx)
            for i := 0; i < int(cx); i++ {
                x := world.bounds.min.x + int64(cellSize) * int64(i)
                for n := 0; n < numZones; n++ {
                    if world.zones[i].Inside(x, y, z) {
                        inside[n] = true
                    } else {
                        inside[n] = false
                    }
                }
            }
        } 
    }

    fmt.Printf("finished crunching world grid\n")

    return grid
}

// ---------------------------------------------------------
