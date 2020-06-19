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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/google/triemap"
)

type decodeError string

// Error implements error interface, returns itself since it's already a string.
func (err decodeError) Error() string {
	return string(err)
}

const (
	// UnexpectedChar is thrown when an unexpected rune or characters appears outside of an attribute
	// value or CharData token.
	UnexpectedChar decodeError = "unexpected char"
)

// Decoder processes an XML input and generates tokens or processes into a given struct.
type Decoder struct {
	// ReadComment enables reading and returning back the comment contents. Otherwise returns an empty
	// node. Disabled by default.
	ReadComment bool

	// ReadComment enables reading and returning back the directive contents. Otherwise returns an
	// empty node. Disabled by default.
	//
	// Note that we DO NOT process directives, we simply return back the string within `<! ... >`
	ReadDirective bool

	r   io.RuneReader
	row int
	col int

	// startedTag indicates whether the current last token consumed an open angle bracket (<)
	startedTag bool

	// selfClosingTag indicates that the last StartTag token self closed, and a CloseTag token should
	// be emitted instead of consuming more characters.
	selfClosingTag *Name

	// Buffers for input read so far for the _current token_. This buffer is cleared on every new
	// token, identifier like tag names or attributes, and string values.
	buf   *bytes.Buffer
	attrs *attrBuffer
	names triemap.RuneSliceMap

	// The following are object buffers to save on allocations by reusing the same instance every
	// time the Decoder.Token function is called.
	// Because returning plain structs would copy by value, it would cause a large amount of
	// allocations for medium to large files, and this allows returning the same pointer multiple
	// times.
	startTagBuf  StartTag
	closeTagBuf  CloseTag
	charDataBuf  CharData
	commentBuf   Comment
	procInstBuf  ProcInst
	directiveBuf Directive
}

// NewDecoder instantiates a Decoder to process a Reader input.
func NewDecoder(r io.Reader) *Decoder {
	var attrBuf attrBuffer
	attrBuf.growBy(30)
	var buf bytes.Buffer
	buf.Grow(1000)
	return &Decoder{
		r:     bufio.NewReader(r),
		buf:   &buf,
		attrs: &attrBuf,
	}
}

// Token will decode the next token from the current XML position.
//
// The token is meant to be processed BEFORE the next token is called.
// Contents of previous tokens can be modified at any time during tokenization.
func (d *Decoder) Token() (Token, error) {
	// TODO: Add option to Decoder so Token pushes/pops tag names onto a stack to verify tags match 1:1.
	t, err := d.token()
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w at row: %d col: %d", err, d.row+1, d.col)
	}
	return t, err
}

func (d *Decoder) token() (Token, error) {
	if d.startedTag {
		d.startedTag = false
		return d.angleStart()
	}
	if d.selfClosingTag != nil {
		d.closeTagBuf.Name = d.selfClosingTag
		d.selfClosingTag = nil
		return &d.closeTagBuf, nil
	}
	r, err := d.next()
	if err != nil {
		return nil, err
	}
	switch {
	case r == '<':
		// StartElement
		// EndElement
		// Comment
		// ProcInst
		// Directive
		return d.angleStart()
	case r == '>':
		return nil, unexpectedChar(r)
	}
	//CharData
	return d.charData(r)
}

// unexpectedChar is a utility function to attach the rune to the UnexpectedChar error.
func unexpectedChar(r rune) error {
	return fmt.Errorf("%w %q", UnexpectedChar, r)
}

// next reads the next rune and updates col/row positions for better error messaging.
func (d *Decoder) next() (rune, error) {
	r, _, err := d.r.ReadRune()
	if r == '\n' {
		d.col = 0
		d.row++
	} else {
		d.col++
	}
	return r, err
}

// checkUnexpectedEOF is a helper function to catch an EOF and transform it to UnexpectedEOF
// when it happens mid-way during parsing.
func checkUnexpectedEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}
	return err
}

func (d *Decoder) charData(start rune) (Token, error) {
	d.buf.Reset()
	// Normalize whitespace
	// TODO: Add an option on Decoder to not-normalize whitespace
	space := unicode.IsSpace(start)
	if space {
		start = ' '
	}
	d.buf.WriteRune(start)
	for {
		r, err := d.next()
		if err != nil {
			d.charDataBuf.Data = d.buf.Bytes()
			return &d.charDataBuf, nil
		}
		if r == '<' {
			d.startedTag = true
			d.charDataBuf.Data = d.buf.Bytes()
			return &d.charDataBuf, nil
		}
		if r == '>' {
			return nil, fmt.Errorf("%w on chardata", unexpectedChar(r))
		}
		// Normalize whitespace
		// TODO: Add an option on Decoder to not-normalize whitespace
		if unicode.IsSpace(r) {
			if space {
				continue
			}
			space = true
			r = ' '
		} else {
			space = false
		}
		d.buf.WriteRune(r)
	}
}

// angleStart will return the token corresponding to the previous `<` character
//
// At this point it could be StartTag, Comment, EndTag, Directive, or ProcInst
func (d *Decoder) angleStart() (Token, error) {
	r, err := d.next()
	if err != nil {
		return nil, checkUnexpectedEOF(err)
	}
	switch {
	case isASCIILetter(r):
		// StartElement
		d.buf.Reset()
		d.buf.WriteRune(r)
		return d.startTag()
	case r == '/':
		// EndElement
		return d.closeTag()
	case r == '!':
		// Comment
		// Directive
		d.buf.Reset()

		r, err := d.next()
		if err != nil {
			return nil, checkUnexpectedEOF(err)
		}
		if r != '-' {
			return d.directive(r)
		}

		r, err = d.next()
		if err != nil {
			return nil, checkUnexpectedEOF(err)
		}
		if r != '-' {
			return nil, fmt.Errorf("%w, expected '<--'", unexpectedChar(r))
		}

		return d.comment()
	case r == '?':
		// ProcInst
		return d.procInst()
	}
	return nil, unexpectedChar(r)
}

// startTag processes a token like: <foo> or <foo bar="baz" biz='x' boz>
func (d *Decoder) startTag() (Token, error) {
	name, last, err := d.readIdentifier(false)
	if err != nil {
		return nil, fmt.Errorf("%w, expected tag identifier", err)
	}

	d.startTagBuf.Name = name
	if last == '>' {
		return &d.startTagBuf, nil
	}

	// attributes
	d.attrs.reset()
	for {
		last, err = d.consumeSpace()
		if err != nil {
			return nil, fmt.Errorf("%w, expected attribute identifier", err)
		}

		if last == '/' {
			d.selfClosingTag = d.startTagBuf.Name
			last, err = d.next()
			if err != nil {
				return nil, fmt.Errorf("%w, expected '>' for self-close tag", err)
			}
			if last != '>' {
				return nil, fmt.Errorf("%w, expected '>' for self-close tag", unexpectedChar(last))
			}
		}

		// See if there are no more attributes
		switch {
		case last == '>':
			d.startTagBuf.Attr = d.attrs.get()
			return &d.startTagBuf, nil
		case !isASCIILetter(last):
			return nil, fmt.Errorf("%w on tag <%s>", unexpectedChar(last), d.startTagBuf.Name)
		}

		// Find the attribute name
		d.buf.Reset()
		d.buf.WriteRune(last)
		name, last, err := d.readIdentifier(true)
		if err != nil {
			return nil, fmt.Errorf("%w for attribute on tag <%s>", err, d.startTagBuf.Name)
		}
		if unicode.IsSpace(last) {
			last, err = d.consumeSpace()
			if err != nil {
				return nil, fmt.Errorf("%w for attribute %s on tag <%s>", err, name, d.startTagBuf.Name)
			}
		}

		// attribute without value looks like <foo name> or <foo name bar="baz">
		attr := Attr{Name: name}
		if last == '=' || last == '>' || isASCIILetter(last) {
			d.attrs.add(&attr)
		} else {
			return nil, fmt.Errorf("%w for attribute %s on tag <%s>", unexpectedChar(last), name, d.startTagBuf.Name)
		}
		if last == '>' {
			d.startTagBuf.Attr = d.attrs.get()
			return &d.startTagBuf, nil
		}

		if last != '=' {
			continue
		}

		// Find attribute value, they are surrounded by quotes
		last, err = d.consumeSpace()
		if err != nil {
			return nil, fmt.Errorf("%w after attribute %s on tag <%s>", err, name, d.startTagBuf.Name)
		}
		// TODO: support naked attribute values, i.e. without quotes
		if last != '"' && last != '\'' {
			return nil, fmt.Errorf("%w, expected value for attribute %s on tag <%s>", unexpectedChar(last), name, d.startTagBuf.Name)
		}
		d.buf.Reset()
		attr.Value, err = d.readString(last)
		if err != nil {
			return nil, fmt.Errorf("%w reading attribute %s value on tag <%s>", err, name, d.startTagBuf.Name)
		}
	}
}

// readString reads a string ending in a given quote rune, assumes initial quote has
// already been consumed.
//
// It doesn't support escaping with backslash or HTML entities like &quot;
func (d *Decoder) readString(quote rune) (string, error) {
	for {
		r, err := d.next()
		if err != nil {
			return "", checkUnexpectedEOF(err)
		}
		if r == quote {
			return d.buf.String(), nil
		}
		d.buf.WriteRune(r)
	}
}

// closeTag processes a token like: </foo>
func (d *Decoder) closeTag() (Token, error) {
	last, err := d.consumeSpace()
	if err != nil {
		return nil, fmt.Errorf("%w, expected closing tag", err)
	}
	if !isASCIILetter(last) {
		return nil, fmt.Errorf("%w, expected closing tag", unexpectedChar(last))
	}
	d.buf.Reset()
	d.buf.WriteRune(last)
	name, last, err := d.readIdentifier(false)
	if err != nil {
		return nil, fmt.Errorf("%w, expected closing tag", err)
	}
	if unicode.IsSpace(last) {
		last, err = d.consumeSpace()
		if err != nil {
			return nil, fmt.Errorf("%w on closing tag </%v>", err, name)
		}
	}
	if last != '>' {
		return nil, fmt.Errorf("%w, expected '>' for closing tag </%s>", unexpectedChar(last), name)
	}
	d.closeTagBuf.Name = name
	return &d.closeTagBuf, nil
}

// comment processes a token like: <-- -->
func (d *Decoder) comment() (Token, error) {
	var count int
	for {
		r, err := d.next()
		if err != nil {
			return nil, checkUnexpectedEOF(err)
		}
		if r == '-' {
			count++
		}
		if r == '>' {
			// TODO: Only allow == 2
			if count >= 2 {
				if d.ReadComment {
					d.commentBuf.Data = d.buf.Bytes()
					d.commentBuf.Data = d.commentBuf.Data[:len(d.commentBuf.Data)-2]
				}
				return &d.commentBuf, nil
			}
			return nil, errors.New("comment closed too early, must end in '-->'")
		}
		if d.ReadComment {
			d.buf.WriteRune(r)
		}
	}
}

// procInst processes a token like: <?  ?>
func (d *Decoder) procInst() (Token, error) {
	// TODO: Read contents, but disabled by default, only enabled with an option.
	// TODO: Only allow at the beginning of the file
	var questionMark bool
	for {
		r, err := d.next()
		if err != nil {
			return nil, checkUnexpectedEOF(err)
		}
		if r == '>' {
			if questionMark {
				return &d.procInstBuf, nil
			}
			return nil, errors.New("proc inst closed too early, must end in '?>'")
		}
		questionMark = r == '?'
	}
}

// directive processes a token like: <!  > or <! [] > or <! {} >
func (d *Decoder) directive(last rune) (Token, error) {
	if last == '>' {
		if d.ReadDirective {
			d.directiveBuf.Data = d.buf.Bytes()
		}
		return &d.directiveBuf, nil
	} else if d.ReadDirective {
		d.buf.WriteRune(last)
	}
	for {
		r, err := d.next()
		if err != nil {
			return nil, checkUnexpectedEOF(err)
		}
		// looping because []{}[]{}
		for r == '[' || r == '{' {
			if d.ReadDirective {
				d.buf.WriteRune(r)
			}

			target := ']'
			if r == '{' {
				target = '}'
			}
			isCloseBracket := func(r rune) bool { return r != target }
			r, err = d.consume(isCloseBracket, d.ReadDirective)
			if err != nil {
				return nil, fmt.Errorf("%w, expected %q", err, target)
			}
		}
		if r == '>' {
			if d.ReadDirective {
				d.directiveBuf.Data = d.buf.Bytes()
			}
			return &d.directiveBuf, nil
		}
		if d.ReadDirective {
			d.buf.WriteRune(r)
		}
	}
}

// consume reads out all runes matching the function and return the last non-space rune
func (d *Decoder) consume(match func(rune) bool, read bool) (rune, error) {
	for {
		r, err := d.next()
		if err != nil {
			return 0, checkUnexpectedEOF(err)
		}
		if !match(r) {
			return r, nil
		}
		if read {
			d.buf.WriteRune(r)
		}
	}
}

// consumeSpace reads out all spaces and return the last non-space rune
func (d *Decoder) consumeSpace() (rune, error) {
	return d.consume(unicode.IsSpace, false)
}

// readIdentifier reads the next Name for attribute or tag names
//
// the distinction between attribute and tag name is important because attributes can be
// follwed up by an equals sign (=) character.
func (d *Decoder) readIdentifier(isAttribute bool) (*Name, rune, error) {
	var prev, r rune
	var err error
	var foundNS bool
loop:
	for {
		r, err = d.next()
		if err != nil {
			return nil, 0, checkUnexpectedEOF(err)
		}
		switch {
		case r == ':' && !foundNS:
			foundNS = true
			fallthrough
		case isIdentifierChar(r):
			d.buf.WriteRune(r)
		case unicode.IsSpace(r), (r == '=' && isAttribute):
			last := prev
			if !isASCIILetter(last) {
				return nil, 0, fmt.Errorf("%w reading identifier", unexpectedChar(last))
			}
			break loop
		case r == '>':
			break loop
		default:
			return nil, 0, fmt.Errorf("%w reading identifier", unexpectedChar(r))
		}
		prev = r
	}

	// Somehow implementing a []rune buffer is worse performing than casting buf.String()
	runes := []rune(d.buf.String())
	name, ok := d.names.Get(runes)
	if ok {
		return name.(*Name), r, nil
	}

	if foundNS {
		parts := strings.SplitN(d.buf.String(), ":", 2)
		if len(parts[1]) == 0 {
			// We only validate the second part because the first part can't be empty for the code
			// to enter this function.
			return nil, 0, fmt.Errorf("%w reading identifier", unexpectedChar(':'))
		}
		name = &Name{space: parts[0], local: parts[1]}
	} else {
		name = &Name{local: d.buf.String()}
	}
	d.names.Put(runes, name)
	return name.(*Name), r, nil
}

func isIdentifierChar(r rune) bool {
	return isASCIILetter(r) || r == '-' || r == '_'
}

func isASCIILetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
