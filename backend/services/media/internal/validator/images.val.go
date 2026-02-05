package validator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"social-network/services/media/internal/configs"
	ce "social-network/shared/go/commonerrors"
)

var (
	ErrImageValidator   = errors.New("image validator")
	ErrInvalidImage     = errors.New("invalid image")
	ErrImageTooLarge    = errors.New("image exceeds size limit")
	ErrUnsupportedType  = errors.New("unsupported image type")
	ErrInvalidDimension = errors.New("invalid image dimensions")
)

type ImageValidator struct {
	Config configs.FileConstraints
}

// ValidateImage validates an uploaded image according to the configured constraints.
//
// This function performs multiple layers of checks to ensure the image is safe, supported,
// and within allowed limits. It takes a context and an io.Reader for the image data.
//
// Validation Steps:
//
// 1️⃣ Enforce maximum upload size:
//   - Wraps the input reader with io.LimitReader to prevent reading more than the allowed size + 1 byte.
//   - Reads the entire content into memory (buf).
//   - If the image exceeds MaxImageUpload, it returns ErrImageTooLarge.
//
// 2️⃣ Detect MIME type by content:
//   - Uses the first 512 bytes to detect the MIME type (http.DetectContentType).
//   - This prevents trusting file extensions which can be spoofed.
//   - If the MIME type is not allowed (not in AllowedMIMEs), returns ErrUnsupportedType.
//
// 3️⃣ Decode image config only:
//   - Uses image.DecodeConfig to quickly parse the image metadata without fully decoding pixel data.
//   - Retrieves width, height, and format for validation.
//   - If decoding fails, returns ErrInvalidImage.
//
// 4️⃣ Validate image dimensions:
//   - Ensures width and height are greater than zero.
//   - Ensures dimensions do not exceed MaxWidth or MaxHeight to avoid decompression bombs.
//   - Returns ErrInvalidDimension if the image is too small, zero-sized, or too large.
//
// 5️⃣ Optional: restrict image formats:
//   - Checks the decoded format (jpeg, png, gif, etc.) against AllowedMIMEs.
//   - Prevents images in unsupported formats from being accepted.
//   - Returns ErrUnsupportedType if the format is disallowed.
//
// 6️⃣ Fully decode the image:
//   - Decodes the entire image to ensure it is not corrupted and is a valid image.
//   - If decoding fails, returns ErrInvalidImage.
//
// Returns:
//   - nil if the image passes all checks.
//   - A wrapped error containing ErrImageValidator and a specific error type otherwise.
//
// Notes:
//   - This function reads the entire image into memory, which may be a concern for very large images.
//   - The function is defensive, designed to prevent invalid or malicious images from being processed.
func (v *ImageValidator) ValidateImage(ctx context.Context, r io.Reader) error {

	// 1️⃣ Enforce max upload size
	limited := io.LimitReader(r, v.Config.MaxImageUpload+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return ce.Wrap(
			ce.ErrInternal,
			fmt.Errorf("read failed: %w", err),
			"limit reader")
	}

	if int64(len(buf)) > v.Config.MaxImageUpload {
		return ce.Wrap(ce.ErrPermissionDenied,
			fmt.Errorf("%w: actual: %d, max allowed: %d", ErrImageTooLarge, len(buf), v.Config.MaxImageUpload),
		).WithPublic(fmt.Sprintf("image too large. Max size %v", v.Config.MaxImageUpload))
	}

	// 2️⃣ Detect MIME type by content (NOT filename)
	mime := http.DetectContentType(buf[:min(512, len(buf))])
	if !v.Config.AllowedMIMEs[mime] {
		return ce.Wrap(
			ce.ErrPermissionDenied,
			fmt.Errorf("%w: %s", ErrUnsupportedType, mime),
			"detect content type",
		).WithPublic(fmt.Sprintf("unsuported mime type %v", mime))
	}

	// 3️⃣ Decode config only (fast, safe)
	cfg, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		return ce.Wrap(
			ce.ErrPermissionDenied,
			ErrInvalidImage,
			"first pass decode",
		).WithPublic("unsupported file type")
	}

	// 4️⃣ Dimension validation (prevents decompression bombs)
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return ce.Wrap(
			ce.ErrPermissionDenied,
			ErrInvalidDimension,
			"decompression bomb check",
		).WithPublic("invalid dimensions")
	}

	if cfg.Width > v.Config.MaxWidth || cfg.Height > v.Config.MaxHeight {
		return ce.Wrap(ce.ErrPermissionDenied,
			fmt.Errorf(
				"%w: %dx%d",
				ErrInvalidDimension,
				cfg.Width,
				cfg.Height,
			),
			"dimentions check",
		).WithPublic("invalid dimensions")
	}

	// 5️⃣ Optional: restrict formats (jpeg/png/gif/etc)
	if !v.Config.AllowedMIMEs["image/"+format] {
		return ce.Wrap(
			ce.ErrPermissionDenied,
			fmt.Errorf("%w: %s", ErrUnsupportedType, format),
			"allowed mimes check",
		).WithPublic(fmt.Sprintf("unsuported mime type %v", format))
	}

	// 6️⃣ Fully decode to ensure the image is valid
	_, _, err = image.Decode(bytes.NewReader(buf))
	if err != nil {
		return ce.Wrap(ce.ErrPermissionDenied,
			ErrInvalidImage,
			"second pass decode",
		).WithPublic("invalid file type")
	}

	return nil
}
