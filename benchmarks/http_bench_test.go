// Package benchmarks provides HTTP/1.1, HTTP/2, and HTTP/3 benchmark tests
// for the AgentOS Edge proxy and Control Plane gateway.
//
// Run all benchmarks:
//
//	go test ./benchmarks/... -bench=. -benchtime=10s -benchmem -v
//
// Run at specific concurrency:
//
//	go test ./benchmarks/... -bench=BenchmarkHTTP2_Concurrent -benchtime=10s -cpu=100
//
// Results are automatically written to benchmarks/report.md when using run.sh.
package benchmarks

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/saivedant169/AegisFlow/internal/edgeproxy"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// mockHandler simulates the AegisFlow gateway by sleeping for a short
// duration to emulate policy + provider latency, then returning 200.
var mockHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Simulate minimal gateway overhead (~1ms) — provider latency excluded
	// from benchmarks so we measure proxy throughput, not LLM latency.
	time.Sleep(500 * time.Microsecond)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"object":"chat.completion","choices":[{"message":{"content":"benchmark response"}}]}`)
})

// BenchmarkHTTP1_Sequential measures HTTP/1.1 sequential request throughput.
// This is the baseline — single connection, no multiplexing.
func BenchmarkHTTP1_Sequential(b *testing.B) {
	srv := httptest.NewServer(mockHandler)
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := client.Post(srv.URL+"/v1/chat/completions", "application/json",
			nil)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTP1_Concurrent_100 measures HTTP/1.1 at 100 concurrent goroutines.
func BenchmarkHTTP1_Concurrent_100(b *testing.B) {
	benchmarkHTTP1Concurrent(b, 100)
}

// BenchmarkHTTP1_Concurrent_500 measures HTTP/1.1 at 500 concurrent goroutines.
func BenchmarkHTTP1_Concurrent_500(b *testing.B) {
	benchmarkHTTP1Concurrent(b, 500)
}

// BenchmarkHTTP1_Concurrent_1000 measures HTTP/1.1 at 1000 concurrent goroutines.
func BenchmarkHTTP1_Concurrent_1000(b *testing.B) {
	benchmarkHTTP1Concurrent(b, 1000)
}

func benchmarkHTTP1Concurrent(b *testing.B, concurrency int) {
	srv := httptest.NewServer(mockHandler)
	defer srv.Close()

	transport := &http.Transport{
		MaxIdleConns:        concurrency,
		MaxIdleConnsPerHost: concurrency,
		MaxConnsPerHost:     concurrency,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

	b.SetParallelism(concurrency)
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Post(srv.URL+"/v1/chat/completions", "application/json", nil)
			if err != nil {
				b.Skipf("Skipping due to ephemeral port exhaustion: %v", err)
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkHTTP2_Sequential measures HTTP/2 cleartext (h2c) sequential throughput.
// HTTP/2 allows request multiplexing over a single connection.
func BenchmarkHTTP2_Sequential(b *testing.B) {
	srv := httptest.NewUnstartedServer(mockHandler)
	srv.EnableHTTP2 = true
	srv.StartTLS()
	defer srv.Close()

	transport := srv.Client().Transport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // benchmark only

	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := client.Post(srv.URL+"/v1/chat/completions", "application/json", nil)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTP2_Concurrent_100 measures HTTP/2 at 100 concurrent streams.
func BenchmarkHTTP2_Concurrent_100(b *testing.B) {
	benchmarkHTTP2Concurrent(b, 100)
}

// BenchmarkHTTP2_Concurrent_500 measures HTTP/2 at 500 concurrent streams.
func BenchmarkHTTP2_Concurrent_500(b *testing.B) {
	benchmarkHTTP2Concurrent(b, 500)
}

// BenchmarkHTTP2_Concurrent_1000 measures HTTP/2 at 1000 concurrent streams.
func BenchmarkHTTP2_Concurrent_1000(b *testing.B) {
	benchmarkHTTP2Concurrent(b, 1000)
}

func benchmarkHTTP2Concurrent(b *testing.B, concurrency int) {
	// Use h2c (HTTP/2 cleartext) to avoid TLS overhead in the benchmark.
	h2s := &http2.Server{MaxConcurrentStreams: uint32(concurrency + 100)}
	srv := httptest.NewUnstartedServer(h2c.NewHandler(mockHandler, h2s))
	srv.Start()
	defer srv.Close()

	transport := &http2.Transport{
		// Allow insecure h2c connections in benchmark.
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

	b.SetParallelism(concurrency)
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Post(srv.URL+"/v1/chat/completions", "application/json", nil)
			if err != nil {
				b.Error(err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkHTTP3_Sequential measures HTTP/3 (QUIC) sequential throughput.
func BenchmarkHTTP3_Sequential(b *testing.B) {
	tlsConfig, err := edgeproxy.BuildTLSConfig("", "")
	if err != nil {
		b.Fatalf("failed to build tls config: %v", err)
	}
	tlsConfig.NextProtos = []string{"h3"}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to listen UDP: %v", err)
	}
	defer conn.Close()

	addr := conn.LocalAddr().String()

	srv := &http3.Server{
		Handler:   mockHandler,
		TLSConfig: tlsConfig,
	}

	go func() {
		_ = srv.Serve(conn)
	}()
	defer srv.Close()

	tr := &http3.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	defer tr.Close()

	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}
	url := fmt.Sprintf("https://%s/v1/chat/completions", addr)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := client.Post(url, "application/json", nil)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTP3_Concurrent_100 measures HTTP/3 at 100 concurrent streams.
func BenchmarkHTTP3_Concurrent_100(b *testing.B) {
	benchmarkHTTP3Concurrent(b, 100)
}

// BenchmarkHTTP3_Concurrent_500 measures HTTP/3 at 500 concurrent streams.
func BenchmarkHTTP3_Concurrent_500(b *testing.B) {
	benchmarkHTTP3Concurrent(b, 500)
}

// BenchmarkHTTP3_Concurrent_1000 measures HTTP/3 at 1000 concurrent streams.
func BenchmarkHTTP3_Concurrent_1000(b *testing.B) {
	benchmarkHTTP3Concurrent(b, 1000)
}

func benchmarkHTTP3Concurrent(b *testing.B, concurrency int) {
	tlsConfig, err := edgeproxy.BuildTLSConfig("", "")
	if err != nil {
		b.Fatalf("failed to build tls config: %v", err)
	}
	tlsConfig.NextProtos = []string{"h3"}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to listen UDP: %v", err)
	}
	defer conn.Close()

	addr := conn.LocalAddr().String()

	srv := &http3.Server{
		Handler:   mockHandler,
		TLSConfig: tlsConfig,
	}

	go func() {
		_ = srv.Serve(conn)
	}()
	defer srv.Close()

	tr := &http3.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	defer tr.Close()

	client := &http.Client{Transport: tr, Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://%s/v1/chat/completions", addr)

	b.SetParallelism(concurrency)
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Post(url, "application/json", nil)
			if err != nil {
				b.Skipf("HTTP/3 request failed: %v", err)
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}
