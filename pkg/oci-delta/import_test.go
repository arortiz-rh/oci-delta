package ocidelta

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/containers/storage"
	"github.com/opencontainers/go-digest"
)

type mockStore struct {
	storage.Store

	layersByDigest                map[digest.Digest][]storage.Layer
	diffData                      map[string][]byte
	putLayerCallback              func(parentID string, data io.Reader) (*storage.Layer, error)
	layersByUncompressedDigestErr error
	diffErr                       error
}

func (m *mockStore) LayersByUncompressedDigest(d digest.Digest) ([]storage.Layer, error) {
	if m.layersByUncompressedDigestErr != nil {
		return nil, m.layersByUncompressedDigestErr
	}

	layers, ok := m.layersByDigest[d]
	if !ok {
		return nil, nil
	}

	return layers, nil
}

func (m *mockStore) Diff(from, to string, options *storage.DiffOptions) (io.ReadCloser, error) {
	if m.diffErr != nil {
		return nil, m.diffErr
	}

	data, ok := m.diffData[to]
	if !ok {
		return nil, fmt.Errorf("layer %s not found", to)
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *mockStore) PutLayer(id, parent string, names []string, mountLabel string, writeable bool, options *storage.LayerOptions, diff io.Reader) (*storage.Layer, int64, error) {
	if m.putLayerCallback != nil {
		layer, err := m.putLayerCallback(parent, diff)
		if err != nil {
			return nil, -1, err
		}

		return layer, 11, nil
	}

	return nil, -1, fmt.Errorf("putLayerCallback not set")
}

func TestReuseStorageLayer(t *testing.T) {
	tmpDir := t.TempDir()

	layerContent := []byte("mock-layer-content")
	layerDigest := digest.FromBytes(layerContent)

	baseLayerID := "base-layer-id"
	existingLayerID := "existing-layer-id"

	tests := []struct {
		name           string
		diffID         digest.Digest
		parentID       string
		store          *mockStore
		expectError    bool
		errorContains  string
		verifyParent   bool
		expectedParent string
	}{
		{
			name:     "reuse existing layer with same parent",
			diffID:   layerDigest,
			parentID: baseLayerID,
			store: &mockStore{
				layersByDigest: map[digest.Digest][]storage.Layer{
					layerDigest: {{
						ID:                 existingLayerID,
						Parent:             baseLayerID,
						UncompressedDigest: layerDigest,
					}},
				},
			},
			expectError:    false,
			verifyParent:   true,
			expectedParent: baseLayerID,
		},
		{
			name:     "recreate layer with different parent",
			diffID:   layerDigest,
			parentID: "",
			store: &mockStore{
				layersByDigest: map[digest.Digest][]storage.Layer{
					layerDigest: {{
						ID:                 existingLayerID,
						Parent:             baseLayerID,
						UncompressedDigest: layerDigest,
					}},
				},
				diffData: map[string][]byte{
					existingLayerID: layerContent,
				},
				putLayerCallback: func(parentID string, data io.Reader) (*storage.Layer, error) {
					return &storage.Layer{
						ID:                 "new-layer-id",
						Parent:             parentID,
						UncompressedDigest: layerDigest,
					}, nil
				},
			},
			expectError:    false,
			verifyParent:   true,
			expectedParent: "",
		},
		{
			name:     "layer not found",
			diffID:   digest.FromString("nonexistent"),
			parentID: "",
			store: &mockStore{
				layersByDigest: map[digest.Digest][]storage.Layer{},
			},
			expectError:   true,
			errorContains: "not found in storage",
		},
		{
			name:     "error extracting layer diff",
			diffID:   layerDigest,
			parentID: "",
			store: &mockStore{
				layersByDigest: map[digest.Digest][]storage.Layer{
					layerDigest: {{
						ID:                 existingLayerID,
						Parent:             baseLayerID,
						UncompressedDigest: layerDigest,
					}},
				},
				diffErr: fmt.Errorf("simulated diff error"),
			},
			expectError:   true,
			errorContains: "failed to extract layer diff",
		},
		{
			name:     "multiple layers exist but none match parent",
			diffID:   layerDigest,
			parentID: "desired-parent-id",
			store: &mockStore{
				layersByDigest: map[digest.Digest][]storage.Layer{
					layerDigest: {
						{
							ID:                 "layer1",
							Parent:             "parent1",
							UncompressedDigest: layerDigest,
						},
						{
							ID:                 "layer2",
							Parent:             "parent2",
							UncompressedDigest: layerDigest,
						},
					},
				},
				diffData: map[string][]byte{
					"layer1": layerContent,
				},
				putLayerCallback: func(parentID string, data io.Reader) (*storage.Layer, error) {
					return &storage.Layer{
						ID:                 "new-layer",
						Parent:             parentID,
						UncompressedDigest: layerDigest,
					}, nil
				},
			},
			expectError:    false,
			verifyParent:   true,
			expectedParent: "desired-parent-id",
		},
		{
			name:     "error from LayersByUncompressedDigest",
			diffID:   layerDigest,
			parentID: "",
			store: &mockStore{
				layersByUncompressedDigestErr: fmt.Errorf("storage lookup failed"),
			},
			expectError:   true,
			errorContains: "failed to look up layer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reusedLayer, err := reuseStorageLayer(tt.store, tt.diffID, tt.parentID, tmpDir, SilentLogger{})

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Fatalf("reuseStorageLayer failed: %v", err)
				}

				if reusedLayer.UncompressedDigest != tt.diffID {
					t.Errorf("digest mismatch: got %s, want %s", reusedLayer.UncompressedDigest, tt.diffID)
				}

				if tt.verifyParent && reusedLayer.Parent != tt.expectedParent {
					t.Errorf("parent mismatch: got %s, want %s", reusedLayer.Parent, tt.expectedParent)
				}
			}
		})
	}
}
