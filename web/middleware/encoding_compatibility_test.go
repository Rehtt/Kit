package middleware

import (
	"strconv"
	"strings"
	"testing"
)

// negotiateBeforeOptimization is the pre-optimization implementation, kept as
// an independent compatibility oracle for the scanner (including malformed q).
func negotiateBeforeOptimization(header string) string {
	if header == "" {
		return ""
	}
	qValues := make(map[string]float64, 4)
	wildcard := -1.0
	for _, raw := range strings.Split(header, ",") {
		parts := strings.Split(raw, ";")
		name := strings.ToLower(strings.TrimSpace(parts[0]))
		if name == "" {
			continue
		}
		q := 1.0
		for _, parameter := range parts[1:] {
			key, value, ok := strings.Cut(parameter, "=")
			if !ok || !strings.EqualFold(strings.TrimSpace(key), "q") {
				continue
			}
			parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil || parsed < 0 || parsed > 1 {
				q = 0
			} else {
				q = parsed
			}
		}
		if name == "*" {
			wildcard = q
		} else {
			qValues[name] = q
		}
	}
	quality := func(name string) float64 {
		if q, ok := qValues[name]; ok {
			return q
		}
		if wildcard >= 0 {
			return wildcard
		}
		return 0
	}
	gzipQ, deflateQ := quality("gzip"), quality("deflate")
	if gzipQ <= 0 && deflateQ <= 0 {
		return ""
	}
	if gzipQ >= deflateQ {
		return "gzip"
	}
	return "deflate"
}

func TestNegotiateMatchesBeforeOptimization(t *testing.T) {
	items := []string{
		"", "gzip", "GZip", "deflate", "DEFLATE", "br", "*",
		"gzip;q=0", "gzip;q=0.2", "gzip;q=0.8",
		"deflate;q=0", "deflate;q=0.5", "*;q=0", "*;q=0.5",
		" gzip ; Q = 0.25 ", " deflate ; q = 0.75 ",
	}
	// Three items exercise duplicate weights with a competing coding, where
	// selecting the first rather than the last duplicate changes the result.
	for _, a := range items {
		for _, b := range items {
			for _, c := range items {
				header := a + "," + b + "," + c
				if got, want := negotiate(header), negotiateBeforeOptimization(header); got != want {
					t.Fatalf("negotiate(%q) = %q, before optimization %q", header, got, want)
				}
			}
		}
	}
}

func FuzzNegotiateMatchesBeforeOptimization(f *testing.F) {
	for _, header := range []string{
		"", "gzip", "GZip, DEFLATE", "gzip;q=0.2, gzip;q=0.8, deflate;q=0.5",
		"gzip;q=0, *;q=1", " gzip ; Q = 0.25 , deflate ; q=0.75 ",
		"gzip;q=NaN", "*;q=NaN", "gzip;q=1, *;q=NaN",
		"gzip;q=2", "gzip;q=oops", "gzip;q=0;q=1", ",gzip;;,",
	} {
		f.Add(header)
	}
	f.Fuzz(func(t *testing.T, header string) {
		if got, want := negotiate(header), negotiateBeforeOptimization(header); got != want {
			t.Fatalf("negotiate(%q) = %q, before optimization %q", header, got, want)
		}
	})
}

func BenchmarkNegotiateHeaders(b *testing.B) {
	for _, tc := range []struct{ name, header, want string }{
		{"gzip", "gzip", "gzip"},
		{"weighted", "gzip, deflate;q=0.5", "gzip"},
		{"wildcard", "gzip;q=0, *;q=0.5", "deflate"},
		{"caseAndWhitespace", " GZip ; Q = 0.25 , DEFLATE ; q = 0.75 ", "deflate"},
		{"duplicates", "gzip;q=0.8, gzip;q=0.2, deflate;q=0.5", "deflate"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if got := negotiate(tc.header); got != tc.want {
					b.Fatalf("got %q, want %q", got, tc.want)
				}
			}
		})
	}
}
