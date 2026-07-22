// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"

	v1 "github.com/carabiner-dev/openeox/types/v1"
)

// NewParser creates a new OpenEOX parser.
func NewParser() (*Parser, error) {
	return &Parser{}, nil
}

// Parser implements the OpenEOX parser. It handles the "tba" (to be
// announced) values in eox_timestamp_t fields by mapping them to a
// sentinel timestamp. Use [IsTBA] to check parsed timestamps.
type Parser struct {
	// SkipValidation disables JSON schema validation of core documents
	// when parsing.
	SkipValidation bool
}

//nolint:unused
type schemaFinder struct {
	Schema string `json:"$schema"`
}

//nolint:unused
var typeDict = map[string]reflect.Type{
	v1.Schema:           reflect.TypeOf(&v1.Shell{}),
	v1.CoreSchema:       reflect.TypeOf(&v1.Core{}),
	v1.CoreSchemaLegacy: reflect.TypeOf(&v1.Core{}),
}

//nolint:unused
func detectSchema(data []byte) (reflect.Type, error) {
	sf := &schemaFinder{}
	if err := json.Unmarshal(data, sf); err != nil {
		return nil, fmt.Errorf("unable to detect type")
	}
	if t, ok := typeDict[sf.Schema]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("unknown schema type")
}

// ParseShell parses a JSON-encoded OpenEoX shell document. Any "tba"
// values in lifecycle timestamp fields are converted to the sentinel
// timestamp (see [IsTBA]).
func (p *Parser) ParseShell(data []byte) (*Shell, error) {
	processed, err := preprocessTBAShell(data)
	if err != nil {
		return nil, err
	}

	s := &Shell{}
	unmarshaller := protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}
	if err := unmarshaller.Unmarshal(processed, s); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}
	return s, nil
}

// ParseCore parses a JSON-encoded OpenEoX core document. Any "tba"
// values in lifecycle timestamp fields are converted to the sentinel
// timestamp (see [IsTBA]). Unless [Parser.SkipValidation] is set, the
// document is validated against the upstream OpenEoX core JSON schema;
// validation errors wrap [ErrInvalidDocument]. Documents declaring a
// superseded schema URI (such as [CoreSchemaLegacy]) are validated as if
// they declared [CoreSchema].
func (p *Parser) ParseCore(data []byte) (*Core, error) {
	if !p.SkipValidation {
		if err := validateCore(data); err != nil {
			return nil, err
		}
	}

	processed, err := preprocessTBACore(data)
	if err != nil {
		return nil, err
	}

	c := &Core{}
	unmarshaller := protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}
	if err := unmarshaller.Unmarshal(processed, c); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}
	return c, nil
}

// MarshalShell serializes a Shell to JSON in the upstream OpenEoX format.
// Sentinel TBA timestamps are written as "tba" and object keys are sorted.
func MarshalShell(s *Shell) ([]byte, error) {
	marshaler := protojson.MarshalOptions{}
	data, err := marshaler.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshaling json: %w", err)
	}
	return SortKeys(postprocessTBA(data))
}

// MarshalCore serializes a Core to JSON in the upstream OpenEoX format.
// Sentinel TBA timestamps are written as "tba" and object keys are sorted.
func MarshalCore(c *Core) ([]byte, error) {
	marshaler := protojson.MarshalOptions{}
	data, err := marshaler.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("marshaling json: %w", err)
	}
	return SortKeys(postprocessTBA(data))
}

// SortKeys returns the JSON document with all object keys sorted
// lexicographically at every nesting level. The marshal functions apply
// it automatically to their output.
func SortKeys(data []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("decoding document: %w", err)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("encoding document: %w", err)
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}
