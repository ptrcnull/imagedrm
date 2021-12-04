package imagedrm

import (
	"fmt"
	"github.com/kytart/godrm/pkg/drm"
	"github.com/kytart/godrm/pkg/mode"
	"launchpad.net/gommap"
	"os"
)

type Framebuffer struct {
	*mode.FB
	id   uint32
	data []byte
}

type Image struct {
	file     *os.File
	modeset  *mode.SimpleModeset
	displays []*Display
}

type Display struct {
	mode      *mode.Modeset
	fb        *Framebuffer
	savedCrtc *mode.Crtc
}

func NewImage() (*Image, error) {
	file, err := drm.OpenCard(0)
	if err != nil {
		return nil, fmt.Errorf("open drm card: %w", err)
	}
	defer file.Close()

	if !drm.HasDumbBuffer(file) {
		return nil, fmt.Errorf("drm device does not support dumb buffers")
	}

	modeset, err := mode.NewSimpleModeset(file)
	if err != nil {
		return nil, fmt.Errorf("create modeset: %w", err)
	}

	image := &Image{
		file:    file,
		modeset: modeset,
	}

	for _, mod := range modeset.Modesets {
		display, err := image.setupDisplay(mod)
		if err != nil {
			image.Close()
			return nil, fmt.Errorf("setup display: %w", err)
		}

		image.displays = append(image.displays, display)
	}

	return image, nil
}

func (i *Image) Close() error {
	var err error
	for _, display := range i.displays {
		err = i.destroyFramebuffer(display)
	}
	return err
}

func (i *Image) createFramebuffer(dev *mode.Modeset) (*Framebuffer, error) {
	fb, err := mode.CreateFB(i.file, dev.Width, dev.Height, 32)
	if err != nil {
		return nil, fmt.Errorf("create framebuffer: %w", err)
	}

	fbID, err := mode.AddFB(i.file, dev.Width, dev.Height, 24, 32, fb.Pitch, fb.Handle)
	if err != nil {
		return nil, fmt.Errorf("create dumb buffer: %w", err)
	}

	offset, err := mode.MapDumb(i.file, fb.Handle)
	if err != nil {
		return nil, fmt.Errorf("map dumb: %w", err)
	}

	mmap, err := gommap.MapAt(0, i.file.Fd(), int64(offset), int64(fb.Size), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap framebuffer: %w", err)
	}

	for i := uint64(0); i < fb.Size; i++ {
		mmap[i] = 0
	}

	return &Framebuffer{
		FB:   fb,
		id:   fbID,
		data: mmap,
	}, nil
}

func (i *Image) destroyFramebuffer(display *Display) error {
	err := gommap.MMap(display.fb.data).UnsafeUnmap()
	if err != nil {
		return fmt.Errorf("munmap memory: %w", err)
	}

	err = mode.RmFB(i.file, display.fb.id)
	if err != nil {
		return fmt.Errorf("remove frame buffer: %w", err)
	}

	err = mode.DestroyDumb(i.file, display.fb.Handle)
	if err != nil {
		return fmt.Errorf("destroy dumb buffer: %w", err)
	}

	return i.modeset.SetCrtc(display.mode, display.savedCrtc)
}

func (i *Image) setupDisplay(mod mode.Modeset) (*Display, error) {
	framebuf, err := i.createFramebuffer(&mod)
	if err != nil {
		return nil, fmt.Errorf("create framebuffer: %w", err)
	}

	// save current CRTC of this display to restore at exit
	savedCrtc, err := mode.GetCrtc(i.file, mod.Crtc)
	if err != nil {
		return nil, fmt.Errorf("get CRTC for connector %d: %w", mod.Conn, err)
	}

	// change the display
	err = mode.SetCrtc(i.file, mod.Crtc, framebuf.id, 0, 0, &mod.Conn, 1, &mod.Mode)
	if err != nil {
		return nil, fmt.Errorf("set CRTC for connector %d: %w", mod.Conn, err)
	}

	return &Display{
		mode:      &mod,
		fb:        framebuf,
		savedCrtc: savedCrtc,
	}, nil
}
