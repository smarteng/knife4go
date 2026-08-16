// Package oauth 提供 knife4go UI 的 oauth2 登录辅助资源。
package oauth

import "github.com/jasonlabz/knife4go/ui"

// Assets 是 oauth 分类的全部静态资源。
var Assets = []ui.Asset{
	AxiosMinJs,
	Oauth2Html,
}
