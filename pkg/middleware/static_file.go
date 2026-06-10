package middleware

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

const INDEX = "index.html"

type ServeFileSystem interface {
	http.FileSystem
	Exists(prefix string, path string) bool
}

type localFileSystem struct {
	http.FileSystem
	root    string
	indexes bool
}

func LocalFile(root string, indexes bool) *localFileSystem {
	return &localFileSystem{
		FileSystem: gin.Dir(root, indexes),
		root:       root,
		indexes:    indexes,
	}
}

func (l *localFileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		name := path.Join(l.root, p)
		stats, err := os.Stat(name)
		if err != nil {
			return false
		}
		if stats.IsDir() {
			if !l.indexes {
				index := path.Join(name, INDEX)
				_, err := os.Stat(index)
				if err != nil {
					return false
				}
			}
		}
		return true
	}
	return false
}

func ServeRoot(urlPrefix, root string) gin.HandlerFunc {
	return Serve(urlPrefix, LocalFile(root, false))
}

// Static returns a middleware handler that serves static files in the given directory.
func Serve(urlPrefix string, fs ServeFileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if urlPrefix != "/" && c.Request.URL.Path == urlPrefix {
			c.Redirect(http.StatusMovedPermanently, urlPrefix+"/")
			c.Abort()
			return
		}
		if fs.Exists(urlPrefix, c.Request.URL.Path) {
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, reqPath string) bool {
	p := strings.TrimPrefix(reqPath, prefix)
	if len(p) == len(reqPath) && prefix != "/" {
		return false
	}
	// 去掉可能的前导 "/"，embed 文件系统期望相对路径
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		// 使用当前目录表示根
		p = "."
	}
	_, err := e.Open(p)
	return err == nil
}

func EmbedFolder(fsEmbed embed.FS, targetPath string) ServeFileSystem {
	fsys, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}
