package mapping

import (
	"social-network/services/media/internal/db/dbservice"
	md "social-network/services/media/internal/models"
	"social-network/shared/go/ct"
)

// Converts models.FileMeta to dbservice.File
func ModelToDbFile(meta md.FileMeta) dbservice.File {
	return dbservice.File{
		Id:         meta.Id,
		Filename:   meta.Filename,
		MimeType:   meta.MimeType,
		SizeBytes:  meta.SizeBytes,
		Bucket:     meta.Bucket,
		ObjectKey:  meta.ObjectKey,
		Visibility: meta.Visibility,
		Variant:    meta.Variant,
	}
}

// Converts models.FileMeta to dbservice.File with status
func ModelToDbFileWithStatus(meta md.FileMeta, status ct.UploadStatus) dbservice.File {
	f := ModelToDbFile(meta)
	f.Status = status
	return f
}

// Converts dbservice.File to models.FileMeta
func DbToModel(file dbservice.File) md.FileMeta {
	return md.FileMeta{
		Id:         file.Id,
		Filename:   file.Filename,
		MimeType:   file.MimeType,
		SizeBytes:  file.SizeBytes,
		Bucket:     file.Bucket,
		ObjectKey:  file.ObjectKey,
		Visibility: file.Visibility,
		Variant:    file.Variant,
	}
}
