// Package ui 提供 knife4go 内置的前端资源：Knife4j UI 页面、WebJars 静态文件与站点图标。
package ui

import (
	"encoding/hex"
	"fmt"
)

// Asset 描述一条静态 GET 路由：固定路径、Content-Type 与响应内容。
// ContentType 为空串表示不设置 Content-Type 响应头。
type Asset struct {
	Path        string
	ContentType string
	Content     []byte
}

// MustHex 将 16 进制字符串解码为字节；仅在包初始化声明资产时使用，非法输入直接 panic。
func MustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(fmt.Sprintf("knife4go: invalid hex asset content: %v", err))
	}
	return b
}
