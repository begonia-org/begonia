package file_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/file/v1"
	"github.com/minio/minio-go/v7"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
)

var minioBucket = ""
var minioFileAuthor = ""
var sha256Str = ""
var minioUploadId = ""
var minioBigFileSha256 = ""

func newFileMinioBiz() file.FileUsecase {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	cnf := config.ReadConfig(env)
	conf := cfg.NewConfig(cnf)
	repo := data.NewFileRepo(cnf, gateway.Log)

	engines := file.NewFileUsecase(conf, repo)
	return engines[api.FileEngine_FILE_ENGINE_MINIO.String()]
}
func testMinioMkBucket(t *testing.T) {
	fileBiz := newFileMinioBiz()
	snk, _ := tiga.NewSnowflake(1)

	minioFileAuthor = snk.GenerateIDString()
	minioBucket = fmt.Sprintf("bucket-minio-biz-%s", time.Now().Format("20060102150405"))
	c.Convey("test make bucket success", t, func() {
		rsp, err := fileBiz.MakeBucket(context.TODO(), &api.MakeBucketRequest{
			Bucket: minioBucket,
		})
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
	})
	c.Convey("test make bucket fail", t, func() {
		patch := gomonkey.ApplyFuncReturn((*minio.Client).MakeBucket, fmt.Errorf("mkdir error"))
		defer patch.Reset()
		_, err := fileBiz.MakeBucket(context.TODO(), &api.MakeBucketRequest{
			Bucket: "",
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "mkdir error")
	})
}

func testMinioUpload(t *testing.T) {
	fileBiz := newFileMinioBiz()

	c.Convey("test upload file success", t, func() {
		rsp, err := fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
			Bucket:     minioBucket,
			Key:        "test.txt",
			Content:    []byte("hello"),
			UseVersion: true,
		}, minioFileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		shaer := sha256.New()
		shaer.Write([]byte("hello"))
		sha256Str = hex.EncodeToString(shaer.Sum(nil))
	})
	c.Convey("test upload file fail", t, func() {
		patch := gomonkey.ApplyFuncReturn((*minio.Client).PutObject, nil, fmt.Errorf("upload error"))
		defer patch.Reset()
		_, err := fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
			Bucket:  minioBucket,
			Key:     "test.txt",
			Content: []byte("hello"),
		}, minioFileAuthor)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "upload error")
		_, err = fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
			Bucket:  "-13234sdddfe",
			Key:     "test.txt",
			Content: []byte("hello"),
		}, minioFileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not exist")

		patch2 := gomonkey.ApplyFuncReturn(tiga.MySQLDao.Upsert, fmt.Errorf("upsert error"))
		defer patch2.Reset()
		_, err = fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
			Bucket:  minioBucket,
			Key:     "test2.txt",
			Content: []byte("hello2"),
		}, minioFileAuthor)
		patch2.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "failed to upsert file")

	})
}

func testMinioDownload(t *testing.T) {
	fileBiz := newFileMinioBiz()

	c.Convey("test download file success", t, func() {
		_, err := fileBiz.Download(context.TODO(), &api.DownloadRequest{
			Bucket: minioBucket,
			Key:    "test.txt",
		}, minioFileAuthor)
		c.So(err, c.ShouldBeNil)
	})
	c.Convey("test download file fail", t, func() {
		patch := gomonkey.ApplyFuncReturn((*minio.Client).GetObject, nil, fmt.Errorf("download error"))
		defer patch.Reset()
		_, err := fileBiz.Download(context.TODO(), &api.DownloadRequest{
			Bucket: minioBucket,
			Key:    "test.txt",
		}, minioFileAuthor)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "download error")

		patch2 := gomonkey.ApplyFuncReturn((*minio.Object).Read, 0, fmt.Errorf("read error"))
		defer patch2.Reset()
		_, err = fileBiz.Download(context.TODO(), &api.DownloadRequest{
			Bucket: minioBucket,
			Key:    "test.txt",
		}, minioFileAuthor)
		patch2.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "read error")

		patch3 := gomonkey.ApplyFuncReturn((*minio.Object).Stat, nil, fmt.Errorf("stat error"))
		defer patch3.Reset()
		_, err = fileBiz.Download(context.TODO(), &api.DownloadRequest{
			Bucket: minioBucket,
			Key:    "test.txt",
		}, minioFileAuthor)
		patch3.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "stat error")
	})
}
func testMinioMeta(t *testing.T) {
	fileBiz := newFileMinioBiz()
	c.Convey("test get meta success", t, func() {
		rsp, err := fileBiz.Metadata(context.TODO(), &api.FileMetadataRequest{
			Bucket: minioBucket,
			Key:    "test.txt",
			Engine: api.FileEngine_FILE_ENGINE_MINIO.String(),
		}, minioFileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		c.So(rsp.Sha256, c.ShouldEqual, sha256Str)
		c.So(rsp.Version, c.ShouldNotBeEmpty)
	})
	c.Convey("test get meta fail", t, func() {
		patch := gomonkey.ApplyFuncReturn((*minio.Client).StatObject, nil, fmt.Errorf("get meta error"))
		defer patch.Reset()
		_, err := fileBiz.Metadata(context.TODO(), &api.FileMetadataRequest{
			Bucket: minioBucket,
			Key:    "test.txt",
		}, minioFileAuthor)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get meta error")
	})

}
func testMinioVersion(t *testing.T) {
	fileBiz := newFileMinioBiz()
	c.Convey("test get version success", t, func() {
		rsp, err := fileBiz.Version(context.TODO(), minioBucket, "test.txt", minioFileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeEmpty)
	})
	c.Convey("test get version fail", t, func() {
		patch := gomonkey.ApplyFuncReturn((*minio.Client).StatObject, nil, fmt.Errorf("get version error"))
		defer patch.Reset()
		_, err := fileBiz.Version(context.TODO(), minioBucket, "test.txt", minioFileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get version error")
	})
}
func testMinioInitPartsUpload(t *testing.T) {
	fileBiz := newFileMinioBiz()
	c.Convey("test init parts upload success", t, func() {
		rsp, err := fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
			Key: "test-minio.txt",
		})
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		minioUploadId = rsp.UploadId
	})

}
func testMinioUploadParts(t *testing.T) {
	c.Convey("test upload parts success", t, func() {

		bigTmpFile, err := generateRandomFile(1024 * 1024 * 32)
		if err != nil {
			t.Error(err)
		}
		err = uploadParts(bigTmpFile.path, minioUploadId, "test-minio.txt", t)
		c.So(err, c.ShouldBeNil)
		minioBigFileSha256 = bigTmpFile.sha256
	})
}
func testMinioCompleteMultipartUploadFile(t *testing.T) {
	fileBiz := newFileMinioBiz()

	c.Convey("test complete upload fail", t, func() {
		rsp, _ := fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
			Key: "test-minio-upload-fail.txt",
		})
		uploadId := rsp.UploadId
		bigTmpFile, err := generateRandomFile(1024 * 1024 * 32)
		if err != nil {
			t.Error(err)
		}
		_ = uploadParts(bigTmpFile.path, uploadId, "test-minio-upload-fail.txt", t)
		cases := []struct {
			patch     interface{}
			output    []interface{}
			exceptErr error
		}{
			{
				patch:     (*file.FileUsecaseImpl).CompleteMultipartUploadFile,
				output:    []interface{}{nil, fmt.Errorf("complete error")},
				exceptErr: fmt.Errorf("complete error"),
			},
			{
				patch:     os.Open,
				output:    []interface{}{nil, fmt.Errorf("open error")},
				exceptErr: fmt.Errorf("open error"),
			},
			{
				patch:     (*minio.Client).PutObject,
				output:    []interface{}{nil, fmt.Errorf("put error")},
				exceptErr: fmt.Errorf("put error"),
			},
		}
		for _, v := range cases {
			patchx := gomonkey.ApplyFuncReturn(v.patch, v.output...)
			defer patchx.Reset()
			_, err := fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
				Key:         "test-minio.txt",
				UploadId:    uploadId,
				Bucket:      minioBucket,
				ContentType: "text/plain",
				UseVersion:  true,
				Sha256:      minioBigFileSha256,
			}, minioUploadId)
			patchx.Reset()
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, v.exceptErr.Error())

		}
		_, err = fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
			Key:         "test-minio.txt",
			UploadId:    uploadId,
			Bucket:      "",
			ContentType: "text/plain",
			UseVersion:  true,
			Sha256:      minioBigFileSha256,
		}, minioUploadId)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrBucketNotFound.Error())
		t.Logf("test complete upload fail finsih")
	})
	c.Convey("test complete upload success", t, func() {
		rsp, err := fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
			Key:         "test-minio.txt",
			UploadId:    minioUploadId,
			Bucket:      minioBucket,
			ContentType: "text/plain",
			UseVersion:  true,
			Sha256:      minioBigFileSha256,
		}, minioUploadId)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
	})
}
func testMinioDownloadForRange(t *testing.T) {
	fileBiz := newFileMinioBiz()
	fileSize := 1024 * 1024 * 32
	c.Convey("test download file success", t, func() {
		partSize := int64(2 * 1024 * 1024)
		partCount := math.Ceil(float64(fileSize) / float64(partSize))
		shaer := sha256.New()
		for i := 0; i < int(partCount); i++ {
			rangeStartAt := int64(i) * partSize
			rangeEndAt := rangeStartAt + partSize - 1
			if rangeEndAt > int64(fileSize) {
				rangeEndAt = int64(fileSize) - 1
			}
			data, _, err := fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
				Key:    "test-minio.txt",
				Bucket: minioBucket,
			}, rangeStartAt, rangeEndAt, minioFileAuthor)
			c.So(err, c.ShouldBeNil)
			shaer.Write(data)
		}
		b := shaer.Sum(nil)
		c.So(hex.EncodeToString(b), c.ShouldEqual, minioBigFileSha256)

		meta, err := fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{Key: "test-minio.txt", Bucket: minioBucket, Engine: api.FileEngine_FILE_ENGINE_MINIO.String()}, minioFileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(meta.Sha256, c.ShouldEqual, minioBigFileSha256)
	})
	c.Convey("test download file fail", t, func() {
		_, _, err := fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:    "test-minio-2.txt",
			Bucket: minioBucket,
		}, 0, 1024, minioFileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not exist")

		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:    "test-minio.txt",
			Bucket: minioBucket,
		}, 1024, 1, minioUploadId)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrInvalidRange.Error())

		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:    "test-minio.txt",
			Bucket: minioBucket,
		}, 1024, 0, minioFileAuthor)
		c.So(err, c.ShouldBeNil)

		patch := gomonkey.ApplyFuncReturn((*minio.Client).GetObject, nil, fmt.Errorf("get object error"))
		defer patch.Reset()
		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:    "test-minio.txt",
			Bucket: minioBucket,
		}, 1024, 0, minioFileAuthor)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get object error")

		patch2 := gomonkey.ApplyFuncReturn((*minio.Object).ReadAt, 0, fmt.Errorf("read error"))
		defer patch2.Reset()
		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:    "test-minio.txt",
			Bucket: minioBucket,
		}, 1024, 0, minioFileAuthor)
		patch2.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "read error")
	})

}
func testMinioAbortMultipartUpload(t *testing.T) {
	fileBiz := newFileMinioBiz()
	c.Convey("test abort upload success", t, func() {
		rsp, err := fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
			Key: "test-minio-2.txt",
		})
		_, _ = fileBiz.UploadMultipartFileFile(context.TODO(), &api.UploadMultipartFileRequest{
			UploadId:   rsp.UploadId,
			Key:        "test-minio-2.txt",
			PartNumber: 1,
			Content:    []byte("hello"),
		})
		c.So(err, c.ShouldBeNil)
		_, err = fileBiz.AbortMultipartUpload(context.TODO(), &api.AbortMultipartUploadRequest{
			UploadId: rsp.UploadId,
		})
		c.So(err, c.ShouldBeNil)
	})
	c.Convey("test abort upload fail", t, func() {
		rsp, _ := fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
			Key: "test-minio-2.txt",
		})
		patch := gomonkey.ApplyFuncReturn(os.RemoveAll, fmt.Errorf("abort remove all error"))
		defer patch.Reset()
		_, err := fileBiz.AbortMultipartUpload(context.TODO(), &api.AbortMultipartUploadRequest{
			UploadId: rsp.UploadId,
		})
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "abort remove all error")
	})

}
func testMinioDelete(t *testing.T) {
	fileBiz := newFileMinioBiz()
	c.Convey("test delete file success", t, func() {
		_, err := fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: "test-minio.txt", Bucket: minioBucket}, minioFileAuthor)
		c.So(err, c.ShouldBeNil)
	})
	c.Convey("test delete file fail", t, func() {
		patch := gomonkey.ApplyFuncReturn((*minio.Client).RemoveObject, fmt.Errorf("delete error"))
		defer patch.Reset()
		_, err := fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: "test-minio.txt", Bucket: minioBucket}, minioFileAuthor)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "delete error")
	})
}
func testMinioList(t *testing.T) {
	fileBiz := newFileMinioBiz()
	c.Convey("test list success", t, func() {
		rsp, err := fileBiz.List(context.Background(), &api.ListFilesRequest{Bucket: minioBucket, Page: 1, PageSize: 20}, minioFileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		c.So(len(rsp), c.ShouldEqual, 1)
	})
	c.Convey("test list fail", t, func() {
		_, err := fileBiz.List(context.Background(), &api.ListFilesRequest{Bucket: minioBucket, Page: -1, PageSize: -1}, minioFileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "SQL syntax")
	})

}
func TestMinio(t *testing.T) {
	t.Run("testMinioMkBucket", testMinioMkBucket)
	t.Run("testMinioUpload", testMinioUpload)
	t.Run("testMinioDownload", testMinioDownload)
	t.Run("testMinioMeta", testMinioMeta)
	t.Run("testMinioVersion", testMinioVersion)
	t.Run("testMinioInitPartsUpload", testMinioInitPartsUpload)
	t.Run("testMinioUploadParts", testMinioUploadParts)
	t.Run("testMinioCompleteMultipartUploadFile", testMinioCompleteMultipartUploadFile)
	t.Run("testMinioDownloadForRange", testMinioDownloadForRange)
	t.Run("testMinioList", testMinioList)
	t.Run("testMinioAbortMultipartUpload", testMinioAbortMultipartUpload)
	t.Run("testMinioDelete", testMinioDelete)
}
