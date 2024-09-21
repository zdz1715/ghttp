package ghttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/zdz1715/ghttp/encoding"
)

type DebugInterface interface {
	Before() *httptrace.ClientTrace
	After(request *http.Request, response *http.Response)
}

type TraceInfo struct {
	ctx context.Context

	DNSDuration          time.Duration `json:"DNSDuration,omitempty" yaml:"DNSDuration" xml:"DNSDuration"`
	ConnectDuration      time.Duration `json:"connectDuration,omitempty" yaml:"connectDuration" xml:"connectDuration"`
	TLSHandshakeDuration time.Duration `json:"TLSHandshakeDuration,omitempty" yaml:"TLSHandshakeDuration" xml:"TLSHandshakeDuration"`
	RequestDuration      time.Duration `json:"requestDuration,omitempty" yaml:"requestDuration" xml:"requestDuration"`
	WaitResponseDuration time.Duration `json:"waitResponseDuration,omitempty" yaml:"waitResponseDuration" xml:"waitResponseDuration"`
	ResponseDuration     time.Duration `json:"responseDuration,omitempty" yaml:"responseDuration" xml:"responseDuration"`
	TotalDuration        time.Duration `json:"totalDuration,omitempty" yaml:"totalDuration" xml:"totalDuration"`
}

func (t TraceInfo) Context() context.Context {
	return t.ctx
}

func (t TraceInfo) String() string {
	return string(t.Table())
}

func (t TraceInfo) Table() []byte {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 30, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "--------------------------------------------")
	fmt.Fprintln(w, "Trace\tValue\t")
	fmt.Fprintln(w, "--------------------------------------------")
	_, _ = fmt.Fprintf(w, "DNSDuration\t%s\t\n", t.DNSDuration)
	_, _ = fmt.Fprintf(w, "ConnectDuration\t%s\t\n", t.ConnectDuration)
	_, _ = fmt.Fprintf(w, "TLSHandshakeDuration\t%s\t\n", t.TLSHandshakeDuration)
	_, _ = fmt.Fprintf(w, "RequestDuration\t%s\t\n", t.RequestDuration)
	_, _ = fmt.Fprintf(w, "WaitResponseDuration\t%s\t\n", t.WaitResponseDuration)
	_, _ = fmt.Fprintf(w, "TotalDuration\t%s\t\n", t.TotalDuration)
	fmt.Fprintln(w, "--------------------------------------------")

	_ = w.Flush()
	return buf.Bytes()
}

type traceInfo struct {
	host               string
	dnsDoneInfo        *httptrace.DNSDoneInfo
	getConnHostPort    string
	gotConnInfo        *httptrace.GotConnInfo
	tlsConnectionState *tls.ConnectionState

	dnsStartTime             time.Time
	dnsDoneTime              time.Time
	getConnTime              time.Time
	gotConnTime              time.Time
	tlsHandshakeStartTime    time.Time
	tlsHandshakeDoneTime     time.Time
	gotFirstResponseByteTime time.Time
	wroteRequestTime         time.Time

	startTime        time.Time
	responseDoneTime time.Time
}

func (t traceInfo) write(w io.Writer) {
	// print trace
	if t.dnsDoneInfo != nil {
		write(w, "* Host %s was resolved.", t.getConnHostPort)
		for _, ipAddr := range t.dnsDoneInfo.Addrs {
			if len(ipAddr.IP) == net.IPv4len {
				write(w, "* IPv4: %s", ipAddr.IP)
			}
			if len(ipAddr.IP) == net.IPv6len {
				write(w, "* IPv6: %s", ipAddr.IP)
			}
		}
	}

	if t.gotConnInfo != nil {
		remoteAddr := t.gotConnInfo.Conn.RemoteAddr()
		write(w, "*   Trying %s...", remoteAddr)
		ip, port, _ := net.SplitHostPort(remoteAddr.String())
		write(w, "* Connected to %s (%s) port %s", t.host, ip, port)
	}

	if t.tlsConnectionState != nil {
		write(w, "* SSL connection using %s / %s",
			tls.VersionName(t.tlsConnectionState.Version),
			tls.CipherSuiteName(t.tlsConnectionState.CipherSuite),
		)
		write(w, "* ALPN: server accepted %s", t.tlsConnectionState.NegotiatedProtocol)
		if len(t.tlsConnectionState.VerifiedChains) > 0 && len(t.tlsConnectionState.VerifiedChains[0]) > 0 {
			cer := t.tlsConnectionState.VerifiedChains[0][0]
			write(w, `* Server certificate:
*   subject: CN=%s
*   notBefore: %s
*   notAfter: %s
*   issuer: C=%s; ST=%s; L=%s; O=%s; CN=%s
*   SSL certificate verify ok.`, cer.Subject.CommonName, cer.NotBefore, cer.NotAfter,
				getFirst(cer.Issuer.Country), getFirst(cer.Issuer.Province), getFirst(cer.Issuer.Locality),
				getFirst(cer.Issuer.Organization), cer.Issuer.CommonName)
		}
	}

}

func getFirst(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

type Debug struct {
	Writer        io.Writer
	Trace         bool
	TraceCallback func(w io.Writer, info TraceInfo)

	mux       sync.Mutex
	traceInfo traceInfo
}

var DefaultDebug = func() DebugInterface {
	return &Debug{
		Writer: os.Stderr,
		Trace:  true,
		TraceCallback: func(w io.Writer, info TraceInfo) {
			_, _ = w.Write(info.Table())
		},
	}
}

func (d *Debug) statTraceInfo(ctx context.Context) TraceInfo {
	if !d.Trace {
		return TraceInfo{}
	}
	return TraceInfo{
		ctx:                  ctx,
		DNSDuration:          d.traceInfo.dnsDoneTime.Sub(d.traceInfo.dnsStartTime),
		ConnectDuration:      d.traceInfo.gotConnTime.Sub(d.traceInfo.getConnTime),
		TLSHandshakeDuration: d.traceInfo.tlsHandshakeDoneTime.Sub(d.traceInfo.tlsHandshakeStartTime),
		RequestDuration:      d.traceInfo.wroteRequestTime.Sub(d.traceInfo.gotConnTime),
		WaitResponseDuration: d.traceInfo.gotFirstResponseByteTime.Sub(d.traceInfo.wroteRequestTime),

		ResponseDuration: d.traceInfo.responseDoneTime.Sub(d.traceInfo.gotFirstResponseByteTime),
		TotalDuration:    d.traceInfo.responseDoneTime.Sub(d.traceInfo.startTime),
	}
}

func (d *Debug) init() {
	d.mux.Lock()
	defer d.mux.Unlock()
	if d.Writer == nil {
		d.Writer = os.Stderr
	}
}

func write(w io.Writer, format string, args ...any) {
	if format != "" {
		_, _ = fmt.Fprintf(w, format, args...)
	}
	_, _ = fmt.Fprintf(w, "\n")
}

func (d *Debug) Before() *httptrace.ClientTrace {
	d.init()
	var trace *httptrace.ClientTrace
	if d.Trace {
		d.traceInfo.startTime = time.Now()
		trace = &httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) {
				d.traceInfo.dnsStartTime = time.Now()
				d.traceInfo.host = info.Host
			},
			DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
				d.traceInfo.dnsDoneTime = time.Now()
				d.traceInfo.dnsDoneInfo = &dnsInfo
			},
			GetConn: func(hostPort string) {
				d.traceInfo.getConnTime = time.Now()
				d.traceInfo.getConnHostPort = hostPort
			},
			GotConn: func(connInfo httptrace.GotConnInfo) {
				d.traceInfo.gotConnTime = time.Now()
				d.traceInfo.gotConnInfo = &connInfo
			},
			TLSHandshakeStart: func() {
				d.traceInfo.tlsHandshakeStartTime = time.Now()
			},
			TLSHandshakeDone: func(state tls.ConnectionState, err error) {
				d.traceInfo.tlsHandshakeDoneTime = time.Now()
				d.traceInfo.tlsConnectionState = &state
			},
			GotFirstResponseByte: func() {
				d.traceInfo.gotFirstResponseByteTime = time.Now()
			},
			WroteRequest: func(info httptrace.WroteRequestInfo) {
				d.traceInfo.wroteRequestTime = time.Now()
			},
		}
	}

	return trace
}

func (d *Debug) After(request *http.Request, response *http.Response) {
	// print request and response
	path := request.URL.RequestURI()

	if path == "" {
		path = "/"
	}

	if d.Trace {
		d.traceInfo.responseDoneTime = time.Now()
		if d.TraceCallback != nil {
			d.TraceCallback(d.Writer, d.statTraceInfo(request.Context()))
		}
		if d.traceInfo.host == "" {
			d.traceInfo.host = request.URL.Host
		}
		d.traceInfo.write(d.Writer)
		write(d.Writer, "* using %s", request.Proto)
	}

	write(d.Writer, "> %s %s %s", request.Method, path, request.Proto)

	// write request header
	for k, v := range request.Header {
		write(d.Writer, "> %s: %s", k, strings.Join(v, ","))
	}
	write(d.Writer, ">")

	// request body
	if request.GetBody != nil {
		if reqBodyReader, err := request.GetBody(); err == nil {
			reqBody, _ := io.ReadAll(reqBodyReader)
			codec, _ := CodecForRequest(request)
			reqBodyBs, _ := formatIndent(codec, reqBody)
			if len(reqBodyBs) > 0 {
				write(d.Writer, "")
				write(d.Writer, "%s", string(reqBodyBs))
				write(d.Writer, "")
			}
		}
	}

	write(d.Writer, "> %s %s", response.Proto, response.Status)
	for k, v := range response.Header {
		write(d.Writer, "> %s: %s", k, strings.Join(v, ","))
	}

	// response body
	if response.Body != nil {
		//resBodyReader := io.Reader(response.Body)
		if responseBody, err := io.ReadAll(response.Body); err == nil {
			response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
			codec, _ := CodecForResponse(response)
			resBodyBs, _ := formatIndent(codec, responseBody)
			if len(resBodyBs) > 0 {
				write(d.Writer, "")
				write(d.Writer, "%s", string(resBodyBs))
				write(d.Writer, "")
			}
		}
	}

}

func formatIndent(codec encoding.Codec, data []byte) (result []byte, err error) {
	if len(data) == 0 || codec == nil {
		return result, nil
	}

	var anyData any
	if err = codec.Unmarshal(data, &anyData); err != nil {
		return nil, err
	}

	switch codec.Name() {
	case "json":
		result, err = json.MarshalIndent(anyData, "", "    ")
	default:
		result, err = codec.Marshal(anyData)
	}

	return
}
