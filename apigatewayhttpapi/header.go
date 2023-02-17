package apigatewayhttpapi

import (
	"github.com/aws/aws-lambda-go/events"
	"strconv"
	"time"
)

// Modifiers to change the response such as adding headers thereto.
type Opt func(*events.APIGatewayV2HTTPResponse)

// Sets header to the specified value only if both key and value are non-empty strings.
func SetHeader(key, value string) Opt {
	if key == "" || value == "" {
		return func(*events.APIGatewayV2HTTPResponse) {
		}
	}
	return func(res *events.APIGatewayV2HTTPResponse) {
		if res.Headers == nil {
			res.Headers = make(map[string]string)
		}
		res.Headers[key] = value
	}
}

// Sets the cache control header with the specified max age duration.
func SetCacheControlMaxAge(duration time.Duration) Opt {
	return SetHeader("Cache-Control", CacheControlMaxAgeValue(duration))
}

func CacheControlMaxAgeValue(duration time.Duration) string {
	return "max-age=" + strconv.FormatInt(int64(duration/time.Second), 10)
}
