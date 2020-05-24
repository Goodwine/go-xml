package xml

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	stdxml "encoding/xml"
)

func BenchmarkDecodeAll(b *testing.B) {
	f, err := ioutil.ReadFile("testdata/test.xmb")
	if err != nil {
		b.Fatal(err)
	}

	testCases := []struct {
		desc      string
		decodeAll func()
	}{
		{"go-xml",
			func() {
				decoder := NewDecoder(bytes.NewReader(f))
				for {
					_, err := decoder.Token()
					if err != nil {
						if errors.Is(err, io.EOF) {
							return
						}
						b.Fatal("go-xml parsing error")
					}
				}
			},
		},
		{"encoding/xml",
			func() {
				decoder := stdxml.NewDecoder(bytes.NewReader(f))
				for {
					_, err := decoder.RawToken()
					if err != nil {
						if errors.Is(err, io.EOF) {
							return
						}
						b.Fatal("go-xml parsing error")
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.desc, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tc.decodeAll()
			}
		})
	}
}
