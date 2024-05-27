package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Dot(t *testing.T) {

	nx := Meter
	ny := int64(0)
	nz := int64(0)

	dot := Dot(Meter, 0, 0, nx, ny, nz)

	assert.Equal(t, dot, Meter)

	dot = Dot(-Meter, 0, 0, nx, ny, nz)

	assert.Equal(t, dot, -Meter)

	nx = -Meter

	dot = Dot(Meter, 0, 0, nx, ny, nz)
	
	assert.Equal(t, dot, -Meter)

	dot = Dot(-Meter, 0, 0, nx, ny, nz)
	
	assert.Equal(t, dot, Meter)
}

func Test_Inside_Volume(t *testing.T) {

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

    // check inside cases

	assert.True(t, volume.Inside(0,0,0))
	assert.True(t, volume.Inside(Meter/2,Meter/2,Meter/2))
	assert.True(t, volume.Inside(Meter,Meter,Meter))

	// check outside cases

	assert.False(t, volume.Inside(-1,0,0))
	assert.False(t, volume.Inside(0,-1,0))
	assert.False(t, volume.Inside(0,0,-1))
	assert.False(t, volume.Inside(Meter*2,Meter*2,Meter*2))
}
