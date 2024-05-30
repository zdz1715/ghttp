package debug

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"os"
	"sync"
	"time"
)

type Interface interface {
	Before() *httptrace.ClientTrace
	After(request *http.Request, response *http.Response)
}

type TraceInfo struct {
	DNSDuration          time.Duration `json:"DNSDuration,omitempty" yaml:"DNSDuration" xml:"DNSDuration"`
	ConnectDuration      time.Duration `json:"connectDuration,omitempty" yaml:"connectDuration" xml:"connectDuration"`
	TLSHandshakeDuration time.Duration `json:"TLSHandshakeDuration,omitempty" yaml:"TLSHandshakeDuration" xml:"TLSHandshakeDuration"`
	RequestDuration      time.Duration `json:"requestDuration,omitempty" yaml:"requestDuration" xml:"requestDuration"`
	WaitResponseDuration time.Duration `json:"waitResponseDuration,omitempty" yaml:"waitResponseDuration" xml:"waitResponseDuration"`
	ResponseDuration     time.Duration `json:"responseDuration,omitempty" yaml:"responseDuration" xml:"responseDuration"`
	TotalDuration        time.Duration `json:"totalDuration,omitempty" yaml:"totalDuration" xml:"totalDuration"`
}

type traceInfo struct {
	host string

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

type Debug struct {
	Writer        io.Writer
	Trace         bool
	TraceCallback func(*TraceInfo)

	mux       sync.Mutex
	traceInfo traceInfo
}

func NewDefaultDebug() *Debug {
	return &Debug{
		Writer: os.Stdout,
		Trace:  true,
	}
}

func (d *Debug) statTraceInfo() *TraceInfo {
	if !d.Trace {
		return nil
	}
	return &TraceInfo{
		DNSDuration:          d.traceInfo.dnsDoneTime.Sub(d.traceInfo.dnsStartTime),
		ConnectDuration:      d.traceInfo.gotConnTime.Sub(d.traceInfo.getConnTime),
		TLSHandshakeDuration: d.traceInfo.tlsHandshakeDoneTime.Sub(d.traceInfo.tlsHandshakeStartTime),
		RequestDuration:      d.traceInfo.wroteRequestTime.Sub(d.traceInfo.gotConnTime),
		WaitResponseDuration: d.traceInfo.gotFirstResponseByteTime.Sub(d.traceInfo.wroteRequestTime),

		ResponseDuration: d.traceInfo.responseDoneTime.Sub(d.traceInfo.gotFirstResponseByteTime),
		TotalDuration:    d.traceInfo.responseDoneTime.Sub(d.traceInfo.startTime),
	}
}

func (d *Debug) write(format string, args ...any) {
	if d.Writer == nil {
		d.mux.Lock()
		d.Writer = os.Stderr
		d.mux.Unlock()
	}
	_, _ = fmt.Fprintf(d.Writer, format, args...)
	_, _ = fmt.Fprintln(d.Writer)
}

func (d *Debug) Before() *httptrace.ClientTrace {
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
				d.write("*Host %s was resolved.", d.traceInfo.host)
			},
			GetConn: func(hostPort string) {
				d.traceInfo.getConnTime = time.Now()
				d.write("*   Trying %s...", hostPort)
			},
			GotConn: func(connInfo httptrace.GotConnInfo) {
				d.traceInfo.gotConnTime = time.Now()
				d.write("* Connected to %s", connInfo.Conn.RemoteAddr())
			},
			TLSHandshakeStart: func() {
				d.traceInfo.tlsHandshakeStartTime = time.Now()
			},
			TLSHandshakeDone: func(tls.ConnectionState, error) {
				d.traceInfo.tlsHandshakeDoneTime = time.Now()
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
	if d.Trace {
		d.traceInfo.responseDoneTime = time.Now()
		if d.TraceCallback != nil {
			d.TraceCallback(d.statTraceInfo())
		}
	}

	// print request and response
	path := request.URL.RequestURI()

	if path == "" {
		path = "/"
	}

	d.write("> %s %s %s", request.Method, path, request.Proto)
}
