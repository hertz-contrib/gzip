# gzip (This is a community driven project)
English | [中文](README_CN.md)

This is a middleware for hertz to enable `gzip` support.

## Usage

Download and install it:

```sh
go get github.com/hertz-contrib/gzip
```

Import it in your code:

```go
import "github.com/hertz-contrib/gzip"
```

### For server

Canonical example:

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/gzip"
	
)

func main() {
	h := server.Default(server.WithHostPorts(":8080"))
	h.Use(gzip.Gzip(gzip.DefaultCompression))
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "pong "+fmt.Sprint(time.Now().Unix()))
	})
	h.Spin()
}
```


Customized Excluded Extensions

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/gzip"

)

func main() {
	h := server.Default(server.WithHostPorts(":8080"))
	h.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".pdf", ".mp4"})))
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "pong "+fmt.Sprint(time.Now().Unix()))
	})
	h.Spin()
}
```

Customized Excluded Paths

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/gzip"

)

func main() {
	h := server.Default(server.WithHostPorts(":8080"))
	h.Use(gzip.Gzip(gzip.DefaultCompression,gzip.WithExcludedPaths([]string{"/api/"})))
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "pong "+fmt.Sprint(time.Now().Unix()))
	})
	h.Spin()
}
```

Customized Excluded Paths

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/gzip"

)

func main() {
	h := server.Default(server.WithHostPorts(":8080"))
	h.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPathRegexes([]string{".*"})))
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "pong "+fmt.Sprint(time.Now().Unix()))
	})
	h.Spin()
}
```

### For server-Stream compression

The server first compresses the data before streaming it out

> Note: Using this middleware will hijack the response writer and may have an impact on other interfaces.
Therefore, it is only necessary to use this middleware on interfaces with streaming gzip requirements.

Canonical example:
```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/gzip"
)

func main() {
	h := server.Default(server.WithHostPorts(":8081"))
	// Note: Using this middleware will hijack the response writer and may have an impact on other interfaces.
	// Therefore, it is only necessary to use this middleware on interfaces with streaming gzip requirements.
	h.GET("/ping", gzip.GzipStream(gzip.DefaultCompression), func(ctx context.Context, c *app.RequestContext) {
		for i := 0; i < 10; i++ {
			c.Write([]byte(fmt.Sprintf("chunk %d: %s\n", i, strings.Repeat("hi~", i)))) // nolint: errcheck
			c.Flush()                                                                   // nolint: errcheck
			time.Sleep(time.Second)
		}
	})
	h.Spin()
}
```

### For client

Canonical example:

```go
package main

import (
   "context"

   "github.com/cloudwego/hertz/pkg/app/client"
   "github.com/cloudwego/hertz/pkg/protocol"
   "github.com/hertz-contrib/gzip"
)

func main() {
   client, _ := client.NewClient()
   client.Use(gzip.GzipForClient(gzip.DefaultCompression))
   _, _, _ = client.Post(context.Background(),
      []byte{},
      "http://localhost:8080/ping",
      &protocol.Args{})
}
```

Customized Excluded Extensions

```go
package main

import (
   "context"

   "github.com/cloudwego/hertz/pkg/app/client"
   "github.com/cloudwego/hertz/pkg/protocol"
   "github.com/hertz-contrib/gzip"
)

func main() {
   client, _ := client.NewClient()
   client.Use(gzip.GzipForClient(gzip.DefaultCompression,gzip.WithExcludedPathsForClient([]string{"/api/"})))
   _, _, _ = client.Post(context.Background(),
      []byte{},
      "http://localhost:8080/ping",
      &protocol.Args{})
}
```

Customized Excluded Paths

```go
package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/gzip"
)

func main() {
	client, _ := client.NewClient()
	client.Use(gzip.GzipForClient(gzip.DefaultCompression, gzip.WithExcludedExtensionsForClient([]string{".pdf", ".mp4"})))
	statusCode, body, err := client.Post(context.Background(),
		[]byte{},
		"http://localhost:8080/ping",
		&protocol.Args{})
	fmt.Printf("%d, %s, %s", statusCode, body, err)
}

```

Customized Excluded Paths

```go
package main

import (
   "context"
   "fmt"

   "github.com/cloudwego/hertz/pkg/app/client"
   "github.com/cloudwego/hertz/pkg/protocol"
   "github.com/hertz-contrib/gzip"
)

func main() {
   client, _ := client.NewClient()
   client.Use(gzip.GzipForClient(gzip.DefaultCompression, gzip.WithExcludedPathRegexesForClient([]string{".*"})))
   statusCode, body, err := client.Post(context.Background(),
      []byte{},
      "http://localhost:8080/ping",
      &protocol.Args{})
   fmt.Printf("%d, %s, %s", statusCode, body, err)
}
```

## License

This project is under Apache License. See the [LICENSE](LICENSE) file for the full license text.
