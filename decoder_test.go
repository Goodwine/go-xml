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
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToken(t *testing.T) {
	const input = `
	<a>
	<foo > <!-- asd --> </bar>
	    <foo class="start">asd
	<! whatever [<>][<>]{<>}[<>]{<>} >
	<!><? whatever ?> qwe 123 .
	</  lol:foo    ><yay attr="123"/>
	`
	d := NewDecoder(strings.NewReader(input))

	want := []Token{
		&CharData{Data: []byte(" ")},
		&StartTag{Name: &Name{local: "a"}},
		&CharData{Data: []byte(" ")},
		&StartTag{Name: &Name{local: "foo"}},
		&CharData{Data: []byte(" ")},
		&Comment{},
		&CharData{Data: []byte(" ")},
		&CloseTag{&Name{local: "bar"}},
		&CharData{Data: []byte(" ")},
		&StartTag{Name: &Name{local: "foo"}, Attr: []*Attr{{&Name{local: "class"}, "start"}}},
		&CharData{Data: []byte("asd ")},
		&Directive{},
		&CharData{Data: []byte(" ")},
		&Directive{},
		&ProcInst{},
		&CharData{Data: []byte(" qwe 123 . ")},
		&CloseTag{&Name{local: "foo", space: "lol"}},
		&StartTag{Name: &Name{local: "yay"}, Attr: []*Attr{{&Name{local: "attr"}, "123"}}},
		&CloseTag{&Name{local: "yay"}},
		&CharData{Data: []byte(" ")},
	}

	var got []Token
	for {
		tok, err := d.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatal(err)
		}
		got = append(got, tok.Copy())
	}

	opts := cmp.Options{
		cmp.AllowUnexported(Name{}),
		cmp.Transformer("byteToString", func(in []byte) string { return string(in) }),
	}

	if diff := cmp.Diff(want, got, opts); diff != "" {
		t.Error("Token diff (-want +got)\n", diff)
	}
}

func TestTokenOptionalComment(t *testing.T) {
	const input = `<!--
	--- foo ---
	-->`
	testCases := []struct {
		desc        string
		readComment bool
		want        string
	}{
		{desc: "enabled", readComment: true, want: "\n\t--- foo ---\n\t"},
		{desc: "disabled", readComment: false, want: ""},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			d := NewDecoder(strings.NewReader(input))
			d.ReadComment = tc.readComment
			tok, err := d.Token()
			if err != nil {
				t.Fatal(err)
			}
			if got := string(tok.(*Comment).Data); got != tc.want {
				t.Errorf("comment.Data: '%s', want '%s'", got, tc.want)
			}
		})
	}
}

func TestTokenOptionalDirective(t *testing.T) {
	const input = `<!ENTITY
	[<bar>
	</bar>]
	>`
	testCases := []struct {
		desc          string
		readDirective bool
		want          string
	}{
		{desc: "enabled", readDirective: true, want: "ENTITY\n\t[<bar>\n\t</bar>]\n\t"},
		{desc: "disabled", readDirective: false, want: ""},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			d := NewDecoder(strings.NewReader(input))
			d.ReadDirective = tc.readDirective
			tok, err := d.Token()
			if err != nil {
				t.Fatal(err)
			}
			if got := string(tok.(*Directive).Data); got != tc.want {
				t.Errorf("directive.Data '%s', want '%s'", got, tc.want)
			}
		})
	}
}

func TestTokenErrors(t *testing.T) {
	testCases := []struct {
		desc  string
		input string
		want  string
	}{
		{"start colon", "<:foo>", "unexpected char ':'"},
		{"end colon", "<foo:>", "unexpected char ':'"},
		{"multi colon", "<f:o:o>", "unexpected char ':'"},
		{"bad comment open", "<!- -->", "unexpected char ' ', expected '<--'"},
		{"bad comment close", "<!-- ->", "comment closed too early, must end in '-->'"},
		{"early EOF at tag", "<asd", "unexpected EOF, expected tag identifier at"},
		{"early EOF at comment", "<!-- asd --", "unexpected EOF at"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			d := NewDecoder(strings.NewReader(tc.input))
			got, err := d.Token()
			if err == nil {
				t.Fatalf("expected error, got %T", got)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err: '%s' want '%s'", err, tc.want)
			}
		})
	}
}

func TestErrorLineNumber(t *testing.T) {
	const input = `
	<foo>
		ba>r
	</foo>
	`

	const want = "unexpected char '>' on chardata at row: 3 col: 5"

	d := NewDecoder(strings.NewReader(input))

	// 1. CharData
	// 2. <foo>
	// 3. error!
	for i := 0; i < 2; i++ {
		_, err := d.Token()
		if err != nil {
			t.Fatal(err)
		}
	}
	got, err := d.Token()
	if err == nil {
		t.Fatalf("expected error, got %T", got)
	}
	if err.Error() != want {
		t.Fatalf("err: '%s' want '%s'", err, want)
	}
}
