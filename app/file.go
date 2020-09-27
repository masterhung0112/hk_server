package app

import (
	"github.com/disintegration/imaging"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/services/filesstore"
	"github.com/rwcarlsen/goexif/exif"
	"image"
	"io"
)

const (
	/*
	  EXIF Image Orientations
	  1        2       3      4         5            6           7          8

	  888888  888888      88  88      8888888888  88                  88  8888888888
	  88          88      88  88      88  88      88  88          88  88      88  88
	  8888      8888    8888  8888    88          8888888888  8888888888          88
	  88          88      88  88
	  88          88  888888  888888
	*/
	Upright            = 1
	UprightMirrored    = 2
	UpsideDown         = 3
	UpsideDownMirrored = 4
	RotatedCWMirrored  = 5
	RotatedCCW         = 6
	RotatedCCWMirrored = 7
	RotatedCW          = 8

	MaxImageSize         = int64(6048 * 4032) // 24 megapixels, roughly 36MB as a raw image
	ImageThumbnailWidth  = 120
	ImageThumbnailHeight = 100
	ImageThumbnailRatio  = float64(ImageThumbnailHeight) / float64(ImageThumbnailWidth)
	ImagePreviewWidth    = 1920

	maxUploadInitialBufferSize = 1024 * 1024 // 1Mb

	// Deprecated
	IMAGE_THUMBNAIL_PIXEL_WIDTH  = 120
	IMAGE_THUMBNAIL_PIXEL_HEIGHT = 100
	IMAGE_PREVIEW_PIXEL_WIDTH    = 1920
)

func (a *App) FileBackend() (filesstore.FileBackend, *model.AppError) {
	return a.Srv().FileBackend()
}

func (a *App) WriteFile(fr io.Reader, path string) (int64, *model.AppError) {
	backend, err := a.FileBackend()
	if err != nil {
		return 0, err
	}

	return backend.WriteFile(fr, path)
}

func makeImageUpright(img image.Image, orientation int) image.Image {
	switch orientation {
	case UprightMirrored:
		return imaging.FlipH(img)
	case UpsideDown:
		return imaging.Rotate180(img)
	case UpsideDownMirrored:
		return imaging.FlipV(img)
	case RotatedCWMirrored:
		return imaging.Transpose(img)
	case RotatedCCW:
		return imaging.Rotate270(img)
	case RotatedCCWMirrored:
		return imaging.Transverse(img)
	case RotatedCW:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

func getImageOrientation(input io.Reader) (int, error) {
	exifData, err := exif.Decode(input)
	if err != nil {
		return Upright, err
	}

	tag, err := exifData.Get("Orientation")
	if err != nil {
		return Upright, err
	}

	orientation, err := tag.Int(0)
	if err != nil {
		return Upright, err
	}

	return orientation, nil
}
