package oid

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestColumnarOIDStructure(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name            string
		col             ColumnarOID
		expectedBase    OID
		expectedColumn  uint32
		expectedIndexes []uint32
	}{
		{
			name: "Без индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			expectedBase:    base,
			expectedColumn:  2,
			expectedIndexes: nil,
		},
		{
			name: "Один индекс",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			expectedBase:    base,
			expectedColumn:  2,
			expectedIndexes: []uint32{1},
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			expectedBase:    base,
			expectedColumn:  2,
			expectedIndexes: []uint32{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.col.Base.Equal(tt.expectedBase) {
				t.Errorf("Base = %v, want %v", tt.col.Base, tt.expectedBase)
			}

			if tt.col.Column != tt.expectedColumn {
				t.Errorf("Column = %d, want %d", tt.col.Column, tt.expectedColumn)
			}

			if len(tt.col.Indexes) != len(tt.expectedIndexes) {
				t.Errorf("len(Indexes) = %d, want %d",
					len(tt.col.Indexes), len(tt.expectedIndexes))
				return
			}

			for i := range tt.col.Indexes {
				if tt.col.Indexes[i] != tt.expectedIndexes[i] {
					t.Errorf("Indexes[%d] = %d, want %d",
						i, tt.col.Indexes[i], tt.expectedIndexes[i])
				}
			}
		})
	}
}

func TestColumnarOIDZeroValue(t *testing.T) {
	var col ColumnarOID

	if col.Base != nil {
		t.Error("Base должно быть nil для нулевого значения")
	}

	if col.Column != 0 {
		t.Errorf("Column = %d, want 0", col.Column)
	}

	if col.Indexes != nil {
		t.Error("Indexes должны быть nil для нулевого значения")
	}
}

func TestColumnarOIDCopyIndependence(t *testing.T) {
	base := MustParseOID("1.3.6.1")
	col1 := ColumnarOID{
		Base:    base,
		Column:  2,
		Indexes: []uint32{1, 2},
	}

	// Создаем копию структуры
	col2 := col1

	// Изменяем копию
	col2.Column = 99
	col2.Base[0] = 99
	col2.Indexes[0] = 99

	// Оригинал не должен измениться (Base и Indexes - слайсы, общие)
	if col1.Column != 2 {
		t.Error("Column должен быть независимым")
	}

	// Base и Indexes - слайсы, они общие
	if col1.Base[0] != 99 {
		t.Error("Base должен быть общим (слайс)")
	}
	if col1.Indexes[0] != 99 {
		t.Error("Indexes должны быть общими (слайс)")
	}
}

// Пример использования
func ExampleColumnarOID() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col := ColumnarOID{
		Base:    base,
		Column:  2,
		Indexes: []uint32{1},
	}

	fmt.Println(col.Base)
	fmt.Println(col.Column)
	fmt.Println(col.Indexes)
	// Output:
	// 1.3.6.1.2.1.2.2.1
	// 2
	// [1]
}

// Бенчмарк
func BenchmarkColumnarOIDCreation(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = ColumnarOID{
			Base:    base,
			Column:  2,
			Indexes: []uint32{1},
		}
	}
}

func TestNewColumnarOID(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name           string
		base           OID
		column         uint32
		indexes        []uint32
		expectedColumn uint32
		expectedIdxLen int
	}{
		{
			name:           "Без индексов",
			base:           base,
			column:         2,
			indexes:        nil,
			expectedColumn: 2,
			expectedIdxLen: 0,
		},
		{
			name:           "Один индекс",
			base:           base,
			column:         2,
			indexes:        []uint32{1},
			expectedColumn: 2,
			expectedIdxLen: 1,
		},
		{
			name:           "Два индекса",
			base:           base,
			column:         2,
			indexes:        []uint32{1, 2},
			expectedColumn: 2,
			expectedIdxLen: 2,
		},
		{
			name:           "Пустая база",
			base:           OID{},
			column:         0,
			indexes:        nil,
			expectedColumn: 0,
			expectedIdxLen: 0,
		},
		{
			name:           "Nil база",
			base:           nil,
			column:         0,
			indexes:        nil,
			expectedColumn: 0,
			expectedIdxLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := NewColumnarOID(tt.base, tt.column, tt.indexes...)

			if !col.Base.Equal(tt.base) {
				t.Errorf("Base = %v, want %v", col.Base, tt.base)
			}

			if col.Column != tt.expectedColumn {
				t.Errorf("Column = %d, want %d", col.Column, tt.expectedColumn)
			}

			if len(col.Indexes) != tt.expectedIdxLen {
				t.Errorf("len(Indexes) = %d, want %d",
					len(col.Indexes), tt.expectedIdxLen)
			}

			// Проверяем индексы
			for i, idx := range tt.indexes {
				if col.Indexes[i] != idx {
					t.Errorf("Indexes[%d] = %d, want %d", i, col.Indexes[i], idx)
				}
			}
		})
	}
}

func TestNewColumnarOIDStoresReference(t *testing.T) {
	base := MustParseOID("1.3.6.1")
	indexes := []uint32{1, 2}

	col := NewColumnarOID(base, 2, indexes...)

	// Изменяем оригиналы
	base[0] = 99
	indexes[0] = 99

	// ColumnarOID должен хранить ссылки
	if col.Base[0] != 99 {
		t.Error("Base должен хранить ссылку")
	}
	if col.Indexes[0] != 99 {
		t.Error("Indexes должны хранить ссылку")
	}
}

func TestNewColumnarOIDConsistency(t *testing.T) {
	base := MustParseOID("1.3.6.1")

	// Создаем через NewColumnarOID
	col1 := NewColumnarOID(base, 2, 1)

	// Создаем вручную
	col2 := ColumnarOID{
		Base:    base,
		Column:  2,
		Indexes: []uint32{1},
	}

	if !col1.Equal(col2) {
		t.Error("NewColumnarOID должен давать тот же результат, что и ручное создание")
	}
}

// Пример использования
func ExampleNewColumnarOID() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col := NewColumnarOID(base, 2, 1)

	fmt.Println(col.Base)
	fmt.Println(col.Column)
	fmt.Println(col.Indexes)
	// Output:
	// 1.3.6.1.2.1.2.2.1
	// 2
	// [1]
}

// Бенчмарк
func BenchmarkNewColumnarOID(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = NewColumnarOID(base, 2, 1)
	}
}

func TestParseColumnarOID(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name        string
		fullOID     OID
		wantColumn  uint32
		wantIndexes []uint32
		wantErr     error
	}{
		{
			name:        "Без индексов",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2"),
			wantColumn:  2,
			wantIndexes: nil,
			wantErr:     nil,
		},
		{
			name:        "Один индекс",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.1"),
			wantColumn:  2,
			wantIndexes: []uint32{1},
			wantErr:     nil,
		},
		{
			name:        "Два индекса",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2"),
			wantColumn:  2,
			wantIndexes: []uint32{1, 2},
			wantErr:     nil,
		},
		{
			name:        "Три индекса",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2.3"),
			wantColumn:  2,
			wantIndexes: []uint32{1, 2, 3},
			wantErr:     nil,
		},
		{
			name:        "Не принадлежит базе",
			fullOID:     MustParseOID("1.3.6.1.2.1.1.1.0"),
			wantColumn:  0,
			wantIndexes: nil,
			wantErr:     ErrOIDNotInBase,
		},
		{
			name:        "Только база",
			fullOID:     MustParseOID("1.3.6.1.2.1.2.2.1"),
			wantColumn:  0,
			wantIndexes: nil,
			wantErr:     ErrNotEnoughComponents,
		},
		{
			name:        "Пустой OID",
			fullOID:     OID{},
			wantColumn:  0,
			wantIndexes: nil,
			wantErr:     ErrOIDNotInBase,
		},
		{
			name:        "Nil OID",
			fullOID:     nil,
			wantColumn:  0,
			wantIndexes: nil,
			wantErr:     ErrOIDNotInBase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, err := ParseColumnarOID(base, tt.fullOID)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseColumnarOID: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseColumnarOID = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseColumnarOID: %v", err)
				return
			}

			if col.Column != tt.wantColumn {
				t.Errorf("Column = %d, want %d", col.Column, tt.wantColumn)
			}

			if len(col.Indexes) != len(tt.wantIndexes) {
				t.Errorf("len(Indexes) = %d, want %d",
					len(col.Indexes), len(tt.wantIndexes))
				return
			}

			for i := range col.Indexes {
				if col.Indexes[i] != tt.wantIndexes[i] {
					t.Errorf("Indexes[%d] = %d, want %d",
						i, col.Indexes[i], tt.wantIndexes[i])
				}
			}
		})
	}
}

func TestParseColumnarOIDRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
		{column: 2, indexes: []uint32{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			// Создаем колумнарный OID
			col := NewColumnarOID(base, tt.column, tt.indexes...)

			// Получаем полный OID
			fullOID := col.FullOID()

			// Парсим обратно
			parsed, err := ParseColumnarOID(base, fullOID)
			if err != nil {
				t.Fatalf("ParseColumnarOID: %v", err)
			}

			if !parsed.Equal(col) {
				t.Errorf("Round trip: %v -> %v -> %v", col, fullOID, parsed)
			}
		})
	}
}

func TestParseColumnarOIDNotModifyInput(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	fullOID := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")

	fullOIDCopy := make(OID, len(fullOID))
	copy(fullOIDCopy, fullOID)

	ParseColumnarOID(base, fullOID)

	if !fullOID.Equal(fullOIDCopy) {
		t.Error("ParseColumnarOID не должен изменять входной OID")
	}
}

// Пример использования
func ExampleParseColumnarOID() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	fullOID := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")

	col, err := ParseColumnarOID(base, fullOID)
	if err != nil {
		panic(err)
	}

	fmt.Println(col.Column)
	fmt.Println(col.Indexes)
	// Output:
	// 2
	// [1]
}

// Пример с ошибкой
func ExampleParseColumnarOID_error() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	fullOID := MustParseOID("1.3.6.1.2.1.1.1.0")

	_, err := ParseColumnarOID(base, fullOID)
	fmt.Println(errors.Is(err, ErrOIDNotInBase))
	// Output: true
}

// Бенчмарк
func BenchmarkParseColumnarOID(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	fullOID := MustParseOID("1.3.6.1.2.1.2.2.1.2.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = ParseColumnarOID(base, fullOID)
	}
}

func TestMustColumnarOID(t *testing.T) {
	tests := []struct {
		name        string
		baseStr     string
		column      uint32
		indexes     []uint32
		expectedCol uint32
		expectedIdx []uint32
	}{
		{
			name:        "Без индексов",
			baseStr:     "1.3.6.1.2.1.2.2.1",
			column:      2,
			indexes:     nil,
			expectedCol: 2,
			expectedIdx: nil,
		},
		{
			name:        "Один индекс",
			baseStr:     "1.3.6.1.2.1.2.2.1",
			column:      2,
			indexes:     []uint32{1},
			expectedCol: 2,
			expectedIdx: []uint32{1},
		},
		{
			name:        "Два индекса",
			baseStr:     "1.3.6.1.2.1.2.2.1",
			column:      2,
			indexes:     []uint32{1, 2},
			expectedCol: 2,
			expectedIdx: []uint32{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := MustColumnarOID(tt.baseStr, tt.column, tt.indexes...)

			if col.Column != tt.expectedCol {
				t.Errorf("Column = %d, want %d", col.Column, tt.expectedCol)
			}

			if len(col.Indexes) != len(tt.expectedIdx) {
				t.Errorf("len(Indexes) = %d, want %d",
					len(col.Indexes), len(tt.expectedIdx))
				return
			}

			for i := range col.Indexes {
				if col.Indexes[i] != tt.expectedIdx[i] {
					t.Errorf("Indexes[%d] = %d, want %d",
						i, col.Indexes[i], tt.expectedIdx[i])
				}
			}
		})
	}
}

func TestMustColumnarOIDPanic(t *testing.T) {
	tests := []struct {
		name    string
		baseStr string
	}{
		{"Пустая строка", ""},
		{"Невалидный", "invalid"},
		{"Один компонент", "1"},
		{"Первый > 2", "3.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("MustColumnarOID(%q): expected panic", tt.baseStr)
				}
			}()

			MustColumnarOID(tt.baseStr, 2)
		})
	}
}

func TestMustColumnarOIDConsistency(t *testing.T) {
	baseStr := "1.3.6.1.2.1.2.2.1"
	column := uint32(2)
	indexes := []uint32{1}

	// Через MustColumnarOID
	col1 := MustColumnarOID(baseStr, column, indexes...)

	// Через MustParseOID + NewColumnarOID
	col2 := NewColumnarOID(MustParseOID(baseStr), column, indexes...)

	if !col1.Equal(col2) {
		t.Error("MustColumnarOID должен давать тот же результат")
	}
}

// Пример использования
func ExampleMustColumnarOID() {
	col := MustColumnarOID("1.3.6.1.2.1.2.2.1", 2, 1)

	fmt.Println(col.Base)
	fmt.Println(col.Column)
	fmt.Println(col.Indexes)
	// Output:
	// 1.3.6.1.2.1.2.2.1
	// 2
	// [1]
}

// Пример с паникой
func ExampleMustColumnarOID_panic() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Паника поймана")
		}
	}()

	MustColumnarOID("invalid", 2)
	// Output: Паника поймана
}

// Бенчмарк
func BenchmarkMustColumnarOID(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = MustColumnarOID("1.3.6.1.2.1.2.2.1", 2, 1)
	}
}

func TestColumnarOIDFullOID(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name     string
		col      ColumnarOID
		expected OID
	}{
		{
			name: "Без индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			expected: MustParseOID("1.3.6.1.2.1.2.2.1.2"),
		},
		{
			name: "Один индекс",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			expected: MustParseOID("1.3.6.1.2.1.2.2.1.2.1"),
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			expected: MustParseOID("1.3.6.1.2.1.2.2.1.2.1.2"),
		},
		{
			name: "Пустая база",
			col: ColumnarOID{
				Base:    OID{},
				Column:  0,
				Indexes: nil,
			},
			expected: OID{0},
		},
		{
			name: "Nil база",
			col: ColumnarOID{
				Base:    nil,
				Column:  0,
				Indexes: nil,
			},
			expected: OID{0},
		},
		{
			name:     "Пустая колонка",
			col:      ColumnarOID{},
			expected: OID{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.col.FullOID()

			if !result.Equal(tt.expected) {
				t.Errorf("FullOID() = %v, want %v", result, tt.expected)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("len = %d, want %d", len(result), len(tt.expected))
			}
		})
	}
}

func TestColumnarOIDFullOIDNewSlice(t *testing.T) {
	base := MustParseOID("1.3.6.1")
	indexes := []uint32{1, 2}

	col := ColumnarOID{
		Base:    base,
		Column:  2,
		Indexes: indexes,
	}

	result := col.FullOID()

	// Изменяем оригиналы
	base[0] = 99
	indexes[0] = 99

	// FullOID должен быть независимым
	if result[0] != 1 {
		t.Error("FullOID должен создать новый слайс")
	}
	if result[len(result)-1] != 2 {
		t.Error("FullOID должен создать новый слайс")
	}
}

func TestColumnarOIDFullOIDRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)
			fullOID := col.FullOID()

			// Парсим обратно
			parsed, err := ParseColumnarOID(base, fullOID)
			if err != nil {
				t.Fatalf("ParseColumnarOID: %v", err)
			}

			if !parsed.Equal(col) {
				t.Errorf("Round trip: %v -> %v -> %v", col, fullOID, parsed)
			}
		})
	}
}

func TestColumnarOIDFullOIDStartsWithBase(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col := NewColumnarOID(base, 2, 1)
	fullOID := col.FullOID()

	if !fullOID.StartsWith(base) {
		t.Error("FullOID должен начинаться с базы")
	}
}

// Пример использования
func ExampleColumnarOID_FullOID() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col := NewColumnarOID(base, 2, 1)

	fullOID := col.FullOID()

	fmt.Println(fullOID)
	// Output: 1.3.6.1.2.1.2.2.1.2.1
}

// Бенчмарк
func BenchmarkColumnarOIDFullOID(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col.FullOID()
	}
}

func TestColumnarOIDString(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name     string
		col      ColumnarOID
		expected string
	}{
		{
			name: "Без индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			expected: "1.3.6.1.2.1.2.2.1.2",
		},
		{
			name: "Один индекс",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			expected: "1.3.6.1.2.1.2.2.1.2.1",
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			expected: "1.3.6.1.2.1.2.2.1.2.1.2",
		},
		{
			name:     "Пустая колонка",
			col:      ColumnarOID{},
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.col.String()

			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestColumnarOIDStringRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)
			str := col.String()

			// Парсим строку обратно
			parsed, err := ParseOID(str)
			if err != nil {
				t.Fatalf("ParseOID(%q): %v", str, err)
			}

			// Парсим колумнарный
			parsedCol, err := ParseColumnarOID(base, parsed)
			if err != nil {
				t.Fatalf("ParseColumnarOID: %v", err)
			}

			if !parsedCol.Equal(col) {
				t.Errorf("Round trip: %v -> %q -> %v", col, str, parsedCol)
			}
		})
	}
}

func TestColumnarOIDStringNotModify(t *testing.T) {
	base := MustParseOID("1.3.6.1")
	col := NewColumnarOID(base, 2, 1)

	colCopy := ColumnarOID{
		Base:    make(OID, len(col.Base)),
		Column:  col.Column,
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Base, col.Base)
	copy(colCopy.Indexes, col.Indexes)

	col.String()

	if !col.Equal(colCopy) {
		t.Error("String() не должен изменять ColumnarOID")
	}
}

// Пример использования
func ExampleColumnarOID_String() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col := NewColumnarOID(base, 2, 1)

	fmt.Println(col.String())
	// Output: 1.3.6.1.2.1.2.2.1.2.1
}

// Бенчмарк
func BenchmarkColumnarOIDString(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col.String()
	}
}

func TestColumnarOIDIsValid(t *testing.T) {
	tests := []struct {
		name     string
		col      ColumnarOID
		expected bool
	}{
		{
			name: "Валидная база",
			col: ColumnarOID{
				Base:    MustParseOID("1.3.6.1.2.1.2.2.1"),
				Column:  2,
				Indexes: []uint32{1},
			},
			expected: true,
		},
		{
			name: "Короткая валидная база",
			col: ColumnarOID{
				Base:   MustParseOID("1.3.6.1"),
				Column: 1,
			},
			expected: true,
		},
		{
			name: "Пустая база",
			col: ColumnarOID{
				Base:   OID{},
				Column: 0,
			},
			expected: false,
		},
		{
			name: "Nil база",
			col: ColumnarOID{
				Base:   nil,
				Column: 0,
			},
			expected: false,
		},
		{
			name: "Один компонент база",
			col: ColumnarOID{
				Base:   OID{1},
				Column: 0,
			},
			expected: false,
		},
		{
			name: "Первый > 2",
			col: ColumnarOID{
				Base:   OID{3, 1},
				Column: 0,
			},
			expected: false,
		},
		{
			name: "Второй > 39",
			col: ColumnarOID{
				Base:   OID{1, 40},
				Column: 0,
			},
			expected: false,
		},
		{
			name:     "Пустая колонка",
			col:      ColumnarOID{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.col.IsValid()

			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestColumnarOIDIsValidNotModify(t *testing.T) {
	col := ColumnarOID{
		Base:    MustParseOID("1.3.6.1"),
		Column:  2,
		Indexes: []uint32{1},
	}

	colCopy := ColumnarOID{
		Base:    make(OID, len(col.Base)),
		Column:  col.Column,
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Base, col.Base)
	copy(colCopy.Indexes, col.Indexes)

	col.IsValid()

	if !col.Equal(colCopy) {
		t.Error("IsValid() не должен изменять ColumnarOID")
	}
}

// Пример использования
func ExampleColumnarOID_IsValid() {
	validCol := ColumnarOID{
		Base:   MustParseOID("1.3.6.1"),
		Column: 2,
	}

	invalidCol := ColumnarOID{}

	fmt.Println(validCol.IsValid())
	fmt.Println(invalidCol.IsValid())
	// Output:
	// true
	// false
}

// Бенчмарк
func BenchmarkColumnarOIDIsValid(b *testing.B) {
	col := ColumnarOID{
		Base:    MustParseOID("1.3.6.1.2.1.2.2.1"),
		Column:  2,
		Indexes: []uint32{1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col.IsValid()
	}
}

func TestColumnarOIDHasIndexes(t *testing.T) {
	tests := []struct {
		name     string
		col      ColumnarOID
		expected bool
	}{
		{
			name: "С одним индексом",
			col: ColumnarOID{
				Indexes: []uint32{1},
			},
			expected: true,
		},
		{
			name: "С двумя индексами",
			col: ColumnarOID{
				Indexes: []uint32{1, 2},
			},
			expected: true,
		},
		{
			name: "Без индексов (nil)",
			col: ColumnarOID{
				Indexes: nil,
			},
			expected: false,
		},
		{
			name: "Без индексов (пустой)",
			col: ColumnarOID{
				Indexes: []uint32{},
			},
			expected: false,
		},
		{
			name:     "Пустая колонка",
			col:      ColumnarOID{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.col.HasIndexes()

			if result != tt.expected {
				t.Errorf("HasIndexes() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestColumnarOIDHasIndexesNotModify(t *testing.T) {
	col := ColumnarOID{
		Indexes: []uint32{1, 2},
	}

	colCopy := ColumnarOID{
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Indexes, col.Indexes)

	col.HasIndexes()

	if !col.Equal(colCopy) {
		t.Error("HasIndexes() не должен изменять ColumnarOID")
	}
}

func TestColumnarOIDHasIndexesAfterAppend(t *testing.T) {
	col := ColumnarOID{
		Indexes: nil,
	}

	if col.HasIndexes() {
		t.Error("Без индексов HasIndexes должен вернуть false")
	}

	col = col.AppendIndex(1)

	if !col.HasIndexes() {
		t.Error("После AppendIndex HasIndexes должен вернуть true")
	}
}

// Пример использования
func ExampleColumnarOID_HasIndexes() {
	withIndexes := ColumnarOID{
		Indexes: []uint32{1},
	}

	withoutIndexes := ColumnarOID{}

	fmt.Println(withIndexes.HasIndexes())
	fmt.Println(withoutIndexes.HasIndexes())
	// Output:
	// true
	// false
}

// Бенчмарк
func BenchmarkColumnarOIDHasIndexes(b *testing.B) {
	col := ColumnarOID{
		Indexes: []uint32{1, 2},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col.HasIndexes()
	}
}

func TestColumnarOIDIndexString(t *testing.T) {
	tests := []struct {
		name     string
		col      ColumnarOID
		expected string
	}{
		{
			name: "Без индексов (nil)",
			col: ColumnarOID{
				Indexes: nil,
			},
			expected: "",
		},
		{
			name: "Без индексов (пустой)",
			col: ColumnarOID{
				Indexes: []uint32{},
			},
			expected: "",
		},
		{
			name: "Один индекс",
			col: ColumnarOID{
				Indexes: []uint32{1},
			},
			expected: "1",
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Indexes: []uint32{1, 2},
			},
			expected: "1.2",
		},
		{
			name: "Три индекса",
			col: ColumnarOID{
				Indexes: []uint32{1, 2, 3},
			},
			expected: "1.2.3",
		},
		{
			name: "С нулями",
			col: ColumnarOID{
				Indexes: []uint32{0, 0},
			},
			expected: "0.0",
		},
		{
			name: "С большими числами",
			col: ColumnarOID{
				Indexes: []uint32{MaxOIDComponent},
			},
			expected: "268435455",
		},
		{
			name:     "Пустая колонка",
			col:      ColumnarOID{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.col.IndexString()

			if result != tt.expected {
				t.Errorf("IndexString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestColumnarOIDIndexStringNotModify(t *testing.T) {
	col := ColumnarOID{
		Indexes: []uint32{1, 2},
	}

	colCopy := ColumnarOID{
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Indexes, col.Indexes)

	col.IndexString()

	if !col.Equal(colCopy) {
		t.Error("IndexString() не должен изменять ColumnarOID")
	}
}

func TestColumnarOIDIndexStringRoundTrip(t *testing.T) {
	tests := [][]uint32{
		{},
		{1},
		{1, 2},
		{1, 2, 3},
	}

	for _, indexes := range tests {
		t.Run(fmt.Sprintf("idx=%v", indexes), func(t *testing.T) {
			col := ColumnarOID{Indexes: indexes}
			str := col.IndexString()

			// Проверяем, что строка соответствует
			if len(indexes) == 0 {
				if str != "" {
					t.Error("IndexString должен вернуть пустую строку")
				}
			} else {
				if str == "" {
					t.Error("IndexString не должен быть пустым")
				}
			}
		})
	}
}

// Пример использования
func ExampleColumnarOID_IndexString() {
	withIndexes := ColumnarOID{
		Indexes: []uint32{1, 2},
	}

	withoutIndexes := ColumnarOID{}

	fmt.Println(withIndexes.IndexString())
	fmt.Println(withoutIndexes.IndexString())
	// Output:
	// 1.2
	//
}

// Бенчмарк
func BenchmarkColumnarOIDIndexString(b *testing.B) {
	col := ColumnarOID{
		Indexes: []uint32{1, 2},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col.IndexString()
	}
}

func TestColumnarOIDWithIndexes(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name           string
		col            ColumnarOID
		indexes        []uint32
		expectedIdxLen int
		expectedIdx    []uint32
	}{
		{
			name: "Добавление индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			indexes:        []uint32{1},
			expectedIdxLen: 1,
			expectedIdx:    []uint32{1},
		},
		{
			name: "Замена индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			indexes:        []uint32{2, 3},
			expectedIdxLen: 2,
			expectedIdx:    []uint32{2, 3},
		},
		{
			name: "Очистка индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			indexes:        nil,
			expectedIdxLen: 0,
			expectedIdx:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.col.WithIndexes(tt.indexes...)

			// Проверяем, что Base и Column сохранились
			if !result.Base.Equal(tt.col.Base) {
				t.Errorf("Base = %v, want %v", result.Base, tt.col.Base)
			}
			if result.Column != tt.col.Column {
				t.Errorf("Column = %d, want %d", result.Column, tt.col.Column)
			}

			// Проверяем индексы
			if len(result.Indexes) != tt.expectedIdxLen {
				t.Errorf("len(Indexes) = %d, want %d",
					len(result.Indexes), tt.expectedIdxLen)
			}

			for i := range result.Indexes {
				if result.Indexes[i] != tt.expectedIdx[i] {
					t.Errorf("Indexes[%d] = %d, want %d",
						i, result.Indexes[i], tt.expectedIdx[i])
				}
			}
		})
	}
}

func TestColumnarOIDWithIndexesNotModifyOriginal(t *testing.T) {
	base := MustParseOID("1.3.6.1")
	col := ColumnarOID{
		Base:    base,
		Column:  2,
		Indexes: []uint32{1},
	}

	colCopy := ColumnarOID{
		Base:    make(OID, len(col.Base)),
		Column:  col.Column,
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Base, col.Base)
	copy(colCopy.Indexes, col.Indexes)

	col.WithIndexes(2, 3)

	if !col.Equal(colCopy) {
		t.Error("WithIndexes() не должен изменять оригинал")
	}
}

func TestColumnarOIDWithIndexesSharesBase(t *testing.T) {
	base := MustParseOID("1.3.6.1")
	col := ColumnarOID{
		Base:   base,
		Column: 2,
	}

	result := col.WithIndexes(1)

	// Base - слайс, должен быть общим
	if &result.Base[0] != &col.Base[0] {
		t.Error("Base должен быть общим (слайс)")
	}
}

// Пример использования
func ExampleColumnarOID_WithIndexes() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	col := NewColumnarOID(base, 2)

	withIndexes := col.WithIndexes(1, 2)

	fmt.Println(col.IndexString())
	fmt.Println(withIndexes.IndexString())
	// Output:
	//
	// 1.2
}

// Бенчмарк
func BenchmarkColumnarOIDWithIndexes(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col.WithIndexes(1)
	}
}

func TestColumnarOIDAppendIndex(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name           string
		col            ColumnarOID
		index          uint32
		expectedIdxLen int
		expectedLast   uint32
	}{
		{
			name: "Добавление к пустым индексам",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			index:          1,
			expectedIdxLen: 1,
			expectedLast:   1,
		},
		{
			name: "Добавление к одному индексу",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			index:          2,
			expectedIdxLen: 2,
			expectedLast:   2,
		},
		{
			name: "Добавление к двум индексам",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			index:          3,
			expectedIdxLen: 3,
			expectedLast:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.col.AppendIndex(tt.index)

			// Проверяем Base и Column
			if !result.Base.Equal(tt.col.Base) {
				t.Errorf("Base = %v, want %v", result.Base, tt.col.Base)
			}
			if result.Column != tt.col.Column {
				t.Errorf("Column = %d, want %d", result.Column, tt.col.Column)
			}

			// Проверяем длину
			if len(result.Indexes) != tt.expectedIdxLen {
				t.Errorf("len(Indexes) = %d, want %d",
					len(result.Indexes), tt.expectedIdxLen)
			}

			// Проверяем последний индекс
			if len(result.Indexes) > 0 {
				if result.Indexes[len(result.Indexes)-1] != tt.expectedLast {
					t.Errorf("Last index = %d, want %d",
						result.Indexes[len(result.Indexes)-1], tt.expectedLast)
				}
			}
		})
	}
}

func TestColumnarOIDAppendIndexNotModifyOriginal(t *testing.T) {
	base := MustParseOID("1.3.6.1")
	col := ColumnarOID{
		Base:    base,
		Column:  2,
		Indexes: []uint32{1},
	}

	colCopy := ColumnarOID{
		Base:    make(OID, len(col.Base)),
		Column:  col.Column,
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Base, col.Base)
	copy(colCopy.Indexes, col.Indexes)

	col.AppendIndex(2)

	if !col.Equal(colCopy) {
		t.Error("AppendIndex() не должен изменять оригинал")
	}
}

func TestColumnarOIDAppendIndexNewSlice(t *testing.T) {
	col := ColumnarOID{
		Indexes: []uint32{1},
	}

	result := col.AppendIndex(2)

	// Изменяем оригинальные индексы
	col.Indexes[0] = 99

	// Результат не должен измениться
	if result.Indexes[0] != 1 {
		t.Error("AppendIndex должен создать новый слайс")
	}
}

func TestColumnarOIDAppendIndexChain(t *testing.T) {
	base := MustParseOID("1.3.6.1")

	col := NewColumnarOID(base, 2)
	col = col.AppendIndex(1)
	col = col.AppendIndex(2)
	col = col.AppendIndex(3)

	if len(col.Indexes) != 3 {
		t.Errorf("len(Indexes) = %d, want 3", len(col.Indexes))
	}

	if col.IndexString() != "1.2.3" {
		t.Errorf("IndexString() = %q, want '1.2.3'", col.IndexString())
	}
}

// Пример использования
func ExampleColumnarOID_AppendIndex() {
	base := MustParseOID("1.3.6.1")

	col := NewColumnarOID(base, 2)

	col = col.AppendIndex(1)
	col = col.AppendIndex(2)

	fmt.Println(col.IndexString())
	// Output: 1.2
}

// Бенчмарк
func BenchmarkColumnarOIDAppendIndex(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col.AppendIndex(1)
	}
}

func TestColumnarOIDParent(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name           string
		col            ColumnarOID
		expectedIdxLen int
		expectedIdx    []uint32
	}{
		{
			name: "Без индексов (nil)",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			expectedIdxLen: 0,
			expectedIdx:    nil,
		},
		{
			name: "Без индексов (пустой)",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{},
			},
			expectedIdxLen: 0,
			expectedIdx:    nil,
		},
		{
			name: "Один индекс",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			expectedIdxLen: 0,
			expectedIdx:    nil,
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			expectedIdxLen: 1,
			expectedIdx:    []uint32{1},
		},
		{
			name: "Три индекса",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2, 3},
			},
			expectedIdxLen: 2,
			expectedIdx:    []uint32{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.col.Parent()

			// Проверяем Base и Column
			if !result.Base.Equal(tt.col.Base) {
				t.Errorf("Base = %v, want %v", result.Base, tt.col.Base)
			}
			if result.Column != tt.col.Column {
				t.Errorf("Column = %d, want %d", result.Column, tt.col.Column)
			}

			// Проверяем индексы
			if len(result.Indexes) != tt.expectedIdxLen {
				t.Errorf("len(Indexes) = %d, want %d",
					len(result.Indexes), tt.expectedIdxLen)
			}

			for i := range result.Indexes {
				if result.Indexes[i] != tt.expectedIdx[i] {
					t.Errorf("Indexes[%d] = %d, want %d",
						i, result.Indexes[i], tt.expectedIdx[i])
				}
			}
		})
	}
}

func TestColumnarOIDParentNotModifyOriginal(t *testing.T) {
	base := MustParseOID("1.3.6.1")
	col := ColumnarOID{
		Base:    base,
		Column:  2,
		Indexes: []uint32{1, 2},
	}

	colCopy := ColumnarOID{
		Base:    make(OID, len(col.Base)),
		Column:  col.Column,
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Base, col.Base)
	copy(colCopy.Indexes, col.Indexes)

	col.Parent()

	if !col.Equal(colCopy) {
		t.Error("Parent() не должен изменять оригинал")
	}
}

func TestColumnarOIDParentChain(t *testing.T) {
	base := MustParseOID("1.3.6.1")

	col := NewColumnarOID(base, 2, 1, 2, 3)

	parent1 := col.Parent()
	parent2 := parent1.Parent()
	parent3 := parent2.Parent()

	if col.IndexString() != "1.2.3" {
		t.Error("col должен иметь все индексы")
	}
	if parent1.IndexString() != "1.2" {
		t.Error("parent1 должен иметь индексы 1.2")
	}
	if parent2.IndexString() != "1" {
		t.Error("parent2 должен иметь индекс 1")
	}
	if parent3.IndexString() != "" {
		t.Error("parent3 должен быть без индексов")
	}
}

func TestColumnarOIDParentSharesIndexes(t *testing.T) {
	col := ColumnarOID{
		Indexes: []uint32{1, 2},
	}

	parent := col.Parent()

	// Parent должен делить слайс с оригиналом
	if &parent.Indexes[0] != &col.Indexes[0] {
		t.Error("Parent должен делить слайс индексов")
	}
}

// Пример использования
func ExampleColumnarOID_Parent() {
	base := MustParseOID("1.3.6.1")

	col := NewColumnarOID(base, 2, 1, 2)

	parent := col.Parent()

	fmt.Println(col.IndexString())
	fmt.Println(parent.IndexString())
	// Output:
	// 1.2
	// 1
}

// Бенчмарк
func BenchmarkColumnarOIDParent(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1, 2)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col.Parent()
	}
}

func TestColumnarOIDLastIndex(t *testing.T) {
	tests := []struct {
		name     string
		col      ColumnarOID
		expected uint32
		wantErr  error
	}{
		{
			name: "Один индекс",
			col: ColumnarOID{
				Indexes: []uint32{1},
			},
			expected: 1,
			wantErr:  nil,
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Indexes: []uint32{1, 2},
			},
			expected: 2,
			wantErr:  nil,
		},
		{
			name: "Три индекса",
			col: ColumnarOID{
				Indexes: []uint32{1, 2, 3},
			},
			expected: 3,
			wantErr:  nil,
		},
		{
			name: "С нулем",
			col: ColumnarOID{
				Indexes: []uint32{1, 0},
			},
			expected: 0,
			wantErr:  nil,
		},
		{
			name: "С MaxOIDComponent",
			col: ColumnarOID{
				Indexes: []uint32{MaxOIDComponent},
			},
			expected: MaxOIDComponent,
			wantErr:  nil,
		},
		{
			name: "Без индексов (nil)",
			col: ColumnarOID{
				Indexes: nil,
			},
			expected: 0,
			wantErr:  ErrNoIndexes,
		},
		{
			name: "Без индексов (пустой)",
			col: ColumnarOID{
				Indexes: []uint32{},
			},
			expected: 0,
			wantErr:  ErrNoIndexes,
		},
		{
			name:     "Пустая колонка",
			col:      ColumnarOID{},
			expected: 0,
			wantErr:  ErrNoIndexes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.col.LastIndex()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("LastIndex: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("LastIndex = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("LastIndex: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("LastIndex() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestColumnarOIDLastIndexNotModify(t *testing.T) {
	col := ColumnarOID{
		Indexes: []uint32{1, 2},
	}

	colCopy := ColumnarOID{
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Indexes, col.Indexes)

	col.LastIndex()

	if !col.Equal(colCopy) {
		t.Error("LastIndex() не должен изменять ColumnarOID")
	}
}

func TestColumnarOIDLastIndexAfterAppend(t *testing.T) {
	col := ColumnarOID{
		Indexes: []uint32{1},
	}

	col = col.AppendIndex(2)

	last, err := col.LastIndex()
	if err != nil {
		t.Fatalf("LastIndex: %v", err)
	}

	if last != 2 {
		t.Errorf("LastIndex = %d, want 2", last)
	}
}

// Пример использования
func ExampleColumnarOID_LastIndex() {
	col := ColumnarOID{
		Indexes: []uint32{1, 2},
	}

	last, err := col.LastIndex()
	if err != nil {
		panic(err)
	}

	fmt.Println(last)
	// Output: 2
}

// Пример с ошибкой
func ExampleColumnarOID_LastIndex_error() {
	col := ColumnarOID{}

	_, err := col.LastIndex()
	fmt.Println(errors.Is(err, ErrNoIndexes))
	// Output: true
}

// Бенчмарк
func BenchmarkColumnarOIDLastIndex(b *testing.B) {
	col := ColumnarOID{
		Indexes: []uint32{1, 2},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = col.LastIndex()
	}
}

func TestColumnarOIDValidate(t *testing.T) {
	tests := []struct {
		name    string
		col     ColumnarOID
		wantErr error
	}{
		{
			name: "Валидная база",
			col: ColumnarOID{
				Base:    MustParseOID("1.3.6.1.2.1.2.2.1"),
				Column:  2,
				Indexes: []uint32{1},
			},
			wantErr: nil,
		},
		{
			name: "Короткая валидная база",
			col: ColumnarOID{
				Base:   MustParseOID("1.3.6.1"),
				Column: 1,
			},
			wantErr: nil,
		},
		{
			name: "Пустая база",
			col: ColumnarOID{
				Base:   OID{},
				Column: 0,
			},
			wantErr: ErrOIDTooShort,
		},
		{
			name: "Nil база",
			col: ColumnarOID{
				Base:   nil,
				Column: 0,
			},
			wantErr: ErrOIDTooShort,
		},
		{
			name: "Один компонент база",
			col: ColumnarOID{
				Base:   OID{1},
				Column: 0,
			},
			wantErr: ErrOIDTooShort,
		},
		{
			name: "Первый > 2",
			col: ColumnarOID{
				Base:   OID{3, 1},
				Column: 0,
			},
			wantErr: ErrFirstComponentTooBig,
		},
		{
			name: "Второй > 39",
			col: ColumnarOID{
				Base:   OID{1, 40},
				Column: 0,
			},
			wantErr: ErrSecondComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.col.Validate()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Validate: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate: %v", err)
			}
		})
	}
}

func TestColumnarOIDValidateNotModify(t *testing.T) {
	col := ColumnarOID{
		Base:    MustParseOID("1.3.6.1"),
		Column:  2,
		Indexes: []uint32{1},
	}

	colCopy := ColumnarOID{
		Base:    make(OID, len(col.Base)),
		Column:  col.Column,
		Indexes: make([]uint32, len(col.Indexes)),
	}
	copy(colCopy.Base, col.Base)
	copy(colCopy.Indexes, col.Indexes)

	col.Validate()

	if !col.Equal(colCopy) {
		t.Error("Validate() не должен изменять ColumnarOID")
	}
}

func TestColumnarOIDValidateConsistency(t *testing.T) {
	col := ColumnarOID{
		Base:   MustParseOID("1.3.6.1"),
		Column: 2,
	}

	// Validate должен совпадать с Base.Validate
	err := col.Validate()
	baseErr := col.Base.Validate()

	if (err == nil) != (baseErr == nil) {
		t.Error("Validate должен совпадать с Base.Validate")
	}
}

// Пример использования
func ExampleColumnarOID_Validate() {
	validCol := ColumnarOID{
		Base:   MustParseOID("1.3.6.1"),
		Column: 2,
	}

	invalidCol := ColumnarOID{}

	fmt.Println(validCol.Validate() == nil)
	fmt.Println(errors.Is(invalidCol.Validate(), ErrOIDTooShort))
	// Output:
	// true
	// true
}

// Бенчмарк
func BenchmarkColumnarOIDValidate(b *testing.B) {
	col := ColumnarOID{
		Base:    MustParseOID("1.3.6.1.2.1.2.2.1"),
		Column:  2,
		Indexes: []uint32{1},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col.Validate()
	}
}

func TestColumnarOIDEqual(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name     string
		col1     ColumnarOID
		col2     ColumnarOID
		expected bool
	}{
		{
			name: "Одинаковые без индексов",
			col1: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			col2: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			expected: true,
		},
		{
			name: "Одинаковые с индексами",
			col1: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			col2: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			expected: true,
		},
		{
			name: "Разные Column",
			col1: ColumnarOID{
				Base:   base,
				Column: 2,
			},
			col2: ColumnarOID{
				Base:   base,
				Column: 3,
			},
			expected: false,
		},
		{
			name: "Разные индексы",
			col1: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			col2: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{2},
			},
			expected: false,
		},
		{
			name: "Разная длина индексов",
			col1: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			col2: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			expected: false,
		},
		{
			name: "Разные Base",
			col1: ColumnarOID{
				Base:   base,
				Column: 2,
			},
			col2: ColumnarOID{
				Base:   MustParseOID("1.3.6.1.2.1.2.2.2"),
				Column: 2,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.col1.Equal(tt.col2)

			if result != tt.expected {
				t.Errorf("Equal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestColumnarOIDEqualProperties(t *testing.T) {
	base := MustParseOID("1.3.6.1")

	t.Run("Рефлексивность", func(t *testing.T) {
		col := NewColumnarOID(base, 2, 1)

		if !col.Equal(col) {
			t.Error("Equal(col, col) = false, want true")
		}
	})

	t.Run("Симметричность", func(t *testing.T) {
		col1 := NewColumnarOID(base, 2, 1)
		col2 := NewColumnarOID(base, 2, 1)

		if col1.Equal(col2) != col2.Equal(col1) {
			t.Error("Equal должен быть симметричным")
		}
	})

	t.Run("Транзитивность", func(t *testing.T) {
		col1 := NewColumnarOID(base, 2, 1)
		col2 := NewColumnarOID(base, 2, 1)
		col3 := NewColumnarOID(base, 2, 1)

		if col1.Equal(col2) && col2.Equal(col3) {
			if !col1.Equal(col3) {
				t.Error("Equal должен быть транзитивным")
			}
		}
	})

	t.Run("Не изменяет ColumnarOID", func(t *testing.T) {
		col1 := NewColumnarOID(base, 2, 1)
		col2 := NewColumnarOID(base, 2, 1)

		col1Copy := ColumnarOID{
			Base:    make(OID, len(col1.Base)),
			Column:  col1.Column,
			Indexes: make([]uint32, len(col1.Indexes)),
		}
		copy(col1Copy.Base, col1.Base)
		copy(col1Copy.Indexes, col1.Indexes)

		col1.Equal(col2)

		if !col1.Equal(col1Copy) {
			t.Error("Equal() не должен изменять ColumnarOID")
		}
	})
}

// Пример использования
func ExampleColumnarOID_Equal() {
	base := MustParseOID("1.3.6.1")

	col1 := NewColumnarOID(base, 2, 1)
	col2 := NewColumnarOID(base, 2, 1)
	col3 := NewColumnarOID(base, 3, 1)

	fmt.Println(col1.Equal(col2))
	fmt.Println(col1.Equal(col3))
	// Output:
	// true
	// false
}

// Бенчмарк
func BenchmarkColumnarOIDEqual(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col1 := NewColumnarOID(base, 2, 1)
	col2 := NewColumnarOID(base, 2, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = col1.Equal(col2)
	}
}

func TestColumnarOIDMarshalBinary(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name    string
		col     ColumnarOID
		wantErr error
	}{
		{
			name: "Без индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			wantErr: nil,
		},
		{
			name: "Один индекс",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			wantErr: nil,
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			wantErr: nil,
		},
		{
			name: "Пустая база",
			col: ColumnarOID{
				Base:   OID{},
				Column: 0,
			},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Пустая колонка",
			col:     ColumnarOID{},
			wantErr: ErrOIDTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.col.MarshalBinary()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("MarshalBinary: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("MarshalBinary = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("MarshalBinary: %v", err)
				return
			}

			if len(data) == 0 {
				t.Error("MarshalBinary: пустой результат")
			}
		})
	}
}

func TestColumnarOIDMarshalBinaryRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)

			// Кодируем
			data, err := col.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// Декодируем как OID
			var oid OID
			if err := oid.UnmarshalBinary(data); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}

			// Парсим колумнарный
			parsed, err := ParseColumnarOID(base, oid)
			if err != nil {
				t.Fatalf("ParseColumnarOID: %v", err)
			}

			if !parsed.Equal(col) {
				t.Errorf("Round trip: %v -> %x -> %v", col, data, parsed)
			}
		})
	}
}

func TestColumnarOIDMarshalBinaryCompareWithFullOID(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	// MarshalBinary колумнарного
	colData, err := col.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	// MarshalBinary полного OID
	fullData, err := col.FullOID().MarshalBinary()
	if err != nil {
		t.Fatalf("FullOID.MarshalBinary: %v", err)
	}

	if !bytes.Equal(colData, fullData) {
		t.Errorf("MarshalBinary = %x, want %x", colData, fullData)
	}
}

// Пример использования
func ExampleColumnarOID_MarshalBinary() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	data, err := col.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", data)
	// Output: 060a2b060102010202010201
}

// Бенчмарк
func BenchmarkColumnarOIDMarshalBinary(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = col.MarshalBinary()
	}
}

func TestColumnarOIDUnmarshalBinary(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	// Создаем правильные данные через MarshalBinary
	colNoIndexes := NewColumnarOID(base, 2)
	dataNoIndexes, _ := colNoIndexes.MarshalBinary()

	tests := []struct {
		name     string
		data     []byte
		expected ColumnarOID
		wantErr  error
	}{
		{
			name:     "Без индексов",
			data:     dataNoIndexes,
			expected: colNoIndexes,
			wantErr:  nil,
		},
		{
			name:     "Пустые данные",
			data:     []byte{},
			expected: ColumnarOID{},
			wantErr:  ErrDataTooShort,
		},
		{
			name:     "Неверный тег",
			data:     []byte{0x05, 0x00},
			expected: ColumnarOID{},
			wantErr:  ErrInvalidASN1Tag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var col ColumnarOID
			err := col.UnmarshalBinary(tt.data, base)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UnmarshalBinary: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UnmarshalBinary = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalBinary: %v", err)
				return
			}

			if !col.Equal(tt.expected) {
				t.Errorf("UnmarshalBinary = %v, want %v", col, tt.expected)
			}
		})
	}
}

func TestColumnarOIDUnmarshalBinaryNotInBase(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	// Создаем OID, который не принадлежит базе
	notInBaseOID := MustParseOID("1.3.6.1.2.1.1.1.0")
	data, _ := notInBaseOID.MarshalBinary()

	var col ColumnarOID
	err := col.UnmarshalBinary(data, base)

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrOIDNotInBase) {
		t.Errorf("UnmarshalBinary = %v, want ErrOIDNotInBase", err)
	}
}

func TestColumnarOIDUnmarshalBinaryNotEnoughComponents(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	// Создаем OID, который равен только базе (без колонки)
	baseOnlyOID := base
	data, _ := baseOnlyOID.MarshalBinary()

	var col ColumnarOID
	err := col.UnmarshalBinary(data, base)

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrNotEnoughComponents) {
		t.Errorf("UnmarshalBinary = %v, want ErrNotEnoughComponents", err)
	}
}

func TestColumnarOIDUnmarshalBinaryRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)

			// Кодируем
			data, err := col.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// Декодируем
			var decoded ColumnarOID
			if err := decoded.UnmarshalBinary(data, base); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}

			if !decoded.Equal(col) {
				t.Errorf("Round trip: %v -> %x -> %v", col, data, decoded)
			}
		})
	}
}

// Пример использования
func ExampleColumnarOID_UnmarshalBinary() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	// Создаем корректные данные через MarshalBinary
	col := NewColumnarOID(base, 2, 1)
	data, _ := col.MarshalBinary()

	var decoded ColumnarOID
	if err := decoded.UnmarshalBinary(data, base); err != nil {
		panic(err)
	}

	fmt.Println(decoded.Column)
	fmt.Println(decoded.Indexes)
	// Output:
	// 2
	// [1]
}

// Бенчмарк
func BenchmarkColumnarOIDUnmarshalBinary(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)
	data, _ := col.MarshalBinary()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var decoded ColumnarOID
		_ = decoded.UnmarshalBinary(data, base)
	}
}

func TestColumnarOIDMarshalBER(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name    string
		col     ColumnarOID
		wantErr error
	}{
		{
			name: "Без индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			wantErr: nil,
		},
		{
			name: "Один индекс",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			wantErr: nil,
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			wantErr: nil,
		},
		{
			name: "Пустая база",
			col: ColumnarOID{
				Base:   OID{},
				Column: 0,
			},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Пустая колонка",
			col:     ColumnarOID{},
			wantErr: ErrOIDTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.col.MarshalBER()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("MarshalBER: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("MarshalBER = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("MarshalBER: %v", err)
				return
			}

			if len(data) == 0 {
				t.Error("MarshalBER: пустой результат")
			}
		})
	}
}

func TestColumnarOIDMarshalBERRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)

			// Кодируем
			data, err := col.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}

			// Декодируем как OID
			var oid OID
			if err := oid.UnmarshalBER(data); err != nil {
				t.Fatalf("UnmarshalBER: %v", err)
			}

			// Парсим колумнарный
			parsed, err := ParseColumnarOID(base, oid)
			if err != nil {
				t.Fatalf("ParseColumnarOID: %v", err)
			}

			if !parsed.Equal(col) {
				t.Errorf("Round trip: %v -> %x -> %v", col, data, parsed)
			}
		})
	}
}

func TestColumnarOIDMarshalBERCompareWithFullOID(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	// MarshalBER колумнарного
	colData, err := col.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	// MarshalBER полного OID
	fullData, err := col.FullOID().MarshalBER()
	if err != nil {
		t.Fatalf("FullOID.MarshalBER: %v", err)
	}

	if !bytes.Equal(colData, fullData) {
		t.Errorf("MarshalBER = %x, want %x", colData, fullData)
	}
}

func TestColumnarOIDMarshalBERCompareWithBinary(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	berData, err := col.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	binData, err := col.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	// Для коротких OID BER == DER
	if !bytes.Equal(berData, binData) {
		t.Errorf("BER = %x, Binary = %x", berData, binData)
	}
}

// Пример использования
func ExampleColumnarOID_MarshalBER() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	data, err := col.MarshalBER()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", data)
	// Output: 060a2b060102010202010201
}

// Бенчмарк
func BenchmarkColumnarOIDMarshalBER(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = col.MarshalBER()
	}
}

func TestColumnarOIDMarshalJSON(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name     string
		col      ColumnarOID
		expected string
	}{
		{
			name: "Без индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			expected: `"1.3.6.1.2.1.2.2.1.2"`,
		},
		{
			name: "Один индекс",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			expected: `"1.3.6.1.2.1.2.2.1.2.1"`,
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			expected: `"1.3.6.1.2.1.2.2.1.2.1.2"`,
		},
		{
			name:     "Пустая колонка",
			col:      ColumnarOID{},
			expected: `"0"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.col.MarshalJSON()

			if err != nil {
				t.Errorf("MarshalJSON: %v", err)
				return
			}

			if string(data) != tt.expected {
				t.Errorf("MarshalJSON = %s, want %s", data, tt.expected)
			}
		})
	}
}

func TestColumnarOIDMarshalJSONValid(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	data, err := col.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if !json.Valid(data) {
		t.Errorf("MarshalJSON = %s, невалидный JSON", data)
	}
}

func TestColumnarOIDMarshalJSONRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)

			// Кодируем
			data, err := col.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}

			// Декодируем как OID
			var oid OID
			if err := oid.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}

			// Парсим колумнарный
			parsed, err := ParseColumnarOID(base, oid)
			if err != nil {
				t.Fatalf("ParseColumnarOID: %v", err)
			}

			if !parsed.Equal(col) {
				t.Errorf("Round trip: %v -> %s -> %v", col, data, parsed)
			}
		})
	}
}

func TestColumnarOIDMarshalJSONCompareWithFullOID(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	colData, err := col.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	fullData, err := col.FullOID().MarshalJSON()
	if err != nil {
		t.Fatalf("FullOID.MarshalJSON: %v", err)
	}

	if !bytes.Equal(colData, fullData) {
		t.Errorf("MarshalJSON = %s, want %s", colData, fullData)
	}
}

// Пример использования
func ExampleColumnarOID_MarshalJSON() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	data, err := col.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output: "1.3.6.1.2.1.2.2.1.2.1"
}

// Бенчмарк
func BenchmarkColumnarOIDMarshalJSON(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = col.MarshalJSON()
	}
}

func TestColumnarOIDUnmarshalJSON(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name     string
		data     []byte
		setBase  bool
		expected ColumnarOID
		wantErr  error
	}{
		{
			name:    "С установленной базой",
			data:    []byte(`"1.3.6.1.2.1.2.2.1.2.1"`),
			setBase: true,
			expected: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			wantErr: nil,
		},
		{
			name:    "Без установленной базы",
			data:    []byte(`"1.3.6.1"`),
			setBase: false,
			expected: ColumnarOID{
				Base:    MustParseOID("1.3.6.1"),
				Column:  0,
				Indexes: nil,
			},
			wantErr: nil,
		},
		{
			name:     "null",
			data:     []byte(`null`),
			setBase:  false,
			expected: ColumnarOID{},
			wantErr:  ErrInvalidJSONType,
		},
		{
			name:     "Невалидный JSON",
			data:     []byte(`invalid`),
			setBase:  false,
			expected: ColumnarOID{},
			wantErr:  ErrInvalidJSONType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var col ColumnarOID

			if tt.setBase {
				col.Base = base
			}

			err := col.UnmarshalJSON(tt.data)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UnmarshalJSON: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UnmarshalJSON = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalJSON: %v", err)
				return
			}

			if !col.Equal(tt.expected) {
				t.Errorf("UnmarshalJSON = %v, want %v", col, tt.expected)
			}
		})
	}
}

func TestColumnarOIDUnmarshalJSONRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)

			// Кодируем
			data, err := col.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}

			// Декодируем с установленной базой
			var decoded ColumnarOID
			decoded.Base = base
			if err := decoded.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}

			if !decoded.Equal(col) {
				t.Errorf("Round trip: %v -> %s -> %v", col, data, decoded)
			}
		})
	}
}

func TestColumnarOIDUnmarshalJSONNotInBase(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	var col ColumnarOID
	col.Base = base

	// OID не принадлежит базе
	data := []byte(`"1.3.6.1.2.1.1.1.0"`)

	err := col.UnmarshalJSON(data)

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrOIDNotInBase) {
		t.Errorf("UnmarshalJSON = %v, want ErrOIDNotInBase", err)
	}
}

// Пример использования
func ExampleColumnarOID_UnmarshalJSON() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	data := []byte(`"1.3.6.1.2.1.2.2.1.2.1"`)

	var col ColumnarOID
	col.Base = base
	if err := col.UnmarshalJSON(data); err != nil {
		panic(err)
	}

	fmt.Println(col.Column)
	fmt.Println(col.Indexes)
	// Output:
	// 2
	// [1]
}

// Бенчмарк
func BenchmarkColumnarOIDUnmarshalJSON(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	data := []byte(`"1.3.6.1.2.1.2.2.1.2.1"`)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var col ColumnarOID
		col.Base = base
		_ = col.UnmarshalJSON(data)
	}
}

func TestColumnarOIDValue(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name     string
		col      ColumnarOID
		expected driver.Value
		wantErr  error
	}{
		{
			name: "Без индексов",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: nil,
			},
			expected: "1.3.6.1.2.1.2.2.1.2",
			wantErr:  nil,
		},
		{
			name: "Один индекс",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			expected: "1.3.6.1.2.1.2.2.1.2.1",
			wantErr:  nil,
		},
		{
			name: "Два индекса",
			col: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1, 2},
			},
			expected: "1.3.6.1.2.1.2.2.1.2.1.2",
			wantErr:  nil,
		},
		{
			name: "Пустая база",
			col: ColumnarOID{
				Base:   OID{},
				Column: 0,
			},
			expected: nil,
			wantErr:  ErrOIDTooShort,
		},
		{
			name:     "Пустая колонка",
			col:      ColumnarOID{},
			expected: nil,
			wantErr:  ErrOIDTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.col.Value()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Value: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Value = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Value: %v", err)
				return
			}

			if value != tt.expected {
				t.Errorf("Value = %v, want %v", value, tt.expected)
			}
		})
	}
}

func TestColumnarOIDValueTypes(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	value, err := col.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	if _, ok := value.(string); !ok {
		t.Errorf("Value тип = %T, want string", value)
	}
}

func TestColumnarOIDValueImplementsValuer(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	var _ driver.Valuer = NewColumnarOID(base, 2, 1)
}

func TestColumnarOIDValueRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)

			value, err := col.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			// Сканируем как OID
			var oid OID
			if err := oid.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			// Парсим колумнарный
			parsed, err := ParseColumnarOID(base, oid)
			if err != nil {
				t.Fatalf("ParseColumnarOID: %v", err)
			}

			if !parsed.Equal(col) {
				t.Errorf("Round trip: %v -> %v -> %v", col, value, parsed)
			}
		})
	}
}

// Пример использования
func ExampleColumnarOID_Value() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	value, err := col.Value()
	if err != nil {
		panic(err)
	}

	fmt.Println(value)
	// Output: 1.3.6.1.2.1.2.2.1.2.1
}

// Бенчмарк
func BenchmarkColumnarOIDValue(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	col := NewColumnarOID(base, 2, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = col.Value()
	}
}

func TestColumnarOIDScan(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		name     string
		input    any
		setBase  bool
		expected ColumnarOID
		wantErr  error
	}{
		{
			name:    "Строка с базой",
			input:   "1.3.6.1.2.1.2.2.1.2.1",
			setBase: true,
			expected: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			wantErr: nil,
		},
		{
			name:    "Байты с базой",
			input:   []byte("1.3.6.1.2.1.2.2.1.2.1"),
			setBase: true,
			expected: ColumnarOID{
				Base:    base,
				Column:  2,
				Indexes: []uint32{1},
			},
			wantErr: nil,
		},
		{
			name:    "Строка без базы",
			input:   "1.3.6.1",
			setBase: false,
			expected: ColumnarOID{
				Base:    MustParseOID("1.3.6.1"),
				Column:  0,
				Indexes: nil,
			},
			wantErr: nil,
		},
		{
			name:     "NULL",
			input:    nil,
			setBase:  false,
			expected: ColumnarOID{},
			wantErr:  nil,
		},
		{
			name:     "Число",
			input:    123,
			setBase:  false,
			expected: ColumnarOID{},
			wantErr:  ErrUnsupportedScanType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var col ColumnarOID

			if tt.setBase {
				col.Base = base
			}

			err := col.Scan(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Scan: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Scan = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if tt.input == nil {
				// NULL
				if err != nil {
					t.Errorf("Scan(nil): %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("Scan: %v", err)
				return
			}

			if !col.Equal(tt.expected) {
				t.Errorf("Scan = %v, want %v", col, tt.expected)
			}
		})
	}
}

func TestColumnarOIDScanRoundTrip(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	tests := []struct {
		column  uint32
		indexes []uint32
	}{
		{column: 2, indexes: nil},
		{column: 2, indexes: []uint32{1}},
		{column: 2, indexes: []uint32{1, 2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("col=%d,idx=%v", tt.column, tt.indexes), func(t *testing.T) {
			col := NewColumnarOID(base, tt.column, tt.indexes...)

			value, err := col.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			var scanned ColumnarOID
			scanned.Base = base
			if err := scanned.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			if !scanned.Equal(col) {
				t.Errorf("Round trip: %v -> %v -> %v", col, value, scanned)
			}
		})
	}
}

func TestColumnarOIDScanImplementsScanner(t *testing.T) {
	var _ sql.Scanner = (*ColumnarOID)(nil)
}

func TestColumnarOIDScanNotInBase(t *testing.T) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	var col ColumnarOID
	col.Base = base

	err := col.Scan("1.3.6.1.2.1.1.1.0")

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrOIDNotInBase) {
		t.Errorf("Scan = %v, want ErrOIDNotInBase", err)
	}
}

// Пример использования
func ExampleColumnarOID_Scan() {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")

	var col ColumnarOID
	col.Base = base

	if err := col.Scan("1.3.6.1.2.1.2.2.1.2.1"); err != nil {
		panic(err)
	}

	fmt.Println(col.Column)
	fmt.Println(col.Indexes)
	// Output:
	// 2
	// [1]
}

// Бенчмарк
func BenchmarkColumnarOIDScan(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.2.2.1")
	oidStr := "1.3.6.1.2.1.2.2.1.2.1"

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var col ColumnarOID
		col.Base = base
		_ = col.Scan(oidStr)
	}
}
