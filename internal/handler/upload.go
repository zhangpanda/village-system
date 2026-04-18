package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10MB
	file, header, err := r.FormFile("file")
	if err != nil {
		errJSON(w, 400, "请选择文件")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".pdf": true, ".doc": true, ".docx": true}
	if !allowed[ext] {
		errJSON(w, 400, "不支持的文件格式")
		return
	}

	name := fmt.Sprintf("%d_%d%s", getUID(r), time.Now().UnixMilli(), ext)
	os.MkdirAll("uploads", 0755)
	dst, err := os.Create(filepath.Join("uploads", name))
	if err != nil {
		errJSON(w, 500, "保存失败")
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	JSON(w, 200, map[string]string{"url": "/uploads/" + name})
}
