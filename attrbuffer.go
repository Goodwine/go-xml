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
