package dbservice

import ct "social-network/shared/go/ct"

// General file meta struct mirroring the files table. In some cases it refers to a file's variant
// when the Variant field is not missing or marked as Original.
type File struct {
	Id        ct.Id  `validation:"nullable"` // db row Id. File Id or variant Id
	Filename  string // the original name given by sender
	MimeType  string // content type
	SizeBytes int64
	Bucket    string // images, videos etc
	ObjectKey string // the name given to file in fileservice

	Visibility ct.FileVisibility
	Status     ct.UploadStatus // pending, processing, complete, failed

	Variant ct.FileVariant `validation:"nullable"` // thumb, small, medium, large, original
}

// Refers to file_variants table. It contains fields that are joined from file table row
// when retrieving a variant from db.
type Variant struct {
	Id        ct.Id  `validation:"nullable"` // variant Id
	FileId    ct.Id  // reference to files table id col.
	Filename  string // the original name given by sender
	MimeType  string // content type
	SizeBytes int64
	Bucket    string          // images, videos etc
	ObjectKey string          // the name given to file in fileservice
	Status    ct.UploadStatus // pending, processing, complete, failed
	Variant   ct.FileVariant  // thumb, small, medium, large, original

	// Joined from files table coresponding row.
	Visibility   ct.FileVisibility
	SrcBucket    string // the variants origin bucket
	SrcObjectKey string // the variants origin key

}
