// Package ui 提供 knife4go 内置的前端资源：Knife4j UI 页面、WebJars 静态文件与站点图标。
//
// 所有静态资源通过 //go:embed 内嵌于 dist/ 子目录，运行时由 AllAssets 遍历生成
// 一组 Asset。相较于把每个资源转成 Go 源码字符串常量的做法，embed 方式编译更快、
// 无需运行时解码、维护成本更低。
package ui

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

// Asset 描述一条静态 GET 路由：固定路径、Content-Type 与响应内容。
// ContentType 为空串表示不设置 Content-Type 响应头。
type Asset struct {
	Path        string
	ContentType string
	Content     []byte
}

// DocHtmlRelativePath 是 Knife4j UI 页面的默认路径。
const DocHtmlRelativePath = "/doc.html"

// distFS 内嵌 dist 目录下的全部静态资源。
//
//go:embed all:dist
var distFS embed.FS

// contentTypeByExt 按扩展名映射静态资产的 Content-Type。
// 未列出的扩展名保持为空串（不设置响应头，由客户端/上游按二进制处理）。
var contentTypeByExt = map[string]string{
	".html": "text/html;charset=UTF-8",
	".css":  "text/css;charset=UTF-8",
	".js":   "application/javascript;charset=UTF-8",
}

// contentTypeFor 根据资源相对路径推断 Content-Type。
// 特例：*.js.LICENSE.txt 归为 .txt，不设 Content-Type，与历史行为保持一致。
func contentTypeFor(rel string) string {
	if strings.HasSuffix(rel, ".LICENSE.txt") {
		return ""
	}
	return contentTypeByExt[strings.ToLower(path.Ext(rel))]
}

// AllAssets 返回内嵌 dist/ 目录下的全部静态资产，
// 但不包含 doc.html —— 后者由 DocHtml 单独提供，注册路径可经 DocPath 定制。
// 结果按目录遍历顺序返回；每个 Asset 的 Path 以 "/" 开头，Content 为原始字节。
func AllAssets() []Asset {
	var assets []Asset
	_ = fs.WalkDir(distFS, "dist", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if p == "dist/doc.html" {
			return nil
		}
		b, readErr := distFS.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		rel := "/" + strings.TrimPrefix(p, "dist/")
		assets = append(assets, Asset{
			Path:        rel,
			ContentType: contentTypeFor(rel),
			Content:     b,
		})
		return nil
	})
	return assets
}

// DocHtml 是 Knife4j UI 页面资产。
var DocHtml = func() Asset {
	b, err := distFS.ReadFile("dist/doc.html")
	if err != nil {
		panic("knife4go: missing embedded doc.html: " + err.Error())
	}
	return Asset{
		Path:        DocHtmlRelativePath,
		ContentType: contentTypeByExt[".html"],
		Content:     b,
	}
}()
