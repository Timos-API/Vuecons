package service

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Vuecon struct {
	VueconID     string `json:"id"`
	LastModified int64  `json:"last_modified"`
	Size         int64  `json:"size"`
}

type VueconsService struct {
	aws    *s3.Client
	bucket *string
}

func NewVueconsService(awsClient *s3.Client, bucket string) *VueconsService {
	return &VueconsService{awsClient, aws.String(bucket)}
}

func (s *VueconsService) GetAll(ctx context.Context) (*[]Vuecon, error) {

	output, err := s.aws.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: s.bucket})

	if err != nil {
		return nil, err
	}

	vuecons := []Vuecon{}

	for _, obj := range output.Contents {
		vuecon := Vuecon{
			VueconID:     *obj.Key,
			LastModified: obj.LastModified.UnixNano() / 1000000,
			Size:         obj.Size,
		}
		vuecons = append(vuecons, vuecon)
	}

	return &vuecons, nil
}

func (s *VueconsService) Upload(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (*Vuecon, error) {

	if !strings.HasSuffix(fileHeader.Filename, ".svg") {
		return nil, errors.New("invalid file type: only svg files allowed")
	}

	buffer := make([]byte, fileHeader.Size)
	file.Read(buffer)

	_, err := s.aws.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               s.bucket,
		Key:                  &fileHeader.Filename,
		ACL:                  "public-read",
		Body:                 bytes.NewReader(buffer),
		ContentLength:        fileHeader.Size,
		ContentType:          aws.String("image/svg+xml"),
		ServerSideEncryption: "AES256",
	})

	if err != nil {
		return nil, err
	}

	vuecon := Vuecon{VueconID: fileHeader.Filename, LastModified: time.Now().UnixNano() / 1000000, Size: fileHeader.Size}

	return &vuecon, nil
}

func (s *VueconsService) Delete(ctx context.Context, id string) (bool, error) {

	_, err := s.aws.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(id),
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *VueconsService) Get(ctx context.Context, id string) (*Vuecon, error) {

	output, err := s.aws.GetObject(ctx, &s3.GetObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(id),
	})

	if err != nil {
		return nil, err
	}

	vuecon := Vuecon{
		VueconID:     id,
		LastModified: output.LastModified.UnixNano() / 1000000,
		Size:         output.ContentLength,
	}

	return &vuecon, nil
}
