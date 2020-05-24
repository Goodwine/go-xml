// Package xml is an alternative to the standard library `encoding/xml` package.
//
// This package focuses on reducing allocations as much as possible through the use of buffers and
// reusable objects. It should be mostly a drop-in replacement too with a few identifiers being
// renamed. For example StartElement -> StartTag.
//
// The library is still incomplete, see the repository's README. But should be ready to be used in
// prod assuming you're currently unmarshalling by manually extracting tokens out of the decoder.
//
// Compared to the standard library `encoding/xml` package, after manually extracting raw tokens as
// well as applying the pending patches from the XML issue about slow parsing, go-xml is even faster
// and uses much less memory. This difference gets magnified with larger files.
//
//    10-34% faster
//    76% less allocated memory
//    66% less memory allocations
package xml
