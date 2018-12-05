// Copyright 2018 The logrange Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package linker

import (
	"testing"
)

func TestParseParams(t *testing.T) {
	var pr parseRes
	err := pr.parseParams(nil)
	if err != nil {
		t.Fatal("Expecting no error, but err=", err)
	}

	err = pr.parseParams([]string{"aaaa"})
	if err != nil || pr.val != "aaaa" {
		t.Fatal("Expecting no error, but err=", err)
	}

	err = pr.parseParams([]string{"aaaa", "bbb"})
	if err == nil {
		t.Fatal("Expecting an error, but err=", err)
	}

	err = pr.parseParams([]string{"", "optional:123 23"})
	if err != nil || pr.defVal != "123 23" {
		t.Fatal("Expecting no error, but err=", err)
	}

	err = pr.parseParams([]string{"", "optional   :\"asd\""})
	if err != nil || pr.defVal != "\"asd\"" || !pr.optional {
		t.Fatal("Expecting no error, but err=", err)
	}

	pr.optional = false
	err = pr.parseParams([]string{"", " optional"})
	if err != nil || !pr.optional {
		t.Fatal("Expecting no error, but err=", err)
	}

	pr.optional = false
	err = pr.parseParams([]string{"optional"})
	if err != nil || pr.optional || pr.val != "optional" {
		t.Fatal("Expecting no error, but err=", err)
	}
}

func TestParseTags(t *testing.T) {
	testParseTags(t, "inject:\"test\", json:\"asdf\"", parseRes{"test", false, ""})
	testParseTags(t, "inject:\"\"", parseRes{"", false, ""})
	testParseTags(t, "json:\"aaa\",inject:\"abc,optional\"", parseRes{"abc", true, ""})
	testParseTags(t, "inject:\"abc,optional :asdf\"", parseRes{"abc", true, "asdf"})
	testParseTags(t, "inject:\"a\\\"bc, optional : asdf\"", parseRes{"a\\\"bc", true, " asdf"})

	testParseTagsError(t, "inject-", nil)
	testParseTagsError(t, "json:\"\"", errTagNotFound)
	testParseTagsError(t, "INJECT:\"\"", errTagNotFound)
	testParseTagsError(t, "inject:\"asdf", nil)
	testParseTagsError(t, "inject:asdf", nil)
	testParseTagsError(t, "inject:\"asdf", nil)
}

func testParseTags(t *testing.T, tags string, pr parseRes) {
	tpr, err := parseTag("inject", tags)
	if err != nil {
		t.Fatal("Unexpected error when parsing tags=", tags, " err=", err)
	}

	if tpr != pr {
		t.Fatal("Expected ", pr, ", but got ", tpr, " for tags=", tags)
	}
}

func testParseTagsError(t *testing.T, tags string, expErr error) {
	_, err := parseTag("inject", tags)
	if err == nil {
		t.Fatal("Expecting an error, but got it nil for tags=", tags, " err=", err)
	}

	if expErr != nil && expErr != err {
		t.Fatal("Expecting ", expErr, ", but got ", err)
	}
}
