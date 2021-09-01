package service

import (
	"Timos-API/Vuecons/persistence"
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/Timos-API/transformer"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
)

type VueconsService struct {
	p      *persistence.VueconPersistor
	aws    *s3.Client
	bucket *string
}

func NewVueconsService(p *persistence.VueconPersistor, awsClient *s3.Client, bucket string) *VueconsService {
	return &VueconsService{p, awsClient, aws.String(bucket)}
}

func (s *VueconsService) getUrl(key string) string {
	return fmt.Sprintf("https://vuecons.s3.%s.amazonaws.com/%s", os.Getenv("AWS_REGION"), key)
}

func (s *VueconsService) GetAll(ctx context.Context) (*[]persistence.Vuecon, error) {
	return s.p.GetAll(ctx)
}

func (s *VueconsService) Upload(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (*persistence.Vuecon, error) {

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

	vuecon := persistence.Vuecon{
		Name:         fileHeader.Filename,
		LastModified: time.Now().UnixNano() / 1000000,
		Size:         fileHeader.Size,
		Src:          s.getUrl(fileHeader.Filename),
		Tags:         make([]string, 0),
		Categories:   make([]string, 0),
	}

	return s.p.Create(ctx, vuecon)
}

func (s *VueconsService) Delete(ctx context.Context, id string) (bool, error) {

	success, err := s.p.Delete(ctx, id)

	if !success && err != nil {
		return false, err
	}

	_, err = s.aws.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(id),
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *VueconsService) Get(ctx context.Context, id string) (*persistence.Vuecon, error) {
	return s.p.GetById(ctx, id)
}

func (s *VueconsService) Update(ctx context.Context, vuecon persistence.Vuecon) (*persistence.Vuecon, error) {
	validate := validator.New()
	err := validate.Struct(vuecon)
	if err != nil {
		return nil, err
	}

	cleaned := transformer.Clean(vuecon, "update")
	return s.p.Update(ctx, vuecon.Name, cleaned)
}
