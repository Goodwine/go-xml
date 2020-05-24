package xml

// Token represents an XML Token:
//
// * StartTag: <foo> or <foo />
// * CloseTag: </foo> implicitly </foo> too
// * Comment: <-- foo -->
// * ProcInst: <? foo ?>
// * Directive: <! foo >
// * CharData: Any string outside of angle brackets <>
type Token interface {
	token()

	// Copy the token into a new instance.
	//
	// Tokens instances are constantly modified by the decoding process, this function makes a copy
	// for the unlikely case when the token value must be stored, and for testing!
	Copy() Token
}

// StartElement is an opening XML tag <tag>
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

// EndElement is a closing XML tag </tag>
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
type Comment struct{}

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
type Directive struct{}

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
	// TODO: Add namespace support
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
