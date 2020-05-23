# go-xml
Go XML parsing library alternative to `encoding/xml`

> Disclaimer: This is not an official Google product.

The standard library `encoding/xml` package is considered by some to be slow. Unfortunately due to
the Go1.0 compatibility promise the package can't be updated to be faster because it requires some
breaking changes. IMO the package shouldn't even be part of the standard library.

This package solves the issue by reducing the amount of time and space via buffers and pointers
instead of structs whose values get copied over and over.

## Features

* Optionally normalizes `CharData` whitespace
* Optionally read `Comment`, `ProcInst`, and `Directive` contents

### Not implemented yet
I am able to manually unmarshal an XML with the current implementation with enough confidence to
use it in prod with the caveat the the files I process are 100% machine generated and correct by
definition, where even panicking is OK.

* Option to disable whitespace normalization on `CharData`
* Option to get `Comment`, `ProcInst`, and `Directive` contents
* Support attribute values without quotes, like `<foo bar=baz>`
* Better `Comment` end-token (`-->`) validation
* Support `xml:` struct tags
* `Marshal` et al
* `Unmarshal` et al
* `Encode` et al
* `decodeElement`
* Optionally decode html entities like `&quot;` or `&lt;`
* Better error handling - currently assumes proper format with only a few validations

## Comparison

### Notes on encoding/xml
The `encoding/xml` package has a [couple proposals](https://github.com/golang/go/issues/21823)
based on `encoding/json` where buffers are added for names. Additionally there are a few tricks
to reduce resource use that can be applied on the Unmarshalling implementation by avoiding
`Unmarshal()` and [manually decoding the XML](https://stackoverflow.com/a/61858457/950582).
All of the above can lead to 50% reduction on resources just within the standard library's package.

### Changes on go-xml
`go-xml` is essentially the same implementation (although not all features are supported), but
instead of returning new token instances, it returns a buffered value. Additional buffers are also
implemented for identifier tag and attribute names, as well as attribute objects themselves. These
buffers are based off the proposals mentioned above.

A key difference is `go-xml` tokens are not supposed to be stored. The token instances are pointers
and the values change every time `decoder.Token()` is called. You must `token.Copy()` the value if
you want to store it, but this should rarely be the case as the token should be evaluated as soon
as it's received.

### benchstat

Reading an XML Message Bundle with 75k entries (30MB). Comparing `encoding/xml` with the tricks
AND patching the performance improvement proposals vs `go-xml`.

Note: Unmarshalling was done manually, according to pprof the allocations were roughly 1:1 with
the XML file size, the rest of the resources are the manual unmarshalling which was done the
same way for both.

```
name          old time/op    new time/op    delta
Unmarshal-16     662ms ± 4%     470ms ± 4%  -28.98%  (p=0.008 n=5+5)

name          old alloc/op   new alloc/op   delta
Unmarshal-16     280MB ± 0%      83MB ± 0%  -70.29%  (p=0.016 n=4+5)

name          old allocs/op  new allocs/op  delta
Unmarshal-16     5.03M ± 0%     1.58M ± 0%  -68.50%  (p=0.008 n=5+5)
```

## Code of Conduct

[Same as Go](https://golang.org/conduct)

## Contributing

Read our [contributions](CONTRIBUTING.md) doc.

tl;dr: Get Google's CLA and send a PR!

## Licence

[Apache 2.0](LICENSE)
