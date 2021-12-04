package imagedrm

import (
	"image"
	"image/color"
	"image/draw"
	"unsafe"
)

func (i *Image) ColorModel() color.Model {
	return &image.Uniform{}
}

func (i *Image) Bounds() image.Rectangle {
	display := i.displays[0]

	mode := display.mode
	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: int(mode.Width),
			Y: int(mode.Height),
		},
	}
}

func (i *Image) At(x, y int) color.Color {
	display := i.displays[0]

	offset := (display.fb.Pitch * uint32(y)) + (uint32(x) * 4)
	val := *(*uint32)(unsafe.Pointer(&display.fb.data[offset]))

	return color.RGBA{
		A: uint8((val & 0xff000000) >> 24),
		R: uint8((val & 0x00ff0000) >> 16),
		G: uint8((val & 0x0000ff00) >> 8),
		B: uint8(val & 0x000000ff),
	}
}

func (i *Image) Set(x, y int, c color.Color) {
	display := i.displays[0]

	r, g, b, a := c.RGBA()
	val := (a << 24) | (r << 16) | (g << 8) | b

	offset := (display.fb.Pitch * uint32(y)) + (uint32(x) * 4)
	*(*uint32)(unsafe.Pointer(&display.fb.data[offset])) = val
}

var _ draw.Image = (*Image)(nil)
