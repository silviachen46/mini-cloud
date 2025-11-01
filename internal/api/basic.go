package api

import (
	"database/sql"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/silviachen46/mini-cloud/internal/storage"
)

// 这里用 io.Reader / io.Writer，而不是 http.ReadCloser
type Store interface {
	Put(key string, r io.Reader) (int64, error)
	Get(key string, w io.Writer) (int64, error)
	Delete(key string) error
}

// 适配 storage.FS 到上面的 Store 接口
type fsAdaptor struct{ s *storage.FS }

func (a fsAdaptor) Put(key string, r io.Reader) (int64, error) {
	return a.s.Put(key, r)
}
func (a fsAdaptor) Get(key string, w io.Writer) (int64, error) {
	return a.s.Get(key, w)
}
func (a fsAdaptor) Delete(key string) error { return a.s.Delete(key) }

// 注册最小对象接口
func RegisterBasic(r *gin.Engine, db *sql.DB, fs *storage.FS) {
	s := fsAdaptor{s: fs}

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// 上传（整文件 PUT）
	r.PUT("/v1/object/:key", func(c *gin.Context) {
		key := c.Param("key")
		// c.Request.Body 实现了 io.ReadCloser，这里按 io.Reader 用即可
		n, err := s.Put(key, c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		// 简化：ETag 先留空
		_, _ = db.Exec(`INSERT INTO objects(k,size,etag) VALUES(?,?,?)
		                ON CONFLICT(k) DO UPDATE SET size=excluded.size`,
			key, n, "")

		c.JSON(http.StatusOK, gin.H{"key": key, "size": n})
	})

	// 元信息（HEAD）
	r.GET("/v1/object/:key/head", func(c *gin.Context) {
		key := c.Param("key")
		row := db.QueryRow(`SELECT size, etag FROM objects WHERE k=?`, key)
		var size int64
		var etag sql.NullString
		if err := row.Scan(&size, &etag); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		if etag.Valid {
			c.Header("ETag", etag.String)
		}
		// 不再用 fmt 的小技巧，直接用 strconv
		c.Header("Content-Length", strconv.FormatInt(size, 10))
		c.Status(http.StatusOK)
	})

	// 下载
	r.GET("/v1/object/:key", func(c *gin.Context) {
		key := c.Param("key")
		if _, err := s.Get(key, c.Writer); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
	})

	// 删除
	r.DELETE("/v1/object/:key", func(c *gin.Context) {
		key := c.Param("key")
		if err := s.Delete(key); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		_, _ = db.Exec(`DELETE FROM objects WHERE k=?`, key)
		c.Status(http.StatusNoContent)
	})
}

