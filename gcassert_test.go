// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package gcassert

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestParseDirectives(t *testing.T) {
	fileSet := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedCompiledGoFiles |
			packages.NeedTypes | packages.NeedTypesInfo,
		Fset: fileSet,
	}, "./testdata")
	if err != nil {
		t.Fatal(err)
	}
	absMap, err := parseDirectives(pkgs, fileSet)
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Convert the map into relative paths for ease of testing, and remove
	// the syntax node so we don't have to test that as well.
	relMap := make(directiveMap, len(absMap))
	for absPath, m := range absMap {
		for k, info := range m {
			info.n = nil
			m[k] = info
		}
		relPath, err := filepath.Rel(cwd, absPath)
		if err != nil {
			t.Fatal(err)
		}
		relMap[relPath] = m
	}

	expectedMap := directiveMap{
		"testdata/bce.go": {
			8:  {directives: []assertDirective{bce}},
			11: {directives: []assertDirective{bce, inline}},
			13: {directives: []assertDirective{bce, inline}},
			17: {directives: []assertDirective{bce, inline}},
			19: {directives: []assertDirective{bce, inline}},
		},
		"testdata/inline.go": {
			45: {directives: []assertDirective{inline}},
			49: {directives: []assertDirective{inline}},
			51: {directives: []assertDirective{inline}},
			55: {directives: []assertDirective{inline}},
			57: {directives: []assertDirective{inline}},
			58: {directives: []assertDirective{inline}},
		},
	}
	assert.Equal(t, expectedMap, relMap)
}

func TestGCAssert(t *testing.T) {
	var w strings.Builder
	err := GCAssert(&w, "./testdata", "./testdata/otherpkg")
	if err != nil {
		t.Fatal(err)
	}

	expectedOutput := `testdata/bce.go:8:	fmt.Println(ints[5]): Found IsInBounds
testdata/bce.go:17:	sum += notInlinable(ints[i]): call was not inlined
testdata/bce.go:19:	sum += notInlinable(ints[i]): call was not inlined
testdata/inline.go:45:	alwaysInlined(3): call was not inlined
testdata/inline.go:51:	sum += notInlinable(i): call was not inlined
testdata/inline.go:55:	sum += 1: call was not inlined
testdata/inline.go:58:	test(0).neverInlinedMethod(10): call was not inlined
testdata/inline.go:60:	otherpkg.A{}.NeverInlined(sum): call was not inlined
`
	assert.Equal(t, expectedOutput, w.String())
}
