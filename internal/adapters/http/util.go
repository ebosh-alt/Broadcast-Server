package http

import (
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"time"
)

func resolveSender(r *http.Request) string {
	if v := r.URL.Query().Get("sender"); v != "" {
		return v
	}
	if v := r.Header.Get("X-Sender"); v != "" {
		return v
	}
	if v := r.Header.Get("Sec-WebSocket-Protocol"); v != "" {
		const key = "sender="
		if len(v) >= len(key) && v[:len(key)] == key {
			return v[len(key):]
		}
	}

	return "anon-" + strconv.FormatInt(time.Now().Unix()%100000, 10)
}

func shortRemote(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	if ip, ok := netip.ParseAddr(host); ok == nil {
		if ip.IsLoopback() {
			return "localhost"
		}
		return ip.String()
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return "localhost"
		}
		return ip.String()
	}
	return host
}
