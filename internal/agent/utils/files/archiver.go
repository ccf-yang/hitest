package files

import (
	"context"
)

type ShellArchiver interface {
	Extract(ctx context.Context, filePath, dstDir string, secret string) error
	Compress(ctx context.Context, sourcePaths []string, dstFile string, secret string) error
}
