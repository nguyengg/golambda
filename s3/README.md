# S3 utilities

```go
package main

import (
	"fmt"

	s4 "github.com/nguyengg/golambda/s3"
)

func main() {
	uri, _ := s4.Parse("s3://my-bucket[1234]/prefix/to/path")
	fmt.Println(uri.Append("my-key.json"))
}
```
