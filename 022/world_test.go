package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Equal_Vector(t *testing.T) {

	type In struct {
		u Vector
		v Vector
	}

	type Params struct {
		in  In
		out bool
	}

	var parameters = []Params{
		{In{Vector{Meter, 0, 0}, Vector{Meter, 0, 0}}, true},
		{In{Vector{0, -Meter, 0}, Vector{0, -Meter, 0}}, true},
		{In{Vector{-Meter, 0, 0}, Vector{Meter, 0, 0}}, false},
		{In{Vector{Meter, 0, 0}, Vector{0, 0, Meter}}, false},
	}

	for index, parameter := range parameters {
		t.Run(fmt.Sprintf("equal_vector_%d", index), func(t *testing.T) {
			assert.Equal(t, parameter.out, parameter.in.u.Equal(&parameter.in.v))
		})
	}

}

func Test_Dot(t *testing.T) {

	type In struct {
		u Vector
		v Vector
	}

	type Params struct {
		in  In
		out int64
	}

	var parameters = []Params{
		{In{Vector{Meter, 0, 0}, Vector{Meter, 0, 0}}, Meter},
		{In{Vector{-Meter, 0, 0}, Vector{Meter, 0, 0}}, -Meter},
		{In{Vector{Meter, 0, 0}, Vector{-Meter, 0, 0}}, -Meter},
		{In{Vector{-Meter, 0, 0}, Vector{-Meter, 0, 0}}, Meter},
		{In{Vector{0, 0, 0}, Vector{-Meter, 0, 0}}, 0},
	}

	for index, parameter := range parameters {
		t.Run(fmt.Sprintf("dot_%d", index), func(t *testing.T) {
			assert.Equal(t, parameter.out, Dot(parameter.in.u.x, parameter.in.u.y, parameter.in.u.z, parameter.in.v.x, parameter.in.v.y, parameter.in.v.z))
		})
	}

}

func Test_InFront_Plane(t *testing.T) {

	type In struct {
		plane Plane
		point Vector
	}

	type Params struct {
		in  In
		out bool
	}

	var parameters = []Params{
		{In{Plane{Vector{Meter, 0, 0}, 0}, Vector{0, 0, 0}}, true},
		{In{Plane{Vector{Meter, 0, 0}, 0}, Vector{Meter, 0, 0}}, true},
		{In{Plane{Vector{Meter, 0, 0}, 0}, Vector{-Meter, 0, 0}}, false},
		{In{Plane{Vector{Meter, 0, 0}, 0}, Vector{0, Meter, 0}}, true},
		{In{Plane{Vector{Meter, 0, 0}, 0}, Vector{0, Meter, -Meter}}, true},
		{In{Plane{Vector{0, Meter, 0}, 0}, Vector{0, Meter, -Meter}}, true},
		{In{Plane{Vector{-Meter, 0, 0}, -Meter}, Vector{0, 0, 0}}, true},
		{In{Plane{Vector{-Meter, 0, 0}, -Meter}, Vector{Meter, 0, 0}}, true},
		{In{Plane{Vector{-Meter, 0, 0}, -Meter}, Vector{2 * Meter, 0, 0}}, false},
	}

	for index, parameter := range parameters {
		t.Run(fmt.Sprintf("infront_plane_%d", index), func(t *testing.T) {
			assert.Equal(t, parameter.out, parameter.in.plane.InFront(parameter.in.point.x, parameter.in.point.y, parameter.in.point.z))
		})
	}

}

func Test_Inside_Volume(t *testing.T) {

	type Params struct {
		in  Vector
		out bool
	}

	var parameters = []Params{
		{Vector{0, 0, 0}, true},
		{Vector{Meter / 2, Meter / 2, Meter / 2}, true},
		{Vector{Meter, Meter, Meter}, true},
		{Vector{-1, 0, 0}, false},
		{Vector{0, -1, 0}, false},
		{Vector{0, 0, -1}, false},
		{Vector{Meter * 2, Meter * 2, Meter * 2}, false},
	}

	volume := Volume{}

	volume.planes = make([]Plane, 6)

	volume.planes[0].normal.x = Meter
	volume.planes[0].normal.y = 0
	volume.planes[0].normal.z = 0
	volume.planes[0].d = 0

	volume.planes[1].normal.x = -Meter
	volume.planes[1].normal.y = 0
	volume.planes[1].normal.z = 0
	volume.planes[1].d = -Meter

	volume.planes[2].normal.x = 0
	volume.planes[2].normal.y = Meter
	volume.planes[2].normal.z = 0
	volume.planes[2].d = 0

	volume.planes[3].normal.x = 0
	volume.planes[3].normal.y = -Meter
	volume.planes[3].normal.z = 0
	volume.planes[3].d = -Meter

	volume.planes[4].normal.x = 0
	volume.planes[4].normal.y = 0
	volume.planes[4].normal.z = Meter
	volume.planes[4].d = 0

	volume.planes[5].normal.x = 0
	volume.planes[5].normal.y = 0
	volume.planes[5].normal.z = -Meter
	volume.planes[5].d = -Meter

	for index, parameter := range parameters {
		t.Run(fmt.Sprintf("inside_volume_%d", index), func(t *testing.T) {
			assert.Equal(t, parameter.out, volume.Inside(parameter.in.x, parameter.in.y, parameter.in.z))
		})
	}
}

func Test_Inside_Zone(t *testing.T) {

	type Params struct {
		in  Vector
		out bool
	}

	var parameters = []Params{
		{Vector{0, 0, 0}, true},
		{Vector{Meter / 2, Meter / 2, Meter / 2}, true},
		{Vector{Meter, Meter, Meter}, true},
		{Vector{Meter * 2, Meter, Meter}, true},
		{Vector{Meter * 2, Meter * 2, Meter}, false},
		{Vector{-Meter * 2, 0, 0}, false},
	}

	volume_1 := Volume{}

	volume_1.planes = make([]Plane, 6)

	// back
	volume_1.planes[0].normal.x = Meter
	volume_1.planes[0].normal.y = 0
	volume_1.planes[0].normal.z = 0
	volume_1.planes[0].d = 0
	// front
	volume_1.planes[1].normal.x = -Meter
	volume_1.planes[1].normal.y = 0
	volume_1.planes[1].normal.z = 0
	volume_1.planes[1].d = -Meter
	// left
	volume_1.planes[2].normal.x = 0
	volume_1.planes[2].normal.y = Meter
	volume_1.planes[2].normal.z = 0
	volume_1.planes[2].d = 0
	// right
	volume_1.planes[3].normal.x = 0
	volume_1.planes[3].normal.y = -Meter
	volume_1.planes[3].normal.z = 0
	volume_1.planes[3].d = -Meter
	// bottom
	volume_1.planes[4].normal.x = 0
	volume_1.planes[4].normal.y = 0
	volume_1.planes[4].normal.z = Meter
	volume_1.planes[4].d = 0
	// top
	volume_1.planes[5].normal.x = 0
	volume_1.planes[5].normal.y = 0
	volume_1.planes[5].normal.z = -Meter
	volume_1.planes[5].d = -Meter

	volume_2 := Volume{}

	volume_2.planes = make([]Plane, 6)
	// back
	volume_2.planes[0].normal.x = Meter
	volume_2.planes[0].normal.y = 0
	volume_2.planes[0].normal.z = 0
	volume_2.planes[0].d = Meter
	// front
	volume_2.planes[1].normal.x = -Meter
	volume_2.planes[1].normal.y = 0
	volume_2.planes[1].normal.z = 0
	volume_2.planes[1].d = -Meter * 2
	// left
	volume_2.planes[2].normal.x = 0
	volume_2.planes[2].normal.y = Meter
	volume_2.planes[2].normal.z = 0
	volume_2.planes[2].d = 0
	// right
	volume_2.planes[3].normal.x = 0
	volume_2.planes[3].normal.y = -Meter
	volume_2.planes[3].normal.z = 0
	volume_2.planes[3].d = -Meter
	// bottom
	volume_2.planes[4].normal.x = 0
	volume_2.planes[4].normal.y = 0
	volume_2.planes[4].normal.z = Meter
	volume_2.planes[4].d = 0
	// top
	volume_2.planes[5].normal.x = 0
	volume_2.planes[5].normal.y = 0
	volume_2.planes[5].normal.z = -Meter
	volume_2.planes[5].d = -Meter

	zone := Zone{}

	zone.volumes = make([]Volume, 2)
	zone.volumes[0] = volume_1
	zone.volumes[1] = volume_2

	for index, parameter := range parameters {
		t.Run(fmt.Sprintf("inside_zone_%d", index), func(t *testing.T) {
			assert.Equal(t, parameter.out, zone.Inside(parameter.in.x, parameter.in.y, parameter.in.z))
		})
	}
}

func Test_Grid_World(t *testing.T) {

	world := generateWorld_Grid(10, 10, 10, Meter)

	_ = world

	for k := 0; k < 10; k++ {
		for j := 0; j < 10; j++ {
			for i := 0; i < 10; i++ {
				// ...
			}
		}
	}
}

func Test_FindZoneId_World(t *testing.T) {

	type In struct {
		i        int64
		j        int64
		k        int64
		cellSize int64
		position Vector
	}

	type Out struct {
		found   bool
		zone_id uint32
	}

	type Params struct {
		in  In
		out Out
	}

	var parameters = []Params{
		{In{2, 2, 2, Meter, Vector{Meter, Meter, Meter}}, Out{true, 1}},
		{In{2, 2, 2, Meter, Vector{-Meter, -Meter, -Meter}}, Out{false, 0}},
	}

	for index, parameter := range parameters {
		t.Run(fmt.Sprintf("inside_volume_%d", index), func(t *testing.T) {
			world := generateWorld_Grid(parameter.in.i, parameter.in.j, parameter.in.k, parameter.in.cellSize)
			var zone_id uint32
			found := world.FindZoneId(parameter.in.position.x, parameter.in.position.y, parameter.in.position.z, &zone_id)
			assert.Equal(t, parameter.out.found, found)
			assert.Equal(t, parameter.out.zone_id, zone_id)
		})
	}
}
