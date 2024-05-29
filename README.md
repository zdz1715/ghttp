# ghttp
golang http客户端

## 安装
```shell
go get -u github.com/zdz1715/ghttp@latest
```
## 使用
### 快速开始
```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zdz1715/ghttp"
)

func main() {
	client := ghttp.NewClient(
		ghttp.WithEndpoint("https://gitlab.com"),
	)

	var reply any
	_, err := client.Invoke(context.Background(), http.MethodGet, "/api/v4/projects", nil, &reply, &ghttp.CallOptions{
		Query: map[string]any{
			"page":       "1",
			"membership": true,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Invoke /api/v4/projects success, reply: %+v\n", reply)
}

```
### 选项
#### 配置客户端的HTTP RoundTripper
`WithTransport(trans http.RoundTripper) ClientOption`
```go
// example: 配置代理和客户端证书
ghttp.WithTransport(&http.Transport{
    Proxy: ghttp.ProxyURL(":7890"), // or http.ProxyFromEnvironment
    TLSClientConfig: &tls.Config{
        InsecureSkipVerify: true,
    },
}),
```
#### 配置客户端的请求默认超时时间，若设置了单独超时时间，则优先使用单独超时时间
`WithTimeout(d time.Duration) ClientOption`
```go
// example: 单独设置超时时间
ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
defer cancel()
_, err := client.Invoke(ctx, http.MethodGet, "/api/v4/projects", nil, nil)
```
#### 配置客户端的默认User-Agent
`WithUserAgent(userAgent string) ClientOption`
#### 配置客户端默认访问的endpoint, 若单独请求一个完整URL，则优先使用单独的完整URL
`WithEndpoint(endpoint string) ClientOption`
```go
// example: 单独设置完整URL，会忽略默认endpoint
ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
defer cancel()
_, err := client.Invoke(ctx, http.MethodGet, "https://gitlab.com/api/v4/projects", nil, nil)
```
#### 配置客户端的`Content-type`, 默认：`application/json`
`WithContentType(contentType string) ClientOption`
#### 配置客户端代理，默认：`http.ProxyFromEnvironment` 
`WithProxy(f func(*http.Request) (*url.URL, error))`
#### 配置响应状态码不是`2xx`时，要`bind`的结构体, 也会直接返回错误，方便后续`bind`预期的响应数据
`WithNot2xxError(f func() Not2xxError) ClientOption`

## Bind
### Request Query
支持以下类型：
- `string`
- `map[any]any`
- `[]any`
- `[...]any`
- `struct`

#### struct tag
> query:"yourName,omitempty,int,unix,del:*delimiter*,time_format:2006-01-02"

- `yourName`: 自定义名字(为"-"则忽略)，未设置则使用字段名
- `omitempty`: 值为空则忽略
- `int`: `bool`类型时，`true`为`1`，`false`为`0`
- `unix`: `time.time`类型时，返回时间戳(秒)
- `unixmilli`: `time.time`类型时，返回时间戳(毫秒)
- `unixnano`: `time.time`类型时，返回时间戳(纳秒)
- `del`: `slice`和`array`类型时，*delimiter*有以下值
  - `comma`: 使用","相连
  - `space`: 使用" "相连
  - `semicolon`: 使用";"相连
  - `brackets`: 形式：`user[]=linda&user[]=liming`
  - *custom*: 自定义的值，可以为任何值
- `time_format`: 时间格式化字符串

### Encoding
> 根据`content-type`自动加载对应的`Codec`实例，`content-type`会提取子部分类型，如：`application/json`或`application/vnd.api+json`都为`json`,
#### 自定义`Codec`
覆盖默认的序列化实例
```go
import (
    "github.com/bytedance/sonic"
    "github.com/zdz1715/ghttp/encoding"
)

func main() {
    encoding.RegisterCodec("json", sonic.ConfigDefault)
}
```
## examples
- [gitlab-demo](./examples/gitlab-demo/main.go): 快速创建一个gitlab sdk demo
- [go-gitlab](https://github.com/zdz1715/go-gitlab): gitlab sdk
- [go-jira](https://github.com/zdz1715/go-jira): jira sdk
- [go-gitee](https://github.com/zdz1715/go-gitee): gitee sdk
- [go-dingtalk](https://github.com/zdz1715/go-dingtalk): dingtalk sdk


