# Media Service Documentation

## Overview
The Media Service is a Go-based microservice that handles image upload, storage, processing, and retrieval for a social network application. It provides a gRPC API for managing images, using MinIO for object storage and PostgreSQL for metadata persistence.

## Architecture
- **Language**: Go
- **Storage**: MinIO (S3-compatible object storage)
- **Database**: PostgreSQL
- **API**: gRPC (defined in media.proto)
- **Deployment**: Docker containerized

## Core Components

### 1. Entry Point (main.go)
Starts the gRPC server with database connection, MinIO client initialization, and background workers.

### 2. Application Layer (`internal/application/`)
- **MediaService**: Main business logic orchestrator
- **UploadImage**: Creates file metadata, generates pre-signed upload URLs, and schedules variant creation
- **GetImage/GetImages**: Provides pre-signed download URLs for images and variants
- **ValidateUpload**: Verifies uploaded files against constraints and marks as complete
- **Variant Worker**: Background process that generates image variants asynchronously

### 3. Handler Layer (`internal/handler/`)
gRPC method implementations that convert protobuf messages to internal types and call application logic.

### 4. Client Layer (`internal/client/`)
MinIO integration for:
- Generating pre-signed upload/download URLs
- File validation and tagging
- Variant generation via image conversion

### 5. Validator (`internal/validator/`)
Image validation ensuring:
- Size limits (max 5MB) (*configurable*)
- Supported formats (JPEG, PNG, GIF, WebP)
- Dimension constraints (max 4096x4096) (*configurable*)
- Content integrity

### 6. Convertor (`convertor/`)
Image processing for variant generation:
- Resizes images to predefined sizes (thumbnail: 150x150, small: 400x400, medium: 800x800, large: 1600x1600)
- Converts to WebP format with 80% quality
- Maintains aspect ratio

## API Methods

### UploadImage
- **Input**: filename, mime_type, size_bytes, visibility, expiration_seconds, variants[]
- **Output**: file_id, upload_url
- **Behavior**: Creates database entries for original and requested variants, returns pre-signed upload URL

### GetImage
- **Input**: image_id, variant
- **Output**: download_url
- **Behavior**: Returns pre-signed download URL, falls back to original if variant unavailable

### GetImages (Batch)
- **Input**: img_ids[], variant
- **Output**: download_urls map, failed_ids[]
- **Behavior**: Batch retrieval for multiple images, excludes original variant

### ValidateUpload
- **Input**: file_id
- **Output**: Empty
- **Behavior**: Validates uploaded file, sets status to complete, tags as validated in MinIO

## Data Flow

1. **Upload Request**: Client calls UploadImage → receives file_id and upload_url
2. **File Upload**: Client uploads directly to MinIO using pre-signed URL
3. **Validation**: Client calls ValidateUpload → service validates file and marks complete
4. **Variant Generation**: Background worker processes pending variants asynchronously
5. **Retrieval**: Client calls GetImage/GetImages → receives download URLs

## Storage Buckets
- **uploads-originals**: Raw uploaded images
- **uploads-variants**: Processed image variants

## Background Workers
- **Variant Worker**: Runs every 30 seconds (*configurable*), generates pending image variants
- **Stale Files Worker**: Runs every 1 hour (*configurable*), cleans up unvalidated files older than 1 day.

## Security Features
- Pre-signed URLs with expiration
- File validation before marking complete
- Automatic cleanup of unvalidated uploads (24-hour lifecycle)
- Content-type and size validation
- Dimension limits to prevent decompression bombs

## Configuration
Environment variables:
- `SERVICE_PORT`: gRPC server port
- `DATABASE_URL`: PostgreSQL connection string
- `MINIO_ENDPOINT`: MinIO server URL
- `MINIO_PUBLIC_ENDPOINT`: Public MinIO URL for URL generation (*only on dev mode*)
- `MINIO_ACCESS_KEY`/`MINIO_SECRET_KEY`: MinIO credentials

## Usage Example
```go
// Upload image
uploadResp, err := mediaClient.UploadImage(ctx, &media.UploadImageRequest{
    Filename: "avatar.jpg",
    MimeType: "image/jpeg",
    SizeBytes: 1024000,
    Visibility: media.FileVisibility_PUBLIC,
    Variants: []media.FileVariant{media.FileVariant_THUMBNAIL, media.FileVariant_LARGE},
    ExpirationSeconds: 600,
})

// Upload file to upload_url
// Then validate
_, err = mediaClient.ValidateUpload(ctx, &media.ValidateUploadRequest{FileId: uploadResp.FileId})

// Get image
getResp, err := mediaClient.GetImage(ctx, &media.GetImageRequest{
    ImageId: uploadResp.FileId,
    Variant: media.FileVariant_THUMBNAIL,
})
```