// oid/columnar_test.go
package oid

import (
	"encoding/json"
	"testing"
)

func TestNewColumnarOID(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name     string
		column   uint32
		indexes  []uint32
		expected string
	}{
		{
			name:     "Без индексов",
			column:   2,
			indexes:  nil,
			expected: "1.3.6.1.2.1.2.2.1.2",
		},
		{
			name:     "Один индекс",
			column:   2,
			indexes:  []uint32{1},
			expected: "1.3.6.1.2.1.2.2.1.2.1",
		},
		{
			name:     "Два индекса",
			column:   2,
			indexes:  []uint32{1, 2},
			expected: "1.3.6.1.2.1.2.2.1.2.1.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)
			if col.String() != tt.expected {
				t.Errorf("String = %s, ожидалось %s", col.String(), tt.expected)
			}
		})
	}
}

func TestParseColumnarOID(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name        string
		fullOID     OID
		wantColumn  uint32
		wantIndexes []uint32
		wantErr     bool
	}{
		{
			name:        "С одним индексом",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.1"),
			wantColumn:  2,
			wantIndexes: []uint32{1},
			wantErr:     false,
		},
		{
			name:        "С двумя индексами",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2"),
			wantColumn:  2,
			wantIndexes: []uint32{1, 2},
			wantErr:     false,
		},
		{
			name:        "Не принадлежит базе",
			fullOID:     MustParseOID("1.3.6.1.2.1.1.1.0"),
			wantColumn:  0,
			wantIndexes: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, err := ParseColumnarOID(base, tt.fullOID)
			if tt.wantErr {
				if err == nil {
					t.Error("ParseColumnarOID: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("ParseColumnarOID: %v", err)
			}
			if col.Column != tt.wantColumn {
				t.Errorf("Column = %d, ожидалось %d", col.Column, tt.wantColumn)
			}
			if len(col.Indexes) != len(tt.wantIndexes) {
				t.Errorf("len(Indexes) = %d, ожидалось %d", len(col.Indexes), len(tt.wantIndexes))
			}
		})
	}
}

func TestColumnarOIDMethods(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 7, 1, 2)

	// HasIndexes
	if !col.HasIndexes() {
		t.Error("HasIndexes: должно быть true")
	}

	// IndexString
	if col.IndexString() != "1.2" {
		t.Errorf("IndexString = %s, ожидалось '1.2'", col.IndexString())
	}

	// LastIndex
	last, err := col.LastIndex()
	if err != nil {
		t.Errorf("LastIndex: %v", err)
	}
	if last != 2 {
		t.Errorf("LastIndex = %d, ожидалось 2", last)
	}

	// Parent
	parent := col.Parent()
	if parent.IndexString() != "1" {
		t.Errorf("Parent.IndexString = %s, ожидалось '1'", parent.IndexString())
	}

	// AppendIndex
	extended := col.AppendIndex(3)
	if extended.IndexString() != "1.2.3" {
		t.Errorf("AppendIndex.IndexString = %s, ожидалось '1.2.3'", extended.IndexString())
	}

	// WithIndexes
	changed := col.WithIndexes(5, 6)
	if changed.IndexString() != "5.6" {
		t.Errorf("WithIndexes.IndexString = %s, ожидалось '5.6'", changed.IndexString())
	}

	// Equal
	col2 := NewColumnarOID(base, 7, 1, 2)
	if !col.Equal(col2) {
		t.Error("Equal: должны быть равны")
	}
}

func TestColumnarOIDJSON(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 7, 1)

	// Marshal
	data, err := json.Marshal(col)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(data) != `"1.3.6.1.2.1.2.2.1.7.1"` {
		t.Errorf("MarshalJSON = %s", data)
	}
}
