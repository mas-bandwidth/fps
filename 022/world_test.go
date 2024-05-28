package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
