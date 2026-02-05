package convertor

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"math"
	"social-network/services/media/internal/configs"
	ct "social-network/shared/go/ct"

	"github.com/chai2010/webp"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

type ImageConvertor struct {
	Configs configs.FileConstraints
}

func NewImageconvertor(c configs.FileConstraints) *ImageConvertor {
	return &ImageConvertor{
		Configs: c,
	}
}

// ConvertImageToVariant reads an image from r, resizes it according to the specified variant,
// and encodes it as a WebP image. Variants control the target width and height (e.g., large, medium, small, thumbnail).
// Returns a bytes.Buffer containing the converted image or an error if reading, decoding, resizing,
// or encoding fails. Ensures the input does not exceed the maximum allowed upload size.
func (i *ImageConvertor) ConvertImageToVariant(
	buf []byte, variant ct.FileVariant,
) (out bytes.Buffer, err error) {

	if int64(len(buf)) > i.Configs.MaxImageUpload {
		return out, fmt.Errorf("image size exceeds limit")
	}

	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return out, fmt.Errorf("failed to decode image: %w", err)
	}

	img, err = decodeWithOrientation(buf)
	if err != nil {
		return out, fmt.Errorf("failed to decode with orientation %w", err)
	}

	resized := resizeForVariant(img, variant)

	if err := webp.Encode(&out, resized, &webp.Options{Quality: 80}); err != nil {
		return out, err
	}
	return out, nil
}

// Extracts orientation meta data and applies it to image
func decodeWithOrientation(buf []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	ex, err := exif.Decode(bytes.NewReader(buf))
	if err != nil {
		return img, nil // no EXIF → assume correct
	}

	tag, err := ex.Get(exif.Orientation)
	if err != nil {
		return img, nil
	}

	orientation, err := tag.Int(0)
	if err != nil {
		return img, nil
	}

	return applyOrientation(img, orientation), nil
}

func applyOrientation(img image.Image, orientation int) image.Image {
	b := img.Bounds()
	w, h := float64(b.Dx()), float64(b.Dy())

	var (
		dstRect image.Rectangle
		m       f64.Aff3
	)

	switch orientation {
	case 1:
		return img

	case 2: // Flip horizontal
		dstRect = image.Rect(0, 0, int(w), int(h))
		m = f64.Aff3{
			-1, 0, w,
			0, 1, 0,
		}

	case 3: // Rotate 180
		dstRect = image.Rect(0, 0, int(w), int(h))
		m = f64.Aff3{
			-1, 0, w,
			0, -1, h,
		}

	case 4: // Flip vertical
		dstRect = image.Rect(0, 0, int(w), int(h))
		m = f64.Aff3{
			1, 0, 0,
			0, -1, h,
		}

	case 5: // Transpose
		dstRect = image.Rect(0, 0, int(h), int(w))
		m = f64.Aff3{
			0, 1, 0,
			1, 0, 0,
		}

	case 6: // Rotate 90 CW
		dstRect = image.Rect(0, 0, int(h), int(w))
		m = f64.Aff3{
			0, -1, h,
			1, 0, 0,
		}

	case 7: // Transverse
		dstRect = image.Rect(0, 0, int(h), int(w))
		m = f64.Aff3{
			0, -1, h,
			-1, 0, w,
		}

	case 8: // Rotate 90 CCW
		dstRect = image.Rect(0, 0, int(h), int(w))
		m = f64.Aff3{
			0, 1, 0,
			-1, 0, w,
		}

	default:
		return img
	}

	return transformImage(img, dstRect, m)
}

func transformImage(
	src image.Image,
	dstRect image.Rectangle,
	m f64.Aff3,
) image.Image {
	dst := image.NewRGBA(dstRect)

	draw.ApproxBiLinear.Transform(
		dst,
		m,
		src,
		src.Bounds(),
		draw.Over,
		nil,
	)

	return dst
}

// Returns and image.Image object resized to the variant specs. If the image dimentions
// are smaller that the variant specs then the object is returned as is.
func resizeForVariant(src image.Image, variant ct.FileVariant) image.Image {
	maxWidth, maxHeight := variantToSize(variant)
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= maxWidth && h <= maxHeight {
		return src
	}

	ratioW := float64(maxWidth) / float64(w)
	ratioH := float64(maxHeight) / float64(h)
	ratio := math.Min(ratioW, ratioH)

	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))

	draw.CatmullRom.Scale(
		dst,
		dst.Bounds(),
		src,
		bounds,
		draw.Over,
		nil,
	)

	return dst
}

func variantToSize(variant ct.FileVariant) (maxWidth, maxHeight int) {
	switch variant {
	case ct.ImgLarge:
		return 1600, 1600

	case ct.ImgMedium:
		return 800, 800

	case ct.ImgSmall:
		return 400, 400

	case ct.ImgThumbnail:
		return 150, 150

	default:
		// fallback (treat as medium)
		return 800, 800
	}
}
