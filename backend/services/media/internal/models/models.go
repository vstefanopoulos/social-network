package models

import ct "social-network/shared/go/ct"

type FileMeta struct {
	Id        ct.Id  // db row Id
	Filename  string // the original name given by sender
	MimeType  string // content type
	SizeBytes int64
	Bucket    string // images, videos etc
	ObjectKey string // the name given to file in fileservice

	Visibility ct.FileVisibility // public, private
	Variant    ct.FileVariant    // thumb, small, medium, large, original
}
