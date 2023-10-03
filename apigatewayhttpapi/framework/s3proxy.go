package framework

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nguyengg/golambda/apigatewayhttpapi"
	"log"
)

// ProxyS3 will call S3 with the appropriate GET or HEAD method and sets the response accordingly.
// See apigatewayhttpapi.ProxyS3. Please be mindful of the payload limit; this method cannot be used to return files
// larger than ~6MB.
func (c *Context) ProxyS3(client *s3.Client, bucket, key string) error {
	res, err := apigatewayhttpapi.ProxyS3(c.ctx, client, c.Method(), bucket, key)
	if err != nil {
		log.Printf("ERROR proxy S3: %v", err)
		_ = c.RespondInternalServerError()
		return err
	}

	c.response.StatusCode = res.StatusCode
	c.response.Body = res.Body
	c.response.IsBase64Encoded = res.IsBase64Encoded
	c.response.Cookies = res.Cookies

	for k, v := range res.Headers {
		c.SetResponseHeader(k, v)
	}
	for k, vs := range res.MultiValueHeaders {
		for _, v := range vs {
			c.AddResponseHeader(k, v)
		}
	}

	return nil
}
