// oid/scalar_test.go
package oid

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestNewScalarOID(t *testing.T) {
	tests := []struct {
		name     string
		input    OID
		expected ScalarOID
	}{
		{
			name:     "Уже скалярный",
			input:    OID{1, 3, 6, 1, 0},
			expected: ScalarOID{1, 3, 6, 1, 0},
		},
		{
			name:     "Без .0",
			input:    OID{1, 3, 6, 1},
			expected: ScalarOID{1, 3, 6, 1, 0},
		},
		{
			name:     "Пустой",
			input:    OID{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewScalarOID(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("NewScalarOID(%v) = %v, ожидалось %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseScalarOID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ScalarOID
		wantErr  error
	}{
		{
			name:     "Валидный скалярный OID",
			input:    "1.3.6.1.2.1.1.1.0",
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "OID без .0 (добавляется автоматически)",
			input:    "1.3.6.1.2.1.1.1",
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Короткий OID",
			input:    "1.3.6.1",
			expected: ScalarOID{1, 3, 6, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Пустая строка",
			input:    "",
			expected: nil,
			wantErr:  ErrEmptyOID,
		},
		{
			name:     "Невалидный OID",
			input:    "invalid",
			expected: nil,
			wantErr:  nil, // Любая ошибка
		},
		{
			name:     "Один компонент",
			input:    "1",
			expected: nil,
			wantErr:  ErrOIDTooShort,
		},
		{
			name:     "Первый компонент > 2",
			input:    "3.1",
			expected: nil,
			wantErr:  ErrFirstComponentTooBig,
		},
		{
			name:     "Второй компонент > 39",
			input:    "1.40",
			expected: nil,
			wantErr:  ErrSecondComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseScalarOID(tt.input)

			// Проверяем ошибку
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseScalarOID(%q): ожидалась ошибка %v, но её нет",
						tt.input, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseScalarOID(%q) = %v, ожидалось %v",
						tt.input, err, tt.wantErr)
				}
				return
			}

			// Если wantErr == nil, но ошибка может быть любой
			if tt.wantErr == nil && tt.expected == nil {
				if err == nil {
					t.Errorf("ParseScalarOID(%q): ожидалась ошибка, но её нет", tt.input)
				}
				return
			}

			// Проверяем успешный результат
			if err != nil {
				t.Errorf("ParseScalarOID(%q): неожиданная ошибка: %v", tt.input, err)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("ParseScalarOID(%q) = %v, ожидалось %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseScalarOIDTableDriven(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ScalarOID
		wantErr bool
	}{
		{
			name:    "Валидный с .0",
			input:   "1.3.6.1.2.1.1.1.0",
			want:    ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr: false,
		},
		{
			name:    "Валидный без .0",
			input:   "1.3.6.1.2.1.1.1",
			want:    ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr: false,
		},
		{
			name:    "Пустая строка",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Невалидный",
			input:   "invalid",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseScalarOID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseScalarOID(%q) error = %v, wantErr %v",
					tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("ParseScalarOID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseScalarOIDSubtests(t *testing.T) {
	t.Run("Валидные", func(t *testing.T) {
		tests := []struct {
			input string
			want  ScalarOID
		}{
			{"1.3.6.1.0", ScalarOID{1, 3, 6, 1, 0}},
			{"1.3.6.1.2.1.1.1.0", ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}},
			{"2.100.3.0", ScalarOID{2, 100, 3, 0}},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				got, err := ParseScalarOID(tt.input)
				if err != nil {
					t.Fatalf("ParseScalarOID(%q): %v", tt.input, err)
				}
				if !got.Equal(tt.want) {
					t.Errorf("ParseScalarOID(%q) = %v, want %v", tt.input, got, tt.want)
				}
			})
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []struct {
			input   string
			wantErr error
		}{
			{"", ErrEmptyOID},
			{"1", ErrOIDTooShort},
			{"3.1", ErrFirstComponentTooBig},
			{"1.40", ErrSecondComponentTooBig},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				_, err := ParseScalarOID(tt.input)
				if err == nil {
					t.Fatalf("ParseScalarOID(%q): ожидалась ошибка", tt.input)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseScalarOID(%q) = %v, want %v", tt.input, err, tt.wantErr)
				}
			})
		}
	})
}

func TestParseScalarOIDProperties(t *testing.T) {
	// Проверяем, что результат всегда скалярный
	t.Run("Всегда скалярный", func(t *testing.T) {
		inputs := []string{
			"1.3.6.1",
			"1.3.6.1.2.1.1.1",
			"2.100.3",
			"0.39.1",
		}

		for _, input := range inputs {
			scalar, err := ParseScalarOID(input)
			if err != nil {
				t.Errorf("ParseScalarOID(%q): %v", input, err)
				continue
			}
			if !scalar.IsScalar() {
				t.Errorf("ParseScalarOID(%q): результат не скалярный", input)
			}
		}
	})

	// Проверяем, что .0 добавляется только если его нет
	t.Run("Не дублирует .0", func(t *testing.T) {
		withZero := "1.3.6.1.0"
		scalar1, _ := ParseScalarOID(withZero)

		withoutZero := "1.3.6.1"
		scalar2, _ := ParseScalarOID(withoutZero)

		if !scalar1.Equal(scalar2) {
			t.Errorf("%v != %v", scalar1, scalar2)
		}
	})
}

func FuzzParseScalarOID(f *testing.F) {
	// Добавляем начальные значения
	f.Add("1.3.6.1.0")
	f.Add("1.3.6.1.2.1.1.1.0")
	f.Add("")
	f.Add("invalid")

	f.Fuzz(func(t *testing.T, input string) {
		scalar, err := ParseScalarOID(input)

		if err != nil {
			// Ошибка - это нормально для невалидного ввода
			return
		}

		// Проверяем, что результат валидный
		if err := scalar.Validate(); err != nil {
			t.Errorf("ParseScalarOID(%q) вернул невалидный результат: %v", input, err)
		}

		// Проверяем, что результат скалярный
		if !scalar.IsScalar() {
			t.Errorf("ParseScalarOID(%q): результат не скалярный", input)
		}
	})
}

func TestScalarOIDMethods(t *testing.T) {
	// Создаем скалярный OID
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	// IsScalar
	if !scalar.IsScalar() {
		t.Error("IsScalar: должно быть true")
	}

	// Base
	base := scalar.Base()
	if !base.Equal(MustParseOID("1.3.6.1.2.1.1.1")) {
		t.Errorf("Base = %v, ожидалось 1.3.6.1.2.1.1.1", base)
	}

	// String
	if scalar.String() != "1.3.6.1.2.1.1.1.0" {
		t.Errorf("String = %s", scalar.String())
	}

	// Validate
	if err := scalar.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}

	// OID
	if !scalar.OID().Equal(MustParseOID("1.3.6.1.2.1.1.1.0")) {
		t.Error("OID: неверный результат")
	}
}

func TestScalarOIDJSON(t *testing.T) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	// Marshal
	data, err := json.Marshal(scalar)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(data) != `"1.3.6.1.2.1.1.1.0"` {
		t.Errorf("MarshalJSON = %s", data)
	}

	// Unmarshal
	var decoded ScalarOID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if !decoded.Equal(scalar) {
		t.Error("UnmarshalJSON: не совпадает")
	}
}

func TestScalarOIDBER(t *testing.T) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	// Marshal
	data, err := scalar.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	// Unmarshal
	var decoded ScalarOID
	if err := decoded.UnmarshalBER(data); err != nil {
		t.Fatalf("UnmarshalBER: %v", err)
	}
	if !decoded.Equal(scalar) {
		t.Error("UnmarshalBER: не совпадает")
	}
}

func BenchmarkParseScalarOID(b *testing.B) {
	inputs := []string{
		"1.3.6.1.0",
		"1.3.6.1.2.1.1.1.0",
		"1.3.6.1.2.1.2.2.1.2.1",
	}

	for _, input := range inputs {
		b.Run(input, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _ = ParseScalarOID(input)
			}
		})
	}
}

func ExampleParseScalarOID() {
	scalar, err := ParseScalarOID("1.3.6.1.2.1.1.1")
	if err != nil {
		panic(err)
	}
	fmt.Println(scalar)
	// Output: 1.3.6.1.2.1.1.1.0
}

func ExampleParseScalarOID_error() {
	_, err := ParseScalarOID("invalid")
	fmt.Println(err != nil)
	// Output: true
}
