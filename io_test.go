// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func testCoreDocument() *Core {
	return &Core{
		Schema:               CoreSchema,
		EndOfLife:            timestamppb.New(time.Date(2028, 12, 31, 23, 59, 59, 0, time.UTC)),
		EndOfSecuritySupport: timestamppb.New(time.Date(2028, 6, 30, 23, 59, 59, 0, time.UTC)),
		EndOfSales:           timestamppb.New(time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)),
		GeneralAvailability:  timestamppb.New(time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)),
		LastUpdated:          timestamppb.New(time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)),
	}
}

func TestMarshalCoreSortsKeys(t *testing.T) {
	data, err := MarshalCore(testCoreDocument())
	require.NoError(t, err)
	//nolint:testifylint // byte-exact comparison: the test asserts key order
	require.Equal(t,
		`{"$schema":"`+CoreSchema+`",`+
			`"end_of_life":"2028-12-31T23:59:59Z",`+
			`"end_of_sales":"2027-12-31T23:59:59Z",`+
			`"end_of_security_support":"2028-06-30T23:59:59Z",`+
			`"general_availability":"2020-01-15T00:00:00Z",`+
			`"last_updated":"2025-07-01T12:00:00Z"}`,
		string(data),
	)
}

func TestSortKeys(t *testing.T) {
	sorted, err := SortKeys([]byte(`{"b": {"d": 1, "c": [{"f": 2, "e": "x&y"}]}, "a": true}`))
	require.NoError(t, err)
	//nolint:testifylint // byte-exact comparison: the test asserts key order
	require.Equal(t, `{"a":true,"b":{"c":[{"e":"x&y","f":2}],"d":1}}`, string(sorted))

	_, err = SortKeys([]byte(`{"unterminated": `))
	require.Error(t, err)
}

func TestCoreFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "core.eox.json")
	require.NoError(t, WriteCoreFile(path, testCoreDocument()))

	parser, err := NewParser()
	require.NoError(t, err)

	core, err := parser.ParseCoreFile(path)
	require.NoError(t, err)
	require.Equal(t, CoreSchema, core.GetSchema())
	require.Equal(t, time.Date(2028, 12, 31, 23, 59, 59, 0, time.UTC), core.GetEndOfLife().AsTime())

	_, err = parser.ParseCoreFile(filepath.Join(t.TempDir(), "missing.json"))
	require.Error(t, err)
}

func TestCoreStreamRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteCore(&buf, testCoreDocument()))

	parser, err := NewParser()
	require.NoError(t, err)

	core, err := parser.ParseCoreReader(&buf)
	require.NoError(t, err)
	require.Equal(t, time.Date(2028, 6, 30, 23, 59, 59, 0, time.UTC), core.GetEndOfSecuritySupport().AsTime())
}

func TestParseCoreURL(t *testing.T) {
	data, err := MarshalCore(testCoreDocument())
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/core.json" {
			w.Write(data) //nolint:errcheck,gosec
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	parser, err := NewParser()
	require.NoError(t, err)

	core, err := parser.ParseCoreURL(t.Context(), srv.URL+"/core.json")
	require.NoError(t, err)
	require.Equal(t, CoreSchema, core.GetSchema())

	_, err = parser.ParseCoreURL(t.Context(), srv.URL+"/nope.json")
	require.Error(t, err)
}

func TestMarshalCoreString(t *testing.T) {
	s, err := MarshalCoreString(testCoreDocument())
	require.NoError(t, err)
	require.Contains(t, s, `"end_of_life":"2028-12-31T23:59:59Z"`)
}

func TestShellIOHelpers(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	shell, err := parser.ParseShellFile("testdata/latest.eox.json")
	require.NoError(t, err)
	require.Len(t, shell.GetStatements(), 2)

	s, err := MarshalShellString(shell)
	require.NoError(t, err)
	require.NotEmpty(t, s)

	path := filepath.Join(t.TempDir(), "shell.eox.json")
	require.NoError(t, WriteShellFile(path, shell))

	var buf bytes.Buffer
	require.NoError(t, WriteShell(&buf, shell))

	reparsed, err := parser.ParseShellReader(&buf)
	require.NoError(t, err)
	require.Len(t, reparsed.GetStatements(), 2)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(s)) //nolint:errcheck,gosec
	}))
	defer srv.Close()

	fetched, err := parser.ParseShellURL(t.Context(), srv.URL)
	require.NoError(t, err)
	require.Len(t, fetched.GetStatements(), 2)
}
