// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	f, err := ioutil.ReadFile("testdata/bench.xmb")
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
		{"encoding_xml",
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
