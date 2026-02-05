// Provides functions to convert from proto buf types to customtypes and vice versa
package mapping

import (
	pb "social-network/shared/gen-go/media"
	ct "social-network/shared/go/ct"
)

// pbToCtFileVariant converts protobuf ImgVariant to customtypes ImgVariant
func PbToCtFileVariant(v pb.FileVariant) ct.FileVariant {
	switch v {
	case pb.FileVariant_THUMBNAIL:
		return ct.ImgThumbnail
	case pb.FileVariant_SMALL:
		return ct.ImgSmall
	case pb.FileVariant_MEDIUM:
		return ct.ImgMedium
	case pb.FileVariant_LARGE:
		return ct.ImgLarge
	case pb.FileVariant_ORIGINAL:
		return ct.Original
	default:
		return ct.FileVariant("") // invalid, but handle gracefully
	}
}

// ctToPbFileVariant converts customtypes FileVariant to protobuf FileVariant
func CtToPbFileVariant(v ct.FileVariant) pb.FileVariant {
	switch v {
	case ct.ImgThumbnail:
		return pb.FileVariant_THUMBNAIL
	case ct.ImgSmall:
		return pb.FileVariant_SMALL
	case ct.ImgMedium:
		return pb.FileVariant_MEDIUM
	case ct.ImgLarge:
		return pb.FileVariant_LARGE
	case ct.Original:
		return pb.FileVariant_ORIGINAL
	default:
		return pb.FileVariant_IMG_VARIANT_UNSPECIFIED
	}
}

// pbToCtFileVisibility converts protobuf FileVisibility to customtypes FileVisibility
func PbToCtFileVisibility(v pb.FileVisibility) ct.FileVisibility {
	switch v {
	case pb.FileVisibility_PRIVATE:
		return ct.Private
	case pb.FileVisibility_PUBLIC:
		return ct.Public
	default:
		return ct.FileVisibility("") // invalid
	}
}

// ctToPbFileVisibility converts customtypes FileVisibility to protobuf FileVisibility
func CtToPbFileVisibility(v ct.FileVisibility) pb.FileVisibility {
	switch v {
	case ct.Private:
		return pb.FileVisibility_PRIVATE
	case ct.Public:
		return pb.FileVisibility_PUBLIC
	default:
		return pb.FileVisibility_FILE_VISIBILITY_UNSPECIFIED
	}
}

// ctToPbUploadStatus converts customtypes UploadStatus to protobuf UploadStatus
func CtToPbUploadStatus(v ct.UploadStatus) pb.UploadStatus {
	switch v {
	case ct.Pending:
		return pb.UploadStatus_UPLOAD_STATUS_PENDING
	case ct.Processing:
		return pb.UploadStatus_UPLOAD_STATUS_PROCESSING
	case ct.Complete:
		return pb.UploadStatus_UPLOAD_STATUS_COMPLETE
	case ct.Failed:
		return pb.UploadStatus_UPLOAD_STATUS_FAILED
	default:
		return pb.UploadStatus_UPLOAD_STATUS_UNSPECIFIED
	}
}

// ctToPbUploadStatus converts protobuf UploadStatus to customtypes UploadStatus
func PbToCtUploadStatus(v pb.UploadStatus) ct.UploadStatus {
	switch v {
	case pb.UploadStatus_UPLOAD_STATUS_PENDING:
		return ct.Pending
	case pb.UploadStatus_UPLOAD_STATUS_PROCESSING:
		return ct.Processing
	case pb.UploadStatus_UPLOAD_STATUS_COMPLETE:
		return ct.Complete
	case pb.UploadStatus_UPLOAD_STATUS_FAILED:
		return ct.Failed
	default:
		return ct.UploadStatus("") // invalid

	}
}
