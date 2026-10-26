// webui.go — الواجهة مدمجة داخل الثنائي (embed) — بلا npm، بلا بناء إضافي.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed ui
var files embed.FS

// UI يعيد نظام الملفات لمحتويات مجلد ui.
func UI() fs.FS {
	sub, err := fs.Sub(files, "ui")
	if err != nil {
		panic(err)
	}
	return sub
}
