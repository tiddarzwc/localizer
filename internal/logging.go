package internal

import (
	"encoding/hex"
	"log"
	"math/rand"
	"net/http"
)

type loggingResponseWriter struct {
	w    http.ResponseWriter
	stat int
	size int
}

func (l *loggingResponseWriter) Header() http.Header {
	return l.w.Header()
}

func (l *loggingResponseWriter) Write(bytes []byte) (int, error) {
	written, err := l.w.Write(bytes)
	l.size += written
	return written, err
}

func (l *loggingResponseWriter) WriteHeader(statusCode int) {
	l.stat = statusCode
	l.w.WriteHeader(statusCode)
}

func LoggingHandlerWrapper(h http.HandlerFunc) http.HandlerFunc {
	var logidBuff = make([]byte, 16)
	_, _ = rand.Read(logidBuff)
	logid := hex.EncodeToString(logidBuff)
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("[log_id: %s] [from: %s] [url: %s] [content-length: %d]\n", logid, request.RemoteAddr, request.RequestURI, request.ContentLength)
		lw := &loggingResponseWriter{w: writer}
		h(lw, request)
		log.Printf("[log_id: %s] [stat_code: %d] [data_sent: %d]\n", logid, lw.stat, lw.size)
	}
}
