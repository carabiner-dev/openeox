// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ParseCoreFile reads an OpenEoX core document from the file system.
func (p *Parser) ParseCoreFile(path string) (*Core, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading core document: %w", err)
	}
	return p.ParseCore(data)
}

// ParseCoreReader reads an OpenEoX core document from a data stream.
func (p *Parser) ParseCoreReader(r io.Reader) (*Core, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading core document: %w", err)
	}
	return p.ParseCore(data)
}

// ParseCoreURL fetches an OpenEoX core document from a URL.
func (p *Parser) ParseCoreURL(ctx context.Context, url string) (*Core, error) {
	data, err := fetchDocument(ctx, url)
	if err != nil {
		return nil, err
	}
	return p.ParseCore(data)
}

// ParseShellFile reads an OpenEoX shell document from the file system.
func (p *Parser) ParseShellFile(path string) (*Shell, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading shell document: %w", err)
	}
	return p.ParseShell(data)
}

// ParseShellReader reads an OpenEoX shell document from a data stream.
func (p *Parser) ParseShellReader(r io.Reader) (*Shell, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading shell document: %w", err)
	}
	return p.ParseShell(data)
}

// ParseShellURL fetches an OpenEoX shell document from a URL.
func (p *Parser) ParseShellURL(ctx context.Context, url string) (*Shell, error) {
	data, err := fetchDocument(ctx, url)
	if err != nil {
		return nil, err
	}
	return p.ParseShell(data)
}

// fetchDocument retrieves a document over HTTP(S).
func fetchDocument(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching document: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching document: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return data, nil
}

// WriteCore serializes a Core document and writes it to a data stream.
func WriteCore(w io.Writer, core *Core) error {
	data, err := MarshalCore(core)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("writing core document: %w", err)
	}
	return nil
}

// WriteCoreFile serializes a Core document and writes it to the file system.
func WriteCoreFile(path string, core *Core) error {
	data, err := MarshalCore(core)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing core document: %w", err)
	}
	return nil
}

// MarshalCoreString serializes a Core document to a string.
func MarshalCoreString(core *Core) (string, error) {
	data, err := MarshalCore(core)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteShell serializes a Shell document and writes it to a data stream.
func WriteShell(w io.Writer, shell *Shell) error {
	data, err := MarshalShell(shell)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("writing shell document: %w", err)
	}
	return nil
}

// WriteShellFile serializes a Shell document and writes it to the file system.
func WriteShellFile(path string, shell *Shell) error {
	data, err := MarshalShell(shell)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing shell document: %w", err)
	}
	return nil
}

// MarshalShellString serializes a Shell document to a string.
func MarshalShellString(shell *Shell) (string, error) {
	data, err := MarshalShell(shell)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
