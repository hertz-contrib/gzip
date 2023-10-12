# gzip (这是一个社区驱动的项目)
[English](README.md) | 中文

这是使 Hertz 拥有支持 `gzip` 能力的中间件

## 使用

下载并且安装它：

```sh
go get github.com/hertz-contrib/gzip
```

导入进你的代码：

```go
import "github.com/hertz-contrib/gzip"
```
### 服务端

建议示例:
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

自定义排除的扩展
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

定制的排除路径

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

定制的排除路径
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

### 服务端-流式压缩

服务端先将数据压缩再流式写出去

> 注意：使用该中间件会劫持 response writer，可能会对其他接口造成影响，因此，只需要在有流式 gzip 需求的接口使用该中间件。

建议示例:
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
			time.Sleep(200 * time.Millisecond)
		}
	})
	h.Spin()
}
```

### 客户端

建议示例：

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

自定义排除的扩展

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

定制的排除路径

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

定制的排除路径

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


## 许可证

本项目采用 Apache 许可证。参见 [LICENSE](LICENSE) 文件中的完整许可证文本。
