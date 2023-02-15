package gzip

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app/client"
)

func TestGzipForClient(t *testing.T) {
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	// TODO
	cli.Use(GzipForClient(DefaultCompression, WithDecompressFnForClient(DecompressFn4Client)))
}
