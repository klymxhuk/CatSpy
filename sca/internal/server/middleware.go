package server

import (
	"bytes"
	"log"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)                  // сохраняем копию
	return w.ResponseWriter.Write(b) // отдаем дальше клиенту
}

func ResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// здесь уже после выполнения хэндлера
		log.Printf("[RESP] %s %s %d -> %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			blw.body.String(),
		)
	}
}
