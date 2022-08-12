package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Handler struct {
	BaseDir string
	DB      DB
}

func (h Handler) UploadHandler() http.HandlerFunc {
	return LoggingHandlerWrapper(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = rw.Write([]byte("请求被拒绝"))
			return
		}
		db := h.DB
		var meta Meta
		meta.Author = req.FormValue("author")
		meta.Description = req.FormValue("description")
		meta.Filename = req.FormValue("filename")
		meta.DisplayName = req.FormValue("display_name")
		meta.Culture = req.FormValue("culture")
		meta.Version = req.FormValue("version")
		meta.ModName = req.FormValue("mod_name")
		in, header, err := req.FormFile("file")
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("表单解析失败:%s", err)))
			return
		}
		if meta.Filename == "" {
			meta.Filename = header.Filename
		}
		err = db.PutMeta(meta, func() error {
			tmpPath := filepath.Join(h.BaseDir, ".uploading_"+meta.Filename)
			{
				out, err := os.Create(tmpPath)
				if err != nil {
					return err
				}
				_, err = io.Copy(out, in)
				_ = out.Close()
				if err != nil {
					return err
				}
			}
			return os.Rename(tmpPath, filepath.Join(h.BaseDir, meta.Filename))
		})
		if err != nil {
			if err == ErrFileExist {
				rw.WriteHeader(http.StatusConflict)
				_, _ = rw.Write([]byte("文件名已存在"))
				return
			}
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("上传失败:%s", err)))
			return
		}
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(fmt.Sprintf("上传成功")))
	})
}

func (h Handler) ListHandler() http.HandlerFunc {
	return LoggingHandlerWrapper(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = rw.Write([]byte("请求被拒绝"))
			return
		}
		db := h.DB
		mod := req.URL.Query().Get("mod")
		if mod == "" {
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte("未指定mod名"))
			return
		}
		if mod == "*" {
			mod = ""
		}
		metas, err := db.ListMetas(mod)
		if err != nil {
			if err == ErrNotFound {
				rw.WriteHeader(http.StatusNotFound)
				_, _ = rw.Write([]byte("此mod暂无汉化包"))
				return
			}
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("查询meta信息失败:%s", err)))
			return
		}
		buff, err := json.Marshal(metas)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("序列化meta信息失败:%s", err)))
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(buff)
	})
}

func (h Handler) MetaHandler() http.HandlerFunc {
	return LoggingHandlerWrapper(func(rw http.ResponseWriter, req *http.Request) {

		if req.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = rw.Write([]byte("请求被拒绝"))
			return
		}
		db := h.DB
		file := req.URL.Query().Get("file")
		if file == "" {
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte("未指定文件名"))
			return
		}
		meta, err := db.GetMeta(file)
		if err != nil {
			if err == ErrNotFound {
				rw.WriteHeader(http.StatusNotFound)
				_, _ = rw.Write([]byte("文件不存在，请刷新列表"))
				return
			}
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("查询meta信息失败:%s", err)))
			return
		}
		f, err := os.Open(filepath.Join(h.BaseDir, file))
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("打开文件失败:%s", err)))
			return
		}
		defer f.Close()
		buff, err := json.Marshal(meta)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("序列化meta信息失败:%s", err)))
			return
		}
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(buff)
	})
}

func (h Handler) DownloadHandler() http.HandlerFunc {
	return LoggingHandlerWrapper(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = rw.Write([]byte("请求被拒绝"))
			return
		}
		db := h.DB
		file := req.URL.Query().Get("file")
		if file == "" {
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte("未指定文件名"))
			return
		}
		meta, err := db.GetMeta(file)
		if err != nil {
			if err == ErrNotFound {
				rw.WriteHeader(http.StatusNotFound)
				_, _ = rw.Write([]byte("文件不存在，请刷新列表"))
				return
			}
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("查询meta信息失败:%s", err)))
			return
		}
		f, err := os.Open(filepath.Join(h.BaseDir, file))
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("打开文件失败:%s", err)))
			return
		}
		defer f.Close()
		rw.Header().Set("author", meta.Author)
		rw.Header().Set("description", meta.Description)
		rw.Header().Set("filename", meta.Filename)
		rw.Header().Set("display_name", meta.DisplayName)
		rw.Header().Set("culture", meta.Culture)
		rw.Header().Set("version", meta.Version)
		rw.Header().Set("mod_name", meta.ModName)
		rw.WriteHeader(http.StatusOK)
		if req.Method == http.MethodHead {
			return
		}
		_, _ = io.Copy(rw, f)
	})
}
