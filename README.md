# go-xml
Go XML parsing library alternative to `encoding/xml`

> Disclaimer: This is not an official Google product.

This package uses buffers and reusable instances to reduce the amount of time and number of
allocations done during parsing for systems that are resource constrained. The package is
mostly be a drop-in replacement with the exception of a few variable names changing.

## Features

* Optionally normalizes `CharData` whitespace
* Optionally read `Comment`, `ProcInst`, and `Directive` contents

### Not implemented yet

The library already allows manually unmarshaling well-formed XML files, but the following features
are not implemented yet so this library should be used with caution, using on a critical prod
system is **not advised**.

* Option to disable whitespace normalization on `CharData`
* Option to get `ProcInst` contents
* Support attribute values without quotes, like `<foo bar=baz>`
* Better `Comment` end-token (`-->`) validation
* Support `xml:` struct tags
* `Marshal` et al
* `Unmarshal` et al
* `Encode` et al
* `decodeElement`
* Optionally decode html entities like `&quot;` or `&lt;`
* Better error handling - currently assumes proper format with only a few validations
* Catch mismatching start/close tags.

## Comparison

### Notes on encoding/xml

The `encoding/xml` package has a [couple proposals](https://github.com/golang/go/issues/21823)
based on `encoding/json` where buffers are added for names. Additionally there are a few tricks
to reduce resource use that can be applied on the Unmarshalling implementation by avoiding
`Unmarshal()` and [manually decoding the XML](https://stackoverflow.com/a/61858457/950582).
All of the above can lead to 50% reduction on resources just within the standard library's package.

The techniques used in this library are well known to be great performance boosts that could be
used on `encoding/xml`, however Go 1.0 has a backwards compatibility promise that applies to all
standard libraries. These techniques would require breaking this compatibilty promise and therefore
can't be used on `encoding/xml` to their full potential.

### Changes on go-xml

`go-xml` is essentially the same implementation (although not all features are supported), but
instead of returning new token instances, it returns a buffered value. Additional buffers are also
implemented for identifier tag and attribute names, as well as attribute objects themselves. These
buffers are based off the proposals mentioned above.

A key difference is `go-xml` tokens are not supposed to be stored. The token instances are pointers
and the values change every time `decoder.Token()` is called. You **must** `token.Copy()` the value
if you want to store it, but this should rarely be the case as the token should be evaluated as soon
as it's received.

### benchstat

> **Note:** These benchmarks compare Raw tokenization between both libraries. Even though the
> library isn't fully implemented, these improvements will carry over and snowball once
> implementation is complete.

Reading an XML Message Bundle with 75k entries (30MB). Comparing `encoding/xml` with the tricks
_AND_ patching the performance improvement proposals mentioned above vs `go-xml`.

```
name                       time/op
DecodeAll/go-xml-16         398ms ± 1%
DecodeAll/encoding/xml-16   452ms ± 2%

name                       alloc/op
DecodeAll/go-xml-16        18.1MB ± 0%
DecodeAll/encoding/xml-16  76.9MB ± 0%

name                       allocs/op
DecodeAll/go-xml-16          676k ± 0%
DecodeAll/encoding/xml-16   1.99M ± 0%
```

## Code of Conduct

[Same as Go](https://golang.org/conduct)

## Contributing

Read our [contributions](CONTRIBUTING.md) doc.

tl;dr: Get Google's CLA and send a PR!

## Licence

[Apache 2.0](LICENSE)
