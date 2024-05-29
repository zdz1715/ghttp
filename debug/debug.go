package debug

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

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
	Writer io.Writer
	Trace  bool

	traceInfo traceInfo
}

func (d *Debug) TraceInfo() *TraceInfo {
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

func (d *Debug) Before(request *http.Request) (*http.Request, error) {
	if d.Trace {
		d.traceInfo.startTime = time.Now()
		trace := &httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) {
				d.traceInfo.dnsStartTime = time.Now()
			},
			DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
				d.traceInfo.dnsDoneTime = time.Now()
			},
			GetConn: func(hostPort string) {
				d.traceInfo.getConnTime = time.Now()
			},
			GotConn: func(connInfo httptrace.GotConnInfo) {
				d.traceInfo.gotConnTime = time.Now()
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

		request = request.WithContext(
			httptrace.WithClientTrace(request.Context(), trace),
		)
	}

	// print request
	if d.Writer == nil {
		return request, nil
	}

	path := request.URL.RequestURI()

	if path == "" {
		path = "/"
	}

	fmt.Fprintf(d.Writer, "> %s %s %s\n", request.Method, path, request.Proto)
	return request, nil
}

func (d *Debug) After(response *http.Response) error {
	if d.Trace {
		d.traceInfo.responseDoneTime = time.Now()
	}
	// print response
	if d.Writer == nil {
		return nil
	}
	return nil
}
