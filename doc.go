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

// Package xml is an alternative to the standard library `encoding/xml` package.
//
// This package uses of buffers and reusable object instances during unmarshalling to reduce
// allocations and struct initialization and the copy-by-value behavior of Go. This saves
// considerable amounts of resources for constrained systems.
//
// The library is still incomplete, see the repository's README. But should be ready to be used in
// prod assuming you're currently unmarshalling by manually extracting tokens out of the decoder.
//
//    10-34% faster
//    76% less allocated memory
//    66% less memory allocations
package xml
