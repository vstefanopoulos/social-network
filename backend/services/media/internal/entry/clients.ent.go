package entry

import (
	"context"
	"reflect"
	"social-network/services/media/internal/configs"
	tele "social-network/shared/go/telemetry"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
)

func NewMinIOConn(ctx context.Context, cfgs configs.FileService, endpoint string, skipBucketCreation bool) (*minio.Client, error) {
	var minioClient *minio.Client
	var err error

	accessKey := cfgs.AccessKey
	secret := cfgs.Secret

	for range 10 {
		minioClient, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secret, ""),
			Secure: false,
			Region: "us-east-1",
		})
		if err == nil {
			break
		}
		tele.Warn(ctx, "MinIO not ready, retrying in 2s...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	tele.Info(ctx, "Connected to minio client")

	if skipBucketCreation {
		return minioClient, nil
	}

	//TODO check if using ctx from entry is ok
	// Ensure bucket exists
	if err := EnsureBuckets(ctx,
		minioClient, cfgs.Buckets); err != nil {
		return nil, err
	}

	tele.Info(ctx, "Setting up lifecycle rules")

	lcfg := lifecycle.NewConfiguration()

	rule := lifecycle.Rule{
		ID:     "delete-unvalidated",
		Status: "Enabled",
		RuleFilter: lifecycle.Filter{
			Tag: lifecycle.Tag{
				Key:   "validated",
				Value: "false",
			},
		},
		Expiration: lifecycle.Expiration{
			Days: lifecycle.ExpirationDays(1),
		},
	}

	lcfg.Rules = append(lcfg.Rules, rule)

	err = minioClient.SetBucketLifecycle(ctx, cfgs.Buckets.Originals, lcfg)
	if err != nil {
		tele.Error(ctx, "Error setting lifecycle. @1", "error", err.Error())
		// We might still continue
	}

	return minioClient, nil
}

func EnsureBuckets(ctx context.Context, client *minio.Client, buckets configs.Buckets) error {
	v := reflect.ValueOf(buckets)
	tele.Info(ctx, "Creating buckets")
	for i := 0; i < v.NumField(); i++ {
		bucketName := v.Field(i).String()

		if bucketName == "" {
			continue
		}

		exists, err := client.BucketExists(ctx, bucketName)
		if err != nil {
			return err
		}

		if !exists {
			tele.Info(ctx, "Creating bucket. @1", "name", bucketName)
			if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
				return err
			}
		}
	}
	tele.Info(ctx, "Buckets created!")

	return nil
}
