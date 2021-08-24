package service

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Vuecon struct {
	VueconID     string `json:"id"`
	LastModified int64  `json:"last_modified"`
	Size         int64  `json:"size"`
}

type VueconsService struct {
	aws    *s3.S3
	bucket *string
}

func NewVueconsService(awsS3 *s3.S3, bucket string) *VueconsService {
	return &VueconsService{awsS3, aws.String(bucket)}
}

func (s *VueconsService) GetAllVuecons(ctx context.Context) (*[]Vuecon, error) {

	output, err := s.aws.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{Bucket: s.bucket})

	if err != nil {
		return nil, err
	}

	vuecons := make([]Vuecon, 0, len(output.Contents))

	for _, obj := range output.Contents {
		vuecon := Vuecon{
			VueconID:     *obj.Key,
			LastModified: obj.LastModified.UnixNano() / 1000000,
			Size:         *obj.Size,
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

	_, err := s.aws.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:               s.bucket,
		Key:                  &fileHeader.Filename,
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(fileHeader.Size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ServerSideEncryption: aws.String("AES256"),
	})

	if err != nil {
		return nil, err
	}

	vuecon := Vuecon{VueconID: fileHeader.Filename, LastModified: time.Now().UnixNano() / 1000000, Size: fileHeader.Size}

	return &vuecon, nil
}

func (s *VueconsService) Delete(ctx context.Context, id string) (bool, error) {

	_, err := s.aws.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(id),
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
