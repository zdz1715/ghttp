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
> 可自定义，需实现`Not2xxError`方法

`WithNot2xxError(f func() Not2xxError) ClientOption`
#### 配置Debug选项
`WithDebug(f func() DebugInterface) ClientOption`
> 可自定义，需实现`DebugInterface`方法

```go
// example: 
ghttp.WithDebug(func() ghttp.DebugInterface {
  return &ghttp.Debug{
    Writer: os.Stdout,
    Trace:  true, // 开启trace
    TraceCallback: func(w io.Writer, info ghttp.TraceInfo) { // trace完成时回调
        _, _ = w.Write(info.Table())
    },
  }
}),
```
### 调用
`Invoke(ctx context.Context, method, path string, args any, reply any, opts ...CallOption) (*http.Response, error)`

`Do(req *http.Request, opts ...CallOption) (*http.Response, error)`

`Calloption`是一个接口，会按添加顺序循环调用，只需实现以下方法即可定制
```go
type CallOption interface {
    Before(request *http.Request) error
    After(response *http.Response) error
}
```
#### `ghttp.CallOptions`
实现了`Calloption`接口，主要实现以下功能
```go
type CallOptions struct {
	// request
	Query any // set query

	// Auth
	Username string // Basic Auth
	Password string
	
	BearerToken string // Bearer Token

	// hooks
	BeforeHook func(request *http.Request) error
	AfterHook  func(response *http.Response) error
}
```

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
## Debug
设置`WithDebug`开启
```go
ghttp.WithDebug(func() ghttp.DebugInterface {
  return &ghttp.Debug{
    Writer: os.Stdout,
    Trace:  true, // 开启trace
    TraceCallback: func(w io.Writer, info ghttp.TraceInfo) { // trace完成时回调
        _, _ = w.Write(info.Table())
    },
  }
}),
```
```shell
--------------------------------------------
Trace                         Value                         
--------------------------------------------
DNSDuration                   0s                            
ConnectDuration               164.1172ms                    
TLSHandshakeDuration          161.5662ms                    
RequestDuration               0s                            
WaitResponseDuration          297.2657ms                    
TotalDuration                 461.3829ms                    
--------------------------------------------
*   Trying 127.0.0.1:63210...
* Connected to gitlab.com (127.0.0.1) port 63210
* SSL connection using TLS 1.3 / TLS_AES_128_GCM_SHA256
* ALPN: server accepted h2
* Server certificate:
*   subject: CN=gitlab.com
*   notBefore: 2024-04-12 00:00:00 +0000 UTC
*   notAfter: 2025-05-11 23:59:59 +0000 UTC
*   issuer: C=GB; ST=Greater Manchester; L=Salford; O=Sectigo Limited; CN=Sectigo RSA Domain Validation Secure Server CA
*   SSL certificate verify ok.
* using HTTP/1.1
> POST /oauth/token?membership=true&page=1 HTTP/1.1
> Accept: application/json
> Content-Type: application/json
>

{
    "grant_type": "password"
}

< HTTP/2.0 400 Bad Request
> X-Download-Options: noopen
> X-Gitlab-Meta: {"correlation_id":"01HZ48D05K3E8F71G09EXDMG48","version":"1"}
> Set-Cookie: _cfuvid=UGRIFEQC.lqlEzqwR6ymzq0zQe6HyFchNQlvwDLltE8-1717056209213-0.0.1.1-604800000; path=/; domain=.gitlab.com; HttpOnly; Secure; SameSite=None
> Content-Type: application/json; charset=utf-8
> Referrer-Policy: strict-origin-when-cross-origin
> Www-Authenticate: Bearer realm="Doorkeeper", error="invalid_grant", error_description="The provided authorization grant is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client."
> Report-To: {"endpoints":[{"url":"https:\/\/a.nel.cloudflare.com\/report\/v4?s=14HQQcEuSoirAmB89OjP1C%2BR0hyiLtcXUgepnx2QfcGFodpQGwARF%2BR%2Bznmxj8th0R1TOVpvTSYjji4DrMLzApbQCqKw4ZY0g72zWI7mA3cq4%2Fcfq8kB2Jign9A%3D"}],"group":"cf-nel","max_age":604800}
> Server: cloudflare
> X-Permitted-Cross-Domain-Policies: none
> Gitlab-Sv: web-gke-us-east1-c
> Cf-Cache-Status: DYNAMIC
> Nel: {"success_fraction":0.01,"report_to":"cf-nel","max_age":604800}
> Cf-Ray: 88bd353a0c848b24-HKG
> Cache-Control: no-store
> Content-Security-Policy: base-uri 'self'; child-src https://www.google.com/recaptcha/ https://www.recaptcha.net/ https://www.googletagmanager.com/ns.html https://*.zuora.com/apps/PublicHostedPageLite.do https://gitlab.com/admin/ https://gitlab.com/assets/ https://gitlab.com/-/speedscope/index.html https://gitlab.com/-/sandbox/ 'self' https://gitlab.com/assets/ blob: data:; connect-src 'self' https://gitlab.com wss://gitlab.com https://sentry.gitlab.net https://new-sentry.gitlab.net https://customers.gitlab.com https://snowplow.trx.gitlab.net https://sourcegraph.com https://collector.prd-278964.gl-product-analytics.com; default-src 'self'; font-src 'self'; form-action 'self' https: http:; frame-ancestors 'self'; frame-src https://www.google.com/recaptcha/ https://www.recaptcha.net/ https://www.googletagmanager.com/ns.html https://*.zuora.com/apps/PublicHostedPageLite.do https://gitlab.com/admin/ https://gitlab.com/assets/ https://gitlab.com/-/speedscope/index.html https://gitlab.com/-/sandbox/; img-src 'self' data: blob: http: https:; manifest-src 'self'; media-src 'self' data: blob: http: https:; object-src 'none'; report-uri https://new-sentry.gitlab.net/api/4/security/?sentry_key=f5573e26de8f4293b285e556c35dfd6e&sentry_environment=gprd; script-src 'strict-dynamic' 'self' 'unsafe-inline' 'unsafe-eval' https://www.google.com/recaptcha/ https://www.gstatic.com/recaptcha/ https://www.recaptcha.net/ https://apis.google.com https://*.zuora.com/apps/PublicHostedPageLite.do 'nonce-tvf6QQ1+r4ahhZwSduumvQ=='; style-src 'self' 'unsafe-inline'; worker-src 'self' https://gitlab.com/assets/ blob: data:
> X-Frame-Options: SAMEORIGIN
> X-Request-Id: 01HZ48D05K3E8F71G09EXDMG48
> X-Runtime: 0.015358
> X-Xss-Protection: 0
> Gitlab-Lb: haproxy-main-31-lb-gprd
> Strict-Transport-Security: max-age=31536000
> Date: Thu, 30 May 2024 08:03:29 GMT
> Vary: Origin
> X-Content-Type-Options: nosniff

{
    "error": "invalid_grant",
    "error_description": "The provided authorization grant is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client."
}
```
## examples
- [gitlab-demo](./examples/gitlab-demo/main.go): 快速创建一个gitlab sdk demo
- [go-gitlab](https://github.com/zdz1715/go-gitlab): gitlab sdk
- [go-jira](https://github.com/zdz1715/go-jira): jira sdk
- [go-gitee](https://github.com/zdz1715/go-gitee): gitee sdk
- [go-dingtalk](https://github.com/zdz1715/go-dingtalk): dingtalk sdk


