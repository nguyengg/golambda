package framework

import (
	"bytes"
	"compress/gzip"
	"github.com/nguyengg/golambda/metrics"
	"log"
	"strconv"
	"strings"
)

const CompressMinimumSizeInBytes = 1024

func CompressResponse(c *Context) error {
	return CompressResponseWithMinimumSize(c, CompressMinimumSizeInBytes)
}

func CompressResponseWithMinimumSize(c *Context, minimum int) error {
	if c.StatusCode() != 200 || c.responseHeader.Get("Content-Encoding") != "" || c.response.IsBase64Encoded || len(c.response.Body) < minimum {
		return nil
	}

	if strings.Contains(c.RequestHeader("Accept-Encoding"), "gzip") {
		return compressGzip(c)
	}

	return nil
}

func compressGzip(c *Context) error {
	var buf bytes.Buffer

	w := gzip.NewWriter(&buf)
	_, err := w.Write([]byte(c.response.Body))
	if err == nil {
		err = w.Close()
	}
	if err != nil {
		log.Printf("ERROR compress response body: %v", err)
		_ = c.RespondInternalServerError()
		return err
	}

	m := metrics.Ctx(c.ctx)
	m.AddCount("uncompressedSize", int64(len(c.response.Body)))
	m.AddCount("compressedSize", int64(len(buf.Bytes())))

	c.SetResponseHeader("Content-Length", strconv.FormatInt(int64(len(buf.Bytes())), 10))
	c.SetResponseHeader("Content-Encoding", "gzip")
	return c.RespondOKWithBase64Data(buf.Bytes())
}
