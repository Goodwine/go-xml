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
	<? whatever ?> qwe 123 .
	</  lol    ><yay attr="123"/>
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
		&ProcInst{},
		&CharData{Data: []byte(" qwe 123 . ")},
		&CloseTag{&Name{local: "lol"}},
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
