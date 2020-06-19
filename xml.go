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

// Token represents an XML Token:
//
//    StartTag:  <foo> or <foo />
//    CloseTag:  </foo> implicitly </foo> too
//    Comment:   <-- foo -->
//    ProcInst:  <? foo ?>
//    Directive: <! foo >
//    CharData:  Any string outside of angle brackets <>
type Token interface {
	token()

	// Copy the token into a new instance.
	//
	// Tokens instances are constantly modified by the decoding process, this function makes a copy
	// for the unlikely case when the token value must be stored, and for testing!
	Copy() Token
}

// StartTag is an opening XML tag <tag>
type StartTag struct {
	Name *Name
	Attr []*Attr
}

func (*StartTag) token() {}

func (s *StartTag) Copy() Token {
	c := StartTag{Name: s.Name}
	if s.Attr != nil {
		c.Attr = make([]*Attr, len(s.Attr))
		copy(c.Attr, s.Attr)
	}
	return &c
}

// CloseTag is a closing XML tag </tag>
type CloseTag struct {
	Name *Name
}

func (*CloseTag) token() {}

func (t *CloseTag) Copy() Token {
	return &CloseTag{t.Name}
}

// CharData contains a text node
type CharData struct {
	Data []byte
}

func (*CharData) token() {}

func (t *CharData) Copy() Token {
	data := make([]byte, len(t.Data))
	copy(data, t.Data)
	return &CharData{data}
}

// Comment has the format <-- -->
//
// It can have two or more `-` at the beginning, but it must have two `-` at the end.
type Comment struct {
	// Data contains the contents of the comment. It is empty by default.
	//
	// Enable `d.ReadComment` to include the contents in the token.
	Data []byte
}

func (*Comment) token() {}

func (t *Comment) Copy() Token {
	c := *t
	return &c
}

// ProcInst has the format <? ... ?>
type ProcInst struct{}

func (*ProcInst) token() {}

func (t *ProcInst) Copy() Token {
	c := *t
	return &c
}

// Directive has the format <! ... >
//
// Note: We do NOT process the directive token. We only read it.
type Directive struct {
	// Data contains the contents of the directive. It is empty by default.
	//
	// Enable `d.ReadDirective` to include the contents in the token.
	Data []byte
}

func (*Directive) token() {}

func (t *Directive) Copy() Token {
	c := *t
	return &c
}

// Attr is a tag attribute like <foo bar="baz">.
// This will store an Attr with name "bar" and value "baz"
type Attr struct {
	Name  *Name
	Value string
}

// Name stores an identifier name from either a tag or an attribute like <foo bar="baz">
// This will generate the names "foo" for the tag, and "bar" for the attribute.
type Name struct {
	local string
	space string
}

// Local returns the identifier name without XML namespace.
//
// For example <a:b> generates the local name "b" with namespace "a"
// This method will return "b".
func (n *Name) Local() string {
	if n == nil {
		return ""
	}
	return string(n.local)
}

// Space returns the identifier name without XML namespace.
//
// For example <a:b> generates the local name "b" with namespace "a"
// This method will return "a".
func (n *Name) Space() string {
	if n == nil {
		return ""
	}
	return string(n.local)
}
