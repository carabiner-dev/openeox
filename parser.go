// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
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
type Parser struct{}

//nolint:unused
type schemaFinder struct {
	Schema string `json:"$schema"`
}

//nolint:unused
var typeDict = map[string]reflect.Type{
	"https://docs.oasis-open.org/openeox/tbd/schema/shell.json": reflect.TypeOf(&v1.Shell{}),
	"https://docs.oasis-open.org/openeox/v1.0/schema/core.json": reflect.TypeOf(&v1.Core{}),
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
// timestamp (see [IsTBA]).
func (p *Parser) ParseCore(data []byte) (*Core, error) {
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
// Sentinel TBA timestamps are written as "tba".
func MarshalShell(s *Shell) ([]byte, error) {
	marshaler := protojson.MarshalOptions{}
	data, err := marshaler.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshaling json: %w", err)
	}
	return postprocessTBA(data), nil
}

// MarshalCore serializes a Core to JSON in the upstream OpenEoX format.
// Sentinel TBA timestamps are written as "tba".
func MarshalCore(c *Core) ([]byte, error) {
	marshaler := protojson.MarshalOptions{}
	data, err := marshaler.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("marshaling json: %w", err)
	}
	return postprocessTBA(data), nil
}
