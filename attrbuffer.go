package xml

// attrBuffer is a helper buffer for Attr pointers inspired on bytes buffer
type attrBuffer struct {
	buf []*Attr
	pos int
}

func (buf *attrBuffer) growBy(n int) {
	buf.buf = append(buf.buf, make([]*Attr, n)...)
}

func (buf *attrBuffer) reset() {
	buf.pos = 0
}

func (buf *attrBuffer) add(attr *Attr) {
	if buf.pos+1 == len(buf.buf) {
		buf.growBy(len(buf.buf) * 2 / 3)
	}
	buf.buf[buf.pos] = attr
	buf.pos++
}

func (buf *attrBuffer) get() []*Attr {
	if buf.pos == 0 {
		return nil
	}
	attrs := buf.buf[:buf.pos]
	buf.reset()
	return attrs
}
