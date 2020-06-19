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

package xml_test

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/Goodwine/go-xml"
)

// This example demonstrates how to decode an XML file using manual tokenization
// into an object, and how to terminate the read-parse loop.
func Example_manualDecodingWithTokens() {
	const data = `
	<msg id="123" desc="flying mammal">
		Bat
	</msg>
	<msg id="456" desc="baseball item">
		Bat
	</msg>
	`

	type Msg struct {
		ID       string
		Desc     string
		Contents string
	}

	var msgs []Msg
	var msg Msg
	d := xml.NewDecoder(strings.NewReader(data))
	for {
		tok, err := d.Token()
		if err != nil {
			// Decoding completes when EOF is returned.
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
			return
		}

		switch tok := tok.(type) {
		case *xml.StartTag:
			if tok.Name.Local() != "msg" {
				log.Fatalf("unexpected start tag: %s", tok.Name.Local())
			}
			for _, attr := range tok.Attr {
				switch attr.Name.Local() {
				case "id":
					msg.ID = attr.Value
				case "desc":
					msg.Desc = attr.Value
				}
			}
		case *xml.CloseTag:
			if tok.Name.Local() != "msg" {
				log.Fatalf("unexpected close tag: %s", tok.Name.Local())
			}
			msgs = append(msgs, msg)
			msg = Msg{}
		case *xml.CharData:
			msg.Contents = string(tok.Data)
		default:
			log.Fatalf("unexpected token: %T", tok)
		}
	}

	for _, m := range msgs {
		fmt.Printf("Msg{ID: '%s', Desc: '%s', Contents: '%s'}\n", m.ID, m.Desc, m.Contents)
	}

	// Output:
	// Msg{ID: '123', Desc: 'flying mammal', Contents: ' Bat '}
	// Msg{ID: '456', Desc: 'baseball item', Contents: ' Bat '}
}
