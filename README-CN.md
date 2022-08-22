# gzip (这是一个社区驱动的项目)
[English](README.md) | 中文

这是使 Hertz 拥有支持 `gzip` 能力的中间件

## 使用

下载并且安装它:

```sh
go get github.com/hertz-contrib/gzip
```

导入进你的代码:

```go
import "github.com/hertz-contrib/gzip"
```

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


## 许可证

本项目采用Apache许可证。参见 [LICENSE](LICENSE) 文件中的完整许可证文本。
