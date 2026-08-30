// oid/scalar_new_test.go
package oid

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/asn1"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestScalarOIDType(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		expected OID
	}{
		{
			name:     "Скалярный OID",
			scalar:   ScalarOID{1, 3, 6, 1, 0},
			expected: OID{1, 3, 6, 1, 0},
		},
		{
			name:     "Пустой",
			scalar:   ScalarOID{},
			expected: OID{},
		},
		{
			name:     "Nil",
			scalar:   nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ScalarOID - это OID (type alias)
			oid := OID(tt.scalar)

			if !oid.Equal(tt.expected) {
				t.Errorf("OID(scalar) = %v, want %v", oid, tt.expected)
			}
		})
	}
}

func TestScalarOIDConversion(t *testing.T) {
	t.Run("OID -> ScalarOID", func(t *testing.T) {
		oid := OID{1, 3, 6, 1, 0}
		scalar := ScalarOID(oid)

		if len(scalar) != len(oid) {
			t.Errorf("len = %d, want %d", len(scalar), len(oid))
		}
	})

	t.Run("ScalarOID -> OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}
		oid := OID(scalar)

		if len(oid) != len(scalar) {
			t.Errorf("len = %d, want %d", len(oid), len(scalar))
		}
	})
}

func TestScalarOIDUnderlyingType(t *testing.T) {
	// ScalarOID - это []uint32 (как OID)
	var scalar ScalarOID = ScalarOID{1, 3, 6}

	// Можно использовать как слайс
	if scalar[0] != 1 {
		t.Error("scalar[0] должен быть 1")
	}

	if len(scalar) != 3 {
		t.Errorf("len = %d, want 3", len(scalar))
	}
}

// Пример использования
func ExampleScalarOID() {
	scalar := ScalarOID{1, 3, 6, 1, 0}

	fmt.Println(scalar)
	// Output: 1.3.6.1.0
}

// Бенчмарк
func BenchmarkScalarOIDCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = ScalarOID{1, 3, 6, 1, 0}
	}
}

func TestNewScalarOID(t *testing.T) {
	tests := []struct {
		name     string
		input    OID
		expected ScalarOID
	}{
		{
			name:     "Уже скалярный (заканчивается на 0)",
			input:    OID{1, 3, 6, 1, 0},
			expected: ScalarOID{1, 3, 6, 1, 0},
		},
		{
			name:     "Без .0 (добавляется)",
			input:    OID{1, 3, 6, 1},
			expected: ScalarOID{1, 3, 6, 1, 0},
		},
		{
			name:     "Длинный без .0",
			input:    OID{1, 3, 6, 1, 2, 1, 1, 1},
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
		},
		{
			name:     "Пустой OID",
			input:    OID{},
			expected: nil,
		},
		{
			name:     "Nil OID",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Один компонент",
			input:    OID{1},
			expected: ScalarOID{1, 0},
		},
		{
			name:     "Два компонента",
			input:    OID{1, 3},
			expected: ScalarOID{1, 3, 0},
		},
		{
			name:     "С первым 2",
			input:    OID{2, 100, 3},
			expected: ScalarOID{2, 100, 3, 0},
		},
		{
			name:     "Заканчивается на 0 (не добавляется)",
			input:    OID{1, 3, 6, 0},
			expected: ScalarOID{1, 3, 6, 0},
		},
		{
			name:     "Несколько нулей в конце",
			input:    OID{1, 3, 0, 0},
			expected: ScalarOID{1, 3, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewScalarOID(tt.input)

			if !result.Equal(tt.expected) {
				t.Errorf("NewScalarOID(%v) = %v, ожидалось %v",
					tt.input, result, tt.expected)
			}

			// Проверяем длину
			if len(result) != len(tt.expected) {
				t.Errorf("len = %d, ожидалось %d", len(result), len(tt.expected))
			}
		})
	}
}

// Тест с проверкой свойств
func TestNewScalarOIDProperties(t *testing.T) {
	t.Run("Всегда возвращает скалярный OID", func(t *testing.T) {
		inputs := []OID{
			{1, 3, 6, 1},
			{1, 3, 6, 1, 0},
			{2, 100, 3},
			{0, 39, 1},
		}

		for _, input := range inputs {
			result := NewScalarOID(input)

			if len(result) > 0 && !result.IsScalar() {
				t.Errorf("NewScalarOID(%v) не скалярный: %v", input, result)
			}
		}
	})

	t.Run("Не изменяет входной OID", func(t *testing.T) {
		input := OID{1, 3, 6, 1}
		inputCopy := make(OID, len(input))
		copy(inputCopy, input)

		NewScalarOID(input)

		if !input.Equal(inputCopy) {
			t.Error("NewScalarOID не должен изменять входной OID")
		}
	})

	t.Run("Не дублирует .0", func(t *testing.T) {
		// Уже заканчивается на 0
		withZero := OID{1, 3, 6, 0}
		result1 := NewScalarOID(withZero)

		// Без .0
		withoutZero := OID{1, 3, 6}
		result2 := NewScalarOID(withoutZero)

		// Оба должны быть одинаковыми
		if !result1.Equal(result2) {
			t.Errorf("%v != %v", result1, result2)
		}
	})
}

// Тест с подтестами
func TestNewScalarOIDCategories(t *testing.T) {
	t.Run("OID без .0", func(t *testing.T) {
		tests := []struct {
			input    OID
			expected ScalarOID
		}{
			{OID{1, 3}, ScalarOID{1, 3, 0}},
			{OID{1, 3, 6}, ScalarOID{1, 3, 6, 0}},
			{OID{1, 3, 6, 1}, ScalarOID{1, 3, 6, 1, 0}},
		}

		for _, tt := range tests {
			result := NewScalarOID(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("NewScalarOID(%v) = %v, ожидалось %v",
					tt.input, result, tt.expected)
			}
		}
	})

	t.Run("OID с .0", func(t *testing.T) {
		tests := []struct {
			input    OID
			expected ScalarOID
		}{
			{OID{1, 3, 0}, ScalarOID{1, 3, 0}},
			{OID{1, 3, 6, 0}, ScalarOID{1, 3, 6, 0}},
		}

		for _, tt := range tests {
			result := NewScalarOID(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("NewScalarOID(%v) = %v, ожидалось %v",
					tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Пустые OID", func(t *testing.T) {
		tests := []OID{
			{},
			nil,
		}

		for _, input := range tests {
			result := NewScalarOID(input)
			if result != nil {
				t.Errorf("NewScalarOID(%v) = %v, ожидался nil", input, result)
			}
		}
	})
}

// Тест с round trip
func TestNewScalarOIDRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			scalar := NewScalarOID(oid)

			// Проверяем, что скалярный
			if !scalar.IsScalar() {
				t.Error("Должен быть скалярным")
			}

			// Получаем base
			base := scalar.Base()

			// Base должен совпадать с оригиналом (если не заканчивался на 0)
			if len(oid) > 0 && oid[len(oid)-1] != 0 {
				if !base.Equal(oid) {
					t.Errorf("Base = %v, ожидалось %v", base, oid)
				}
			}
		})
	}
}

// Пример использования
func ExampleNewScalarOID() {
	// OID без .0
	oid1 := OID{1, 3, 6, 1}
	scalar1 := NewScalarOID(oid1)
	fmt.Println(scalar1)

	// OID с .0
	oid2 := OID{1, 3, 6, 1, 0}
	scalar2 := NewScalarOID(oid2)
	fmt.Println(scalar2)

	// Output:
	// 1.3.6.1.0
	// 1.3.6.1.0
}

// Пример с пустым OID
func ExampleNewScalarOID_empty() {
	scalar := NewScalarOID(OID{})
	fmt.Println(scalar == nil)
	// Output: true
}

// Бенчмарк
func BenchmarkNewScalarOID(b *testing.B) {
	base := MustParseOID("1.3.6.1.2.1.1.1")

	b.ReportAllocs()
	for b.Loop() {
		_ = NewScalarOID(base)
	}
}

func TestMustScalarOID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ScalarOID
	}{
		{
			name:     "Уже скалярный",
			input:    "1.3.6.1.2.1.1.1.0",
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
		},
		{
			name:     "Без .0",
			input:    "1.3.6.1.2.1.1.1",
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
		},
		{
			name:     "Короткий",
			input:    "1.3.6.1",
			expected: ScalarOID{1, 3, 6, 1, 0},
		},
		{
			name:     "С первым 2",
			input:    "2.100.3",
			expected: ScalarOID{2, 100, 3, 0},
		},
		{
			name:     "С первым 0",
			input:    "0.39.1",
			expected: ScalarOID{0, 39, 1, 0},
		},
		{
			name:     "С большим компонентом",
			input:    "1.3.268435455",
			expected: ScalarOID{1, 3, MaxOIDComponent, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MustScalarOID(tt.input)

			if !result.Equal(tt.expected) {
				t.Errorf("MustScalarOID(%q) = %v, ожидалось %v",
					tt.input, result, tt.expected)
			}

			// Проверяем, что результат скалярный
			if !result.IsScalar() {
				t.Errorf("MustScalarOID(%q): результат не скалярный", tt.input)
			}
		})
	}
}

// Тест с паникой
func TestMustScalarOIDPanic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Пустая строка", ""},
		{"Невалидный", "invalid"},
		{"Один компонент", "1"},
		{"Первый > 2", "3.1"},
		{"Второй > 39", "1.40"},
		{"Спецсимволы", "1.3#"},
		{"Отрицательное", "-1.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("MustScalarOID(%q): ожидалась паника", tt.input)
				}
			}()

			MustScalarOID(tt.input)
		})
	}
}

// Тест с проверкой свойств
func TestMustScalarOIDProperties(t *testing.T) {
	t.Run("Всегда возвращает скалярный", func(t *testing.T) {
		inputs := []string{
			"1.3.6.1",
			"1.3.6.1.2.1.1.1",
			"2.100.3",
			"0.39.1",
		}

		for _, input := range inputs {
			result := MustScalarOID(input)

			if !result.IsScalar() {
				t.Errorf("MustScalarOID(%q) не скалярный: %v", input, result)
			}
		}
	})

	t.Run("Эквивалентна NewScalarOID + MustParseOID", func(t *testing.T) {
		input := "1.3.6.1.2.1.1.1"

		result1 := MustScalarOID(input)
		result2 := NewScalarOID(MustParseOID(input))

		if !result1.Equal(result2) {
			t.Error("MustScalarOID и NewScalarOID должны давать одинаковый результат")
		}
	})

	t.Run("Не дублирует .0", func(t *testing.T) {
		withZero := MustScalarOID("1.3.6.1.0")
		withoutZero := MustScalarOID("1.3.6.1")

		if !withZero.Equal(withoutZero) {
			t.Errorf("%v != %v", withZero, withoutZero)
		}
	})
}

// Тест с подтестами
func TestMustScalarOIDCategories(t *testing.T) {
	t.Run("Валидные строки", func(t *testing.T) {
		tests := []struct {
			input    string
			expected ScalarOID
		}{
			{"1.3.6.1", ScalarOID{1, 3, 6, 1, 0}},
			{"1.3.6.1.0", ScalarOID{1, 3, 6, 1, 0}},
			{"2.100.3", ScalarOID{2, 100, 3, 0}},
		}

		for _, tt := range tests {
			result := MustScalarOID(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("MustScalarOID(%q) = %v, ожидалось %v",
					tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Паника при невалидных", func(t *testing.T) {
		inputs := []string{"", "invalid", "1", "3.1", "1.40"}

		for _, input := range inputs {
			func() {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MustScalarOID(%q): ожидалась паника", input)
					}
				}()
				MustScalarOID(input)
			}()
		}
	})
}

// Тест с round trip
func TestMustScalarOIDRoundTrip(t *testing.T) {
	inputs := []string{
		"1.3.6.1",
		"1.3.6.1.2.1.1.1",
		"2.100.3",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			scalar := MustScalarOID(input)

			// Проверяем, что скалярный
			if !scalar.IsScalar() {
				t.Error("Должен быть скалярным")
			}

			// Получаем base
			base := scalar.Base()

			// Base.String() должен совпадать с входом
			if base.String() != input {
				t.Errorf("Base.String() = %s, ожидалось %s", base.String(), input)
			}
		})
	}
}

// Пример использования
func ExampleMustScalarOID() {
	// OID без .0
	scalar1 := MustScalarOID("1.3.6.1.2.1.1.1")
	fmt.Println(scalar1)

	// OID с .0
	scalar2 := MustScalarOID("1.3.6.1.2.1.1.1.0")
	fmt.Println(scalar2)

	// Output:
	// 1.3.6.1.2.1.1.1.0
	// 1.3.6.1.2.1.1.1.0
}

// Пример с паникой
func ExampleMustScalarOID_panic() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Паника поймана")
		}
	}()

	MustScalarOID("invalid")
	// Output: Паника поймана
}

// Бенчмарк
func BenchmarkMustScalarOID(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = MustScalarOID("1.3.6.1.2.1.1.1")
	}
}

// Сравнение с NewScalarOID
func BenchmarkMustScalarOIDVsNewScalarOID(b *testing.B) {
	oid := MustParseOID("1.3.6.1.2.1.1.1")

	b.Run("MustScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = MustScalarOID("1.3.6.1.2.1.1.1")
		}
	})

	b.Run("NewScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = NewScalarOID(oid)
		}
	})
}

func TestParseScalarOID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ScalarOID
		wantErr  error
	}{
		{
			name:     "Уже скалярный",
			input:    "1.3.6.1.2.1.1.1.0",
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Без .0",
			input:    "1.3.6.1.2.1.1.1",
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Короткий",
			input:    "1.3.6.1",
			expected: ScalarOID{1, 3, 6, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			input:    "2.100.3",
			expected: ScalarOID{2, 100, 3, 0},
			wantErr:  nil,
		},
		{
			name:     "С первым 0",
			input:    "0.39.1",
			expected: ScalarOID{0, 39, 1, 0},
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
			wantErr:  nil, // Любая ошибка парсинга
		},
		{
			name:     "Один компонент",
			input:    "1",
			expected: nil,
			wantErr:  ErrOIDTooShort,
		},
		{
			name:     "Первый > 2",
			input:    "3.1",
			expected: nil,
			wantErr:  ErrFirstComponentTooBig,
		},
		{
			name:     "Второй > 39",
			input:    "1.40",
			expected: nil,
			wantErr:  ErrSecondComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseScalarOID(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseScalarOID(%q): ожидалась ошибка %v", tt.input, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseScalarOID(%q) = %v, ожидалось %v",
						tt.input, err, tt.wantErr)
				}
				return
			}

			if tt.expected == nil {
				// Ожидаем любую ошибку
				if err == nil {
					t.Errorf("ParseScalarOID(%q): ожидалась ошибка", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseScalarOID(%q): %v", tt.input, err)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("ParseScalarOID(%q) = %v, ожидалось %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestParseScalarOIDProperties(t *testing.T) {
	t.Run("Всегда возвращает скалярный", func(t *testing.T) {
		inputs := []string{
			"1.3.6.1",
			"1.3.6.1.2.1.1.1",
			"2.100.3",
			"0.39.1",
		}

		for _, input := range inputs {
			result, err := ParseScalarOID(input)
			if err != nil {
				t.Errorf("ParseScalarOID(%q): %v", input, err)
				continue
			}

			if !result.IsScalar() {
				t.Errorf("ParseScalarOID(%q) не скалярный: %v", input, result)
			}
		}
	})

	t.Run("Эквивалентна NewScalarOID + ParseOID", func(t *testing.T) {
		input := "1.3.6.1.2.1.1.1"

		result1, err1 := ParseScalarOID(input)
		if err1 != nil {
			t.Fatalf("ParseScalarOID: %v", err1)
		}

		oid, err2 := ParseOID(input)
		if err2 != nil {
			t.Fatalf("ParseOID: %v", err2)
		}
		result2 := NewScalarOID(oid)

		if !result1.Equal(result2) {
			t.Error("ParseScalarOID и NewScalarOID должны давать одинаковый результат")
		}
	})

	t.Run("Не дублирует .0", func(t *testing.T) {
		withZero, _ := ParseScalarOID("1.3.6.1.0")
		withoutZero, _ := ParseScalarOID("1.3.6.1")

		if !withZero.Equal(withoutZero) {
			t.Errorf("%v != %v", withZero, withoutZero)
		}
	})
}

// Тест с подтестами
func TestParseScalarOIDCategories(t *testing.T) {
	t.Run("Валидные строки", func(t *testing.T) {
		tests := []struct {
			input    string
			expected ScalarOID
		}{
			{"1.3.6.1", ScalarOID{1, 3, 6, 1, 0}},
			{"1.3.6.1.0", ScalarOID{1, 3, 6, 1, 0}},
			{"2.100.3", ScalarOID{2, 100, 3, 0}},
			{"0.39.1", ScalarOID{0, 39, 1, 0}},
		}

		for _, tt := range tests {
			result, err := ParseScalarOID(tt.input)
			if err != nil {
				t.Errorf("ParseScalarOID(%q): %v", tt.input, err)
				continue
			}
			if !result.Equal(tt.expected) {
				t.Errorf("ParseScalarOID(%q) = %v, ожидалось %v",
					tt.input, result, tt.expected)
			}
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
			_, err := ParseScalarOID(tt.input)
			if err == nil {
				t.Errorf("ParseScalarOID(%q): ожидалась ошибка", tt.input)
				continue
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseScalarOID(%q) = %v, ожидалось %v",
					tt.input, err, tt.wantErr)
			}
		}
	})
}

// Тест с round trip
func TestParseScalarOIDRoundTrip(t *testing.T) {
	inputs := []string{
		"1.3.6.1",
		"1.3.6.1.2.1.1.1",
		"2.100.3",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			scalar, err := ParseScalarOID(input)
			if err != nil {
				t.Fatalf("ParseScalarOID: %v", err)
			}

			// Проверяем, что скалярный
			if !scalar.IsScalar() {
				t.Error("Должен быть скалярным")
			}

			// Получаем base
			base := scalar.Base()

			// Base.String() должен совпадать с входом
			if base.String() != input {
				t.Errorf("Base.String() = %s, ожидалось %s", base.String(), input)
			}
		})
	}
}

// Пример использования
func ExampleParseScalarOID() {
	// OID без .0
	scalar1, _ := ParseScalarOID("1.3.6.1.2.1.1.1")
	fmt.Println(scalar1)

	// OID с .0
	scalar2, _ := ParseScalarOID("1.3.6.1.2.1.1.1.0")
	fmt.Println(scalar2)

	// Output:
	// 1.3.6.1.2.1.1.1.0
	// 1.3.6.1.2.1.1.1.0
}

// Пример с ошибкой
func ExampleParseScalarOID_error() {
	_, err := ParseScalarOID("invalid")
	fmt.Println(err != nil)
	// Output: true
}

// Бенчмарк
func BenchmarkParseScalarOID(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = ParseScalarOID("1.3.6.1.2.1.1.1")
	}
}

// Сравнение с MustScalarOID
func BenchmarkParseScalarOIDVsMustScalarOID(b *testing.B) {
	b.Run("ParseScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = ParseScalarOID("1.3.6.1.2.1.1.1")
		}
	})

	b.Run("MustScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = MustScalarOID("1.3.6.1.2.1.1.1")
		}
	})
}

func TestScalarOIDOID(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		expected OID
	}{
		{
			name:     "Скалярный OID",
			scalar:   ScalarOID{1, 3, 6, 1, 0},
			expected: OID{1, 3, 6, 1, 0},
		},
		{
			name:     "Длинный скалярный",
			scalar:   ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
		},
		{
			name:     "Не скалярный",
			scalar:   ScalarOID{1, 3, 6, 1},
			expected: OID{1, 3, 6, 1},
		},
		{
			name:     "Пустой",
			scalar:   ScalarOID{},
			expected: OID{},
		},
		{
			name:     "Nil",
			scalar:   nil,
			expected: nil,
		},
		{
			name:     "Один компонент",
			scalar:   ScalarOID{1},
			expected: OID{1},
		},
		{
			name:     "С первым 2",
			scalar:   ScalarOID{2, 100, 3, 0},
			expected: OID{2, 100, 3, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scalar.OID()

			if !result.Equal(tt.expected) {
				t.Errorf("OID() = %v, ожидалось %v", result, tt.expected)
			}

			// Проверяем длину
			if len(result) != len(tt.expected) {
				t.Errorf("len = %d, ожидалось %d", len(result), len(tt.expected))
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDOIDProperties(t *testing.T) {
	t.Run("Возвращает тот же OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		result := scalar.OID()

		if !result.Equal(OID(scalar)) {
			t.Error("OID() должен возвращать тот же OID")
		}
	})

	t.Run("Не создает копию", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		result := scalar.OID()

		// Изменяем результат
		result[0] = 99

		// Scalar должен измениться (так как это тот же слайс)
		if scalar[0] != 99 {
			t.Error("OID() должен возвращать тот же слайс (не копию)")
		}

		// Восстанавливаем
		scalar[0] = 1
	})

	t.Run("Можно использовать как OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		oid := scalar.OID()

		// Используем методы OID
		if oid.String() != "1.3.6.1.0" {
			t.Errorf("String() = %s", oid.String())
		}

		if err := oid.Validate(); err != nil {
			t.Errorf("Validate: %v", err)
		}

		if !oid.StartsWith(OID{1, 3}) {
			t.Error("StartsWith не работает")
		}
	})
}

// Тест с round trip
func TestScalarOIDOIDRoundTrip(t *testing.T) {
	scalars := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
	}

	for _, scalar := range scalars {
		t.Run(scalar.String(), func(t *testing.T) {
			// ScalarOID -> OID
			oid := scalar.OID()

			// OID -> ScalarOID (обратно)
			backToScalar := ScalarOID(oid)

			if !backToScalar.Equal(scalar) {
				t.Errorf("Round trip: %v -> %v -> %v", scalar, oid, backToScalar)
			}
		})
	}
}

// Тест с подтестами
func TestScalarOIDOIDCategories(t *testing.T) {
	t.Run("Скалярные OID", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
		}

		for _, scalar := range scalars {
			oid := scalar.OID()
			if !oid.Equal(OID(scalar)) {
				t.Error("OID() должен возвращать тот же OID")
			}
		}
	})

	t.Run("Не скалярные OID", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1},
			{2, 100, 3},
		}

		for _, scalar := range scalars {
			oid := scalar.OID()
			if !oid.Equal(OID(scalar)) {
				t.Error("OID() должен возвращать тот же OID")
			}
		}
	})

	t.Run("Пустые OID", func(t *testing.T) {
		scalars := []ScalarOID{
			{},
			nil,
		}

		for _, scalar := range scalars {
			oid := scalar.OID()
			if len(oid) != 0 {
				t.Errorf("OID() = %v, ожидался пустой", oid)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_OID() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	oid := scalar.OID()

	fmt.Println(oid)
	// Output: 1.3.6.1.2.1.1.1.0
}

// Пример с пустым OID
func ExampleScalarOID_OID_empty() {
	scalar := ScalarOID{}

	oid := scalar.OID()

	fmt.Println(len(oid))
	// Output: 0
}

// Бенчмарк
func BenchmarkScalarOIDOID(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_ = scalar.OID()
	}
}

func TestScalarOIDIsScalar(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		expected bool
	}{
		{
			name:     "Заканчивается на 0",
			scalar:   ScalarOID{1, 3, 6, 1, 0},
			expected: true,
		},
		{
			name:     "Длинный заканчивается на 0",
			scalar:   ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: true,
		},
		{
			name:     "Не заканчивается на 0",
			scalar:   ScalarOID{1, 3, 6, 1},
			expected: false,
		},
		{
			name:     "Заканчивается на 1",
			scalar:   ScalarOID{1, 3, 6, 1, 1},
			expected: false,
		},
		{
			name:     "Заканчивается на 39",
			scalar:   ScalarOID{1, 3, 6, 39},
			expected: false,
		},
		{
			name:     "Только 0",
			scalar:   ScalarOID{0},
			expected: true,
		},
		{
			name:     "Два 0",
			scalar:   ScalarOID{0, 0},
			expected: true,
		},
		{
			name:     "0 в середине",
			scalar:   ScalarOID{1, 0, 3},
			expected: false,
		},
		{
			name:     "Пустой",
			scalar:   ScalarOID{},
			expected: false,
		},
		{
			name:     "Nil",
			scalar:   nil,
			expected: false,
		},
		{
			name:     "Один компонент 1",
			scalar:   ScalarOID{1},
			expected: false,
		},
		{
			name:     "Один компонент 0",
			scalar:   ScalarOID{0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scalar.IsScalar()

			if result != tt.expected {
				t.Errorf("IsScalar() = %v, ожидалось %v", result, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDIsScalarProperties(t *testing.T) {
	t.Run("NewScalarOID всегда создает скалярный", func(t *testing.T) {
		oids := []OID{
			{1, 3, 6, 1},
			{1, 3, 6, 1, 2, 1, 1, 1},
			{2, 100, 3},
			{0, 39, 1},
		}

		for _, oid := range oids {
			scalar := NewScalarOID(oid)
			if !scalar.IsScalar() {
				t.Errorf("NewScalarOID(%v).IsScalar() = false", oid)
			}
		}
	})

	t.Run("IsScalar коррелирует с последним компонентом", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 1},
			{1, 3, 6, 1, 2},
			{1, 3, 6, 1, 39},
		}

		for _, scalar := range scalars {
			last := scalar[len(scalar)-1]
			expected := last == 0

			if scalar.IsScalar() != expected {
				t.Errorf("IsScalar(%v) = %v, ожидалось %v", scalar, scalar.IsScalar(), expected)
			}
		}
	})

	t.Run("Пустой OID не скалярный", func(t *testing.T) {
		// Пустой ScalarOID
		emptyScalar := ScalarOID{}
		if emptyScalar.IsScalar() {
			t.Error("Пустой OID не должен быть скалярным")
		}

		// Nil ScalarOID
		var nilScalar ScalarOID
		if nilScalar.IsScalar() {
			t.Error("Nil OID не должен быть скалярным")
		}
	})
}

// Тест с round trip
func TestScalarOIDIsScalarRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			// Создаем скалярный
			scalar := NewScalarOID(oid)

			// Должен быть скалярным
			if !scalar.IsScalar() {
				t.Error("NewScalarOID должен создавать скалярный")
			}

			// Base не должен быть скалярным (если оригинал не заканчивался на 0)
			base := scalar.Base()
			baseScalar := ScalarOID(base)
			if len(base) > 0 && base[len(base)-1] != 0 {
				if baseScalar.IsScalar() {
					t.Error("Base не должен быть скалярным")
				}
			}
		})
	}
}

// Тест с подтестами
func TestScalarOIDIsScalarCategories(t *testing.T) {
	t.Run("Скалярные OID", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{2, 100, 3, 0},
			{0},
			{0, 0},
		}

		for _, scalar := range scalars {
			if !scalar.IsScalar() {
				t.Errorf("IsScalar(%v) = false, ожидалось true", scalar)
			}
		}
	})

	t.Run("Не скалярные OID", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1},
			{1, 3, 6, 1, 1},
			{1, 3, 6, 39},
			{},
			nil,
		}

		for _, scalar := range scalars {
			if scalar.IsScalar() {
				t.Errorf("IsScalar(%v) = true, ожидалось false", scalar)
			}
		}
	})

	t.Run("Граничные случаи", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected bool
		}{
			{ScalarOID{0}, true},
			{ScalarOID{0, 0}, true},
			{ScalarOID{1, 0}, true},
			{ScalarOID{0, 1}, false},
			{ScalarOID{1, 0, 0}, true},
			{ScalarOID{0, 0, 1}, false},
		}

		for _, tt := range tests {
			if got := tt.scalar.IsScalar(); got != tt.expected {
				t.Errorf("IsScalar(%v) = %v, ожидалось %v", tt.scalar, got, tt.expected)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_IsScalar() {
	// Скалярный OID
	scalar1 := ScalarOID{1, 3, 6, 1, 0}
	fmt.Println(scalar1.IsScalar())

	// Не скалярный OID
	scalar2 := ScalarOID{1, 3, 6, 1}
	fmt.Println(scalar2.IsScalar())

	// Пустой OID
	scalar3 := ScalarOID{}
	fmt.Println(scalar3.IsScalar())

	// Output:
	// true
	// false
	// false
}

// Бенчмарк
func BenchmarkScalarOIDIsScalar(b *testing.B) {
	scalars := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1},
		{},
		nil,
	}

	for _, scalar := range scalars {
		b.Run(scalar.String(), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = scalar.IsScalar()
			}
		})
	}
}

func TestScalarOIDBase(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		expected OID
	}{
		{
			name:     "Скалярный с .0",
			scalar:   ScalarOID{1, 3, 6, 1, 0},
			expected: OID{1, 3, 6, 1},
		},
		{
			name:     "Длинный скалярный",
			scalar:   ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: OID{1, 3, 6, 1, 2, 1, 1, 1},
		},
		{
			name:     "Без .0 (не скалярный)",
			scalar:   ScalarOID{1, 3, 6, 1},
			expected: OID{1, 3, 6, 1},
		},
		{
			name:     "Короткий с .0",
			scalar:   ScalarOID{1, 3, 0},
			expected: OID{1, 3},
		},
		{
			name:     "Пустой",
			scalar:   ScalarOID{},
			expected: nil,
		},
		{
			name:     "Nil",
			scalar:   nil,
			expected: nil,
		},
		{
			name:     "Только 0",
			scalar:   ScalarOID{0},
			expected: OID{},
		},
		{
			name:     "Два 0",
			scalar:   ScalarOID{0, 0},
			expected: OID{0},
		},
		{
			name:     "Последний не 0",
			scalar:   ScalarOID{1, 3, 6, 1},
			expected: OID{1, 3, 6, 1},
		},
		{
			name:     "0 в середине",
			scalar:   ScalarOID{1, 0, 3},
			expected: OID{1, 0, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scalar.Base()

			if !result.Equal(tt.expected) {
				t.Errorf("Base() = %v, ожидалось %v", result, tt.expected)
			}

			// Проверяем длину
			if len(result) != len(tt.expected) {
				t.Errorf("len(Base()) = %d, ожидалось %d", len(result), len(tt.expected))
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDBaseProperties(t *testing.T) {
	t.Run("Base всегда короче или равен", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{0},
			{},
		}

		for _, scalar := range scalars {
			base := scalar.Base()
			if len(base) > len(scalar) {
				t.Errorf("Base() длиннее оригинала: %d > %d", len(base), len(scalar))
			}
		}
	})

	t.Run("Base от скалярного на 1 короче", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
		}

		for _, scalar := range scalars {
			base := scalar.Base()
			if len(base) != len(scalar)-1 {
				t.Errorf("len(Base()) = %d, ожидалось %d", len(base), len(scalar)-1)
			}
		}
	})

	t.Run("Base от не скалярного равен оригиналу", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1},
			{2, 100, 3},
		}

		for _, scalar := range scalars {
			base := scalar.Base()
			if len(base) != len(scalar) {
				t.Errorf("len(Base()) = %d, ожидалось %d", len(base), len(scalar))
			}
		}
	})

	t.Run("Не изменяет оригинал", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}
		scalarCopy := make(ScalarOID, len(scalar))
		copy(scalarCopy, scalar)

		scalar.Base()

		if !scalar.Equal(scalarCopy) {
			t.Error("Base() не должен изменять оригинал")
		}
	})
}

// Тест с round trip
func TestScalarOIDBaseRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			// Создаем скалярный
			scalar := NewScalarOID(oid)

			// Получаем base
			base := scalar.Base()

			// Base должен совпадать с оригиналом (если не заканчивался на 0)
			if len(oid) > 0 && oid[len(oid)-1] != 0 {
				if !base.Equal(oid) {
					t.Errorf("Base = %v, ожидалось %v", base, oid)
				}
			}
		})
	}
}

// Тест с подтестами
func TestScalarOIDBaseCategories(t *testing.T) {
	t.Run("Скалярные OID", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected OID
		}{
			{ScalarOID{1, 3, 6, 1, 0}, OID{1, 3, 6, 1}},
			{ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}, OID{1, 3, 6, 1, 2, 1, 1, 1}},
			{ScalarOID{2, 100, 3, 0}, OID{2, 100, 3}},
		}

		for _, tt := range tests {
			result := tt.scalar.Base()
			if !result.Equal(tt.expected) {
				t.Errorf("Base() = %v, ожидалось %v", result, tt.expected)
			}
		}
	})

	t.Run("Не скалярные OID", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected OID
		}{
			{ScalarOID{1, 3, 6, 1}, OID{1, 3, 6, 1}},
			{ScalarOID{2, 100, 3}, OID{2, 100, 3}},
		}

		for _, tt := range tests {
			result := tt.scalar.Base()
			if !result.Equal(tt.expected) {
				t.Errorf("Base() = %v, ожидалось %v", result, tt.expected)
			}
		}
	})

	t.Run("Граничные случаи", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected OID
		}{
			{ScalarOID{}, nil},
			{nil, nil},
			{ScalarOID{0}, OID{}},
			{ScalarOID{0, 0}, OID{0}},
			{ScalarOID{1, 0}, OID{1}},
		}

		for _, tt := range tests {
			result := tt.scalar.Base()
			if !result.Equal(tt.expected) {
				t.Errorf("Base() = %v, ожидалось %v", result, tt.expected)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_Base() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	base := scalar.Base()

	fmt.Println(scalar)
	fmt.Println(base)
	// Output:
	// 1.3.6.1.2.1.1.1.0
	// 1.3.6.1.2.1.1.1
}

// Пример с не скалярным
func ExampleScalarOID_Base_noZero() {
	scalar := ScalarOID{1, 3, 6, 1}

	base := scalar.Base()

	fmt.Println(scalar)
	fmt.Println(base)
	// Output:
	// 1.3.6.1
	// 1.3.6.1
}

// Бенчмарк
func BenchmarkScalarOIDBase(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_ = scalar.Base()
	}
}

func TestScalarOIDString(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		expected string
	}{
		{
			name:     "Скалярный OID",
			scalar:   ScalarOID{1, 3, 6, 1, 0},
			expected: "1.3.6.1.0",
		},
		{
			name:     "Длинный скалярный",
			scalar:   ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: "1.3.6.1.2.1.1.1.0",
		},
		{
			name:     "Не скалярный",
			scalar:   ScalarOID{1, 3, 6, 1},
			expected: "1.3.6.1",
		},
		{
			name:     "Короткий",
			scalar:   ScalarOID{1, 3},
			expected: "1.3",
		},
		{
			name:     "С первым 2",
			scalar:   ScalarOID{2, 100, 3, 0},
			expected: "2.100.3.0",
		},
		{
			name:     "С первым 0",
			scalar:   ScalarOID{0, 39, 1, 0},
			expected: "0.39.1.0",
		},
		{
			name:     "Пустой",
			scalar:   ScalarOID{},
			expected: "",
		},
		{
			name:     "Nil",
			scalar:   nil,
			expected: "",
		},
		{
			name:     "Только 0",
			scalar:   ScalarOID{0},
			expected: "0",
		},
		{
			name:     "С большим компонентом",
			scalar:   ScalarOID{1, 3, MaxOIDComponent, 0},
			expected: "1.3.268435455.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scalar.String()

			if result != tt.expected {
				t.Errorf("String() = %q, ожидалось %q", result, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDStringProperties(t *testing.T) {
	t.Run("Эквивалентна OID.String()", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{2, 100, 3, 0},
		}

		for _, scalar := range scalars {
			scalarStr := scalar.String()
			oidStr := OID(scalar).String()

			if scalarStr != oidStr {
				t.Errorf("String() = %q, OID.String() = %q", scalarStr, oidStr)
			}
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}
		scalarCopy := make(ScalarOID, len(scalar))
		copy(scalarCopy, scalar)

		scalar.String()

		if !scalar.Equal(scalarCopy) {
			t.Error("String() не должен изменять OID")
		}
	})

	t.Run("Содержит точки между компонентами", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		str := scalar.String()

		// Проверяем количество точек
		dotCount := 0
		for _, ch := range str {
			if ch == '.' {
				dotCount++
			}
		}

		expectedDots := len(scalar) - 1
		if dotCount != expectedDots {
			t.Errorf("Количество точек = %d, ожидалось %d", dotCount, expectedDots)
		}
	})
}

// Тест с round trip
func TestScalarOIDStringRoundTrip(t *testing.T) {
	scalars := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
	}

	for _, scalar := range scalars {
		t.Run(scalar.String(), func(t *testing.T) {
			// String -> ParseOID -> ScalarOID
			str := scalar.String()

			parsed, err := ParseOID(str)
			if err != nil {
				t.Fatalf("ParseOID(%q): %v", str, err)
			}

			backToScalar := ScalarOID(parsed)

			if !backToScalar.Equal(scalar) {
				t.Errorf("Round trip: %v -> %q -> %v", scalar, str, backToScalar)
			}
		})
	}
}

// Тест с подтестами
func TestScalarOIDStringCategories(t *testing.T) {
	t.Run("Скалярные OID", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected string
		}{
			{ScalarOID{1, 3, 6, 1, 0}, "1.3.6.1.0"},
			{ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}, "1.3.6.1.2.1.1.1.0"},
		}

		for _, tt := range tests {
			if got := tt.scalar.String(); got != tt.expected {
				t.Errorf("String() = %q, ожидалось %q", got, tt.expected)
			}
		}
	})

	t.Run("Не скалярные OID", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected string
		}{
			{ScalarOID{1, 3, 6, 1}, "1.3.6.1"},
			{ScalarOID{2, 100, 3}, "2.100.3"},
		}

		for _, tt := range tests {
			if got := tt.scalar.String(); got != tt.expected {
				t.Errorf("String() = %q, ожидалось %q", got, tt.expected)
			}
		}
	})

	t.Run("Пустые OID", func(t *testing.T) {
		// Пустой ScalarOID
		emptyScalar := ScalarOID{}
		if got := emptyScalar.String(); got != "" {
			t.Errorf("String() = %q, ожидалась пустая строка", got)
		}

		// Nil ScalarOID
		var nilScalar ScalarOID
		if got := nilScalar.String(); got != "" {
			t.Errorf("String() = %q, ожидалась пустая строка", got)
		}
	})
}

// Пример использования
func ExampleScalarOID_String() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	fmt.Println(scalar.String())
	// Output: 1.3.6.1.2.1.1.1.0
}

// Пример с пустым OID
func ExampleScalarOID_String_empty() {
	scalar := ScalarOID{}

	fmt.Println(scalar.String() == "")
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDString(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_ = scalar.String()
	}
}

func TestScalarOIDValidate(t *testing.T) {
	tests := []struct {
		name    string
		scalar  ScalarOID
		wantErr error
	}{
		{
			name:    "Валидный скалярный",
			scalar:  ScalarOID{1, 3, 6, 1, 0},
			wantErr: nil,
		},
		{
			name:    "Валидный длинный",
			scalar:  ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr: nil,
		},
		{
			name:    "Валидный не скалярный",
			scalar:  ScalarOID{1, 3, 6, 1},
			wantErr: nil,
		},
		{
			name:    "Валидный с первым 2",
			scalar:  ScalarOID{2, 100, 3, 0},
			wantErr: nil,
		},
		{
			name:    "Валидный с первым 0",
			scalar:  ScalarOID{0, 39, 1, 0},
			wantErr: nil,
		},
		{
			name:    "Пустой",
			scalar:  ScalarOID{},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Nil",
			scalar:  nil,
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Один компонент",
			scalar:  ScalarOID{1},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Первый > 2",
			scalar:  ScalarOID{3, 1, 0},
			wantErr: ErrFirstComponentTooBig,
		},
		{
			name:    "Второй > 39 при первом 0",
			scalar:  ScalarOID{0, 40, 0},
			wantErr: ErrSecondComponentTooBig,
		},
		{
			name:    "Второй > 39 при первом 1",
			scalar:  ScalarOID{1, 40, 0},
			wantErr: ErrSecondComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scalar.Validate()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Validate: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate: %v", err)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDValidateProperties(t *testing.T) {
	t.Run("Эквивалентна OID.Validate()", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1},
			{},
		}

		for _, scalar := range scalars {
			scalarErr := scalar.Validate()
			oidErr := OID(scalar).Validate()

			if (scalarErr == nil) != (oidErr == nil) {
				t.Errorf("ScalarOID.Validate() = %v, OID.Validate() = %v",
					scalarErr, oidErr)
			}
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}
		scalarCopy := make(ScalarOID, len(scalar))
		copy(scalarCopy, scalar)

		scalar.Validate()

		if !scalar.Equal(scalarCopy) {
			t.Error("Validate() не должен изменять OID")
		}
	})

	t.Run("Валидные OID проходят", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{2, 100, 3, 0},
		}

		for _, scalar := range scalars {
			if err := scalar.Validate(); err != nil {
				t.Errorf("Validate(%v): %v", scalar, err)
			}
		}
	})
}

// Тест с round trip
func TestScalarOIDValidateRoundTrip(t *testing.T) {
	// NewScalarOID должен создавать валидные OID
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			scalar := NewScalarOID(oid)

			if err := scalar.Validate(); err != nil {
				t.Errorf("NewScalarOID(%v).Validate(): %v", oid, err)
			}
		})
	}
}

// Тест с подтестами
func TestScalarOIDValidateCategories(t *testing.T) {
	t.Run("Валидные OID", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{2, 100, 3, 0},
			{0, 39, 1, 0},
		}

		for _, scalar := range scalars {
			if err := scalar.Validate(); err != nil {
				t.Errorf("Validate(%v): %v", scalar, err)
			}
		}
	})

	t.Run("Невалидные OID", func(t *testing.T) {
		tests := []struct {
			scalar  ScalarOID
			wantErr error
		}{
			{ScalarOID{}, ErrOIDTooShort},
			{ScalarOID{1}, ErrOIDTooShort},
			{ScalarOID{3, 1, 0}, ErrFirstComponentTooBig},
			{ScalarOID{1, 40, 0}, ErrSecondComponentTooBig},
		}

		for _, tt := range tests {
			err := tt.scalar.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate(%v) = %v, ожидалось %v", tt.scalar, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_Validate() {
	// Валидный OID
	valid := ScalarOID{1, 3, 6, 1, 0}
	fmt.Println(valid.Validate() == nil)

	// Невалидный OID
	invalid := ScalarOID{3, 1, 0}
	fmt.Println(errors.Is(invalid.Validate(), ErrFirstComponentTooBig))
	// Output:
	// true
	// true
}

// Пример с пустым OID
func ExampleScalarOID_Validate_empty() {
	empty := ScalarOID{}
	fmt.Println(errors.Is(empty.Validate(), ErrOIDTooShort))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDValidate(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_ = scalar.Validate()
	}
}

func TestScalarOIDEqual(t *testing.T) {
	tests := []struct {
		name     string
		scalar1  ScalarOID
		scalar2  ScalarOID
		expected bool
	}{
		{
			name:     "Одинаковые скалярные",
			scalar1:  ScalarOID{1, 3, 6, 1, 0},
			scalar2:  ScalarOID{1, 3, 6, 1, 0},
			expected: true,
		},
		{
			name:     "Одинаковые длинные",
			scalar1:  ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			scalar2:  ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: true,
		},
		{
			name:     "Одинаковые не скалярные",
			scalar1:  ScalarOID{1, 3, 6, 1},
			scalar2:  ScalarOID{1, 3, 6, 1},
			expected: true,
		},
		{
			name:     "Разные OID",
			scalar1:  ScalarOID{1, 3, 6, 1, 0},
			scalar2:  ScalarOID{1, 3, 6, 2, 0},
			expected: false,
		},
		{
			name:     "Разная длина",
			scalar1:  ScalarOID{1, 3, 6, 0},
			scalar2:  ScalarOID{1, 3, 6, 1, 0},
			expected: false,
		},
		{
			name:     "Оба пустые",
			scalar1:  ScalarOID{},
			scalar2:  ScalarOID{},
			expected: true,
		},
		{
			name:     "Оба nil",
			scalar1:  nil,
			scalar2:  nil,
			expected: true,
		},
		{
			name:     "Пустой и nil",
			scalar1:  ScalarOID{},
			scalar2:  nil,
			expected: true,
		},
		{
			name:     "Пустой и непустой",
			scalar1:  ScalarOID{},
			scalar2:  ScalarOID{1, 3, 0},
			expected: false,
		},
		{
			name:     "Скалярный и не скалярный",
			scalar1:  ScalarOID{1, 3, 6, 0},
			scalar2:  ScalarOID{1, 3, 6},
			expected: false,
		},
		{
			name:     "Разные первые компоненты",
			scalar1:  ScalarOID{1, 3, 6, 0},
			scalar2:  ScalarOID{2, 3, 6, 0},
			expected: false,
		},
		{
			name:     "С первым 2",
			scalar1:  ScalarOID{2, 100, 3, 0},
			scalar2:  ScalarOID{2, 100, 3, 0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scalar1.Equal(tt.scalar2)

			if result != tt.expected {
				t.Errorf("Equal(%v, %v) = %v, ожидалось %v",
					tt.scalar1, tt.scalar2, result, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDEqualProperties(t *testing.T) {
	t.Run("Рефлексивность", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1},
			{},
		}

		for _, scalar := range scalars {
			if !scalar.Equal(scalar) {
				t.Errorf("Equal(%v, %v) = false, ожидалось true", scalar, scalar)
			}
		}
	})

	t.Run("Симметричность", func(t *testing.T) {
		scalar1 := ScalarOID{1, 3, 6, 1, 0}
		scalar2 := ScalarOID{1, 3, 6, 1, 0}

		if scalar1.Equal(scalar2) != scalar2.Equal(scalar1) {
			t.Error("Equal должен быть симметричным")
		}
	})

	t.Run("Транзитивность", func(t *testing.T) {
		scalar1 := ScalarOID{1, 3, 6, 1, 0}
		scalar2 := ScalarOID{1, 3, 6, 1, 0}
		scalar3 := ScalarOID{1, 3, 6, 1, 0}

		if scalar1.Equal(scalar2) && scalar2.Equal(scalar3) {
			if !scalar1.Equal(scalar3) {
				t.Error("Equal должен быть транзитивным")
			}
		}
	})

	t.Run("Эквивалентна OID.Equal()", func(t *testing.T) {
		scalar1 := ScalarOID{1, 3, 6, 1, 0}
		scalar2 := ScalarOID{1, 3, 6, 1, 0}

		if scalar1.Equal(scalar2) != OID(scalar1).Equal(OID(scalar2)) {
			t.Error("ScalarOID.Equal должен совпадать с OID.Equal")
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		scalar1 := ScalarOID{1, 3, 6, 1, 0}
		scalar2 := ScalarOID{1, 3, 6, 1, 0}

		scalar1Copy := make(ScalarOID, len(scalar1))
		copy(scalar1Copy, scalar1)

		scalar1.Equal(scalar2)

		if !scalar1.Equal(scalar1Copy) {
			t.Error("Equal() не должен изменять OID")
		}
	})
}

// Тест с round trip
func TestScalarOIDEqualRoundTrip(t *testing.T) {
	// NewScalarOID создает одинаковые OID
	oid := OID{1, 3, 6, 1}

	scalar1 := NewScalarOID(oid)
	scalar2 := NewScalarOID(oid)

	if !scalar1.Equal(scalar2) {
		t.Error("NewScalarOID должен создавать равные OID")
	}
}

// Тест с подтестами
func TestScalarOIDEqualCategories(t *testing.T) {
	t.Run("Равные OID", func(t *testing.T) {
		tests := []struct {
			scalar1 ScalarOID
			scalar2 ScalarOID
		}{
			{ScalarOID{1, 3, 6, 0}, ScalarOID{1, 3, 6, 0}},
			{ScalarOID{1, 3, 6, 1}, ScalarOID{1, 3, 6, 1}},
			{ScalarOID{}, ScalarOID{}},
			{nil, nil},
		}

		for _, tt := range tests {
			if !tt.scalar1.Equal(tt.scalar2) {
				t.Errorf("Equal(%v, %v) = false, ожидалось true", tt.scalar1, tt.scalar2)
			}
		}
	})

	t.Run("Неравные OID", func(t *testing.T) {
		tests := []struct {
			scalar1 ScalarOID
			scalar2 ScalarOID
		}{
			{ScalarOID{1, 3, 6, 0}, ScalarOID{1, 3, 6, 1}},
			{ScalarOID{1, 3, 6}, ScalarOID{1, 3, 6, 0}},
			{ScalarOID{}, ScalarOID{1, 3, 0}},
		}

		for _, tt := range tests {
			if tt.scalar1.Equal(tt.scalar2) {
				t.Errorf("Equal(%v, %v) = true, ожидалось false", tt.scalar1, tt.scalar2)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_Equal() {
	scalar1 := ScalarOID{1, 3, 6, 1, 0}
	scalar2 := ScalarOID{1, 3, 6, 1, 0}
	scalar3 := ScalarOID{2, 100, 3, 0}

	fmt.Println(scalar1.Equal(scalar2))
	fmt.Println(scalar1.Equal(scalar3))
	// Output:
	// true
	// false
}

// Пример с пустыми OID
func ExampleScalarOID_Equal_empty() {
	empty1 := ScalarOID{}
	empty2 := ScalarOID{}

	fmt.Println(empty1.Equal(empty2))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDEqual(b *testing.B) {
	scalar1 := ScalarOID{1, 3, 6, 1, 0}
	scalar2 := ScalarOID{1, 3, 6, 1, 0}

	b.ReportAllocs()
	for b.Loop() {
		_ = scalar1.Equal(scalar2)
	}
}

func TestScalarOIDStartsWith(t *testing.T) {
	scalar := ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}

	tests := []struct {
		name     string
		prefix   OID
		expected bool
	}{
		{
			name:     "Полное совпадение",
			prefix:   OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: true,
		},
		{
			name:     "Без последнего .0",
			prefix:   OID{1, 3, 6, 1, 2, 1, 1, 1},
			expected: true,
		},
		{
			name:     "Короткий префикс",
			prefix:   OID{1, 3, 6, 1},
			expected: true,
		},
		{
			name:     "Минимальный префикс",
			prefix:   OID{1, 3},
			expected: true,
		},
		{
			name:     "Один компонент",
			prefix:   OID{1},
			expected: true,
		},
		{
			name:     "Пустой префикс",
			prefix:   OID{},
			expected: true,
		},
		{
			name:     "Nil префикс",
			prefix:   nil,
			expected: true,
		},
		{
			name:     "Не совпадает",
			prefix:   OID{1, 3, 6, 1, 2, 1, 1, 2},
			expected: false,
		},
		{
			name:     "Другая ветка",
			prefix:   OID{1, 3, 6, 1, 2, 1, 2},
			expected: false,
		},
		{
			name:     "Префикс длиннее",
			prefix:   OID{1, 3, 6, 1, 2, 1, 1, 1, 0, 1},
			expected: false,
		},
		{
			name:     "Совсем другой",
			prefix:   OID{2, 100, 3},
			expected: false,
		},
		{
			name:     "Первый компонент отличается",
			prefix:   OID{0, 3, 6, 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scalar.StartsWith(tt.prefix)

			if result != tt.expected {
				t.Errorf("StartsWith(%v) = %v, ожидалось %v",
					tt.prefix, result, tt.expected)
			}
		})
	}
}

// Тест с разными ScalarOID
func TestScalarOIDStartsWithMultiple(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		prefix   OID
		expected bool
	}{
		{
			name:     "Короткий скалярный",
			scalar:   ScalarOID{1, 3, 6, 0},
			prefix:   OID{1, 3, 6},
			expected: true,
		},
		{
			name:     "Средний скалярный",
			scalar:   ScalarOID{1, 3, 6, 1, 2, 1, 1, 0},
			prefix:   OID{1, 3, 6, 1, 2},
			expected: true,
		},
		{
			name:     "Пустой скалярный",
			scalar:   ScalarOID{},
			prefix:   OID{1, 3},
			expected: false,
		},
		{
			name:     "Nil скалярный с пустым префиксом",
			scalar:   nil,
			prefix:   OID{},
			expected: true,
		},
		{
			name:     "Пустой с пустым префиксом",
			scalar:   ScalarOID{},
			prefix:   OID{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scalar.StartsWith(tt.prefix)
			if result != tt.expected {
				t.Errorf("StartsWith(%v) = %v, ожидалось %v",
					tt.prefix, result, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDStartsWithProperties(t *testing.T) {
	scalar := ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}

	t.Run("Всегда начинается с самого себя", func(t *testing.T) {
		if !scalar.StartsWith(OID(scalar)) {
			t.Error("StartsWith(self) = false, ожидалось true")
		}
	})

	t.Run("Всегда начинается с Base", func(t *testing.T) {
		base := scalar.Base()
		if !scalar.StartsWith(base) {
			t.Error("StartsWith(Base) = false, ожидалось true")
		}
	})

	t.Run("Всегда начинается с пустого префикса", func(t *testing.T) {
		if !scalar.StartsWith(OID{}) {
			t.Error("StartsWith(empty) = false, ожидалось true")
		}
		if !scalar.StartsWith(nil) {
			t.Error("StartsWith(nil) = false, ожидалось true")
		}
	})

	t.Run("Транзитивность", func(t *testing.T) {
		prefix1 := OID{1, 3, 6}
		prefix2 := OID{1, 3}

		if scalar.StartsWith(prefix1) && prefix1.StartsWith(prefix2) {
			if !scalar.StartsWith(prefix2) {
				t.Error("Транзитивность нарушена")
			}
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		scalarCopy := make(ScalarOID, len(scalar))
		copy(scalarCopy, scalar)

		scalar.StartsWith(OID{1, 3})

		if !scalar.Equal(scalarCopy) {
			t.Error("StartsWith() не должен изменять OID")
		}
	})
}

// Тест с round trip
func TestScalarOIDStartsWithRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			scalar := NewScalarOID(oid)

			// Scalar должен начинаться с оригинального OID
			if !scalar.StartsWith(oid) {
				t.Errorf("StartsWith(%v) = false", oid)
			}

			// Base должен начинаться с оригинального OID
			base := scalar.Base()
			if len(oid) > 0 && oid[len(oid)-1] != 0 {
				if !base.StartsWith(oid) {
					t.Errorf("Base.StartsWith(%v) = false", oid)
				}
			}
		})
	}
}

// Тест с подтестами
func TestScalarOIDStartsWithCategories(t *testing.T) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	t.Run("Совпадающие префиксы", func(t *testing.T) {
		prefixes := []OID{
			{1},
			{1, 3},
			{1, 3, 6},
			{1, 3, 6, 1},
			{1, 3, 6, 1, 2},
			{1, 3, 6, 1, 2, 1},
			{1, 3, 6, 1, 2, 1, 1},
			{1, 3, 6, 1, 2, 1, 1, 1},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
		}

		for _, prefix := range prefixes {
			if !scalar.StartsWith(prefix) {
				t.Errorf("StartsWith(%v) = false, ожидалось true", prefix)
			}
		}
	})

	t.Run("Несовпадающие префиксы", func(t *testing.T) {
		prefixes := []OID{
			{2},
			{0, 3},
			{1, 4},
			{1, 3, 7},
			{1, 3, 6, 2},
			{1, 3, 6, 1, 2, 1, 1, 2},
			{1, 3, 6, 1, 2, 1, 1, 1, 0, 1},
		}

		for _, prefix := range prefixes {
			if scalar.StartsWith(prefix) {
				t.Errorf("StartsWith(%v) = true, ожидалось false", prefix)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_StartsWith() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	fmt.Println(scalar.StartsWith(MustParseOID("1.3.6.1")))
	fmt.Println(scalar.StartsWith(MustParseOID("2.100.3")))
	fmt.Println(scalar.StartsWith(OID{}))
	// Output:
	// true
	// false
	// true
}

// Бенчмарк
func BenchmarkScalarOIDStartsWith(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")
	prefix := MustParseOID("1.3.6.1")

	b.ReportAllocs()
	for b.Loop() {
		_ = scalar.StartsWith(prefix)
	}
}

func TestScalarOIDAppend(t *testing.T) {
	tests := []struct {
		name       string
		scalar     ScalarOID
		components []uint32
		expected   ScalarOID
	}{
		{
			name:       "Добавление одного компонента",
			scalar:     ScalarOID{1, 3, 6, 0},
			components: []uint32{1},
			expected:   ScalarOID{1, 3, 6, 0, 1},
		},
		{
			name:       "Добавление нескольких компонентов",
			scalar:     ScalarOID{1, 3, 6, 0},
			components: []uint32{1, 2, 3},
			expected:   ScalarOID{1, 3, 6, 0, 1, 2, 3},
		},
		{
			name:       "Без компонентов",
			scalar:     ScalarOID{1, 3, 6, 0},
			components: []uint32{},
			expected:   ScalarOID{1, 3, 6, 0},
		},
		{
			name:       "Nil components",
			scalar:     ScalarOID{1, 3, 6, 0},
			components: nil,
			expected:   ScalarOID{1, 3, 6, 0},
		},
		{
			name:       "Добавление нуля",
			scalar:     ScalarOID{1, 3, 6},
			components: []uint32{0},
			expected:   ScalarOID{1, 3, 6, 0},
		},
		{
			name:       "Добавление к пустому",
			scalar:     ScalarOID{},
			components: []uint32{1, 3},
			expected:   ScalarOID{1, 3},
		},
		{
			name:       "Добавление к nil",
			scalar:     nil,
			components: []uint32{1, 3, 6},
			expected:   ScalarOID{1, 3, 6},
		},
		{
			name:       "Добавление больших значений",
			scalar:     ScalarOID{1, 3},
			components: []uint32{MaxOIDComponent},
			expected:   ScalarOID{1, 3, MaxOIDComponent},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scalar.Append(tt.components...)

			if !result.Equal(tt.expected) {
				t.Errorf("Append(%v...) = %v, ожидалось %v",
					tt.components, result, tt.expected)
			}

			// Проверяем длину
			if len(result) != len(tt.expected) {
				t.Errorf("len(Append) = %d, ожидалось %d",
					len(result), len(tt.expected))
			}
		})
	}
}

// Тест с проверкой неизменности оригинала
func TestScalarOIDAppendImmutability(t *testing.T) {
	original := ScalarOID{1, 3, 6, 0}
	originalCopy := make(ScalarOID, len(original))
	copy(originalCopy, original)

	// Выполняем Append
	result := original.Append(1, 2, 3)

	// Проверяем, что оригинал не изменился
	if !original.Equal(originalCopy) {
		t.Errorf("Оригинал изменился: %v -> %v", originalCopy, original)
	}

	// Проверяем, что результат отличается от оригинала
	if result.Equal(original) {
		t.Error("Результат должен отличаться от оригинала")
	}

	// Проверяем, что результат длиннее
	if len(result) <= len(original) {
		t.Error("Результат должен быть длиннее оригинала")
	}
}

// Тест с проверкой свойств
func TestScalarOIDAppendProperties(t *testing.T) {
	t.Run("Append увеличивает длину", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 0}

		result := scalar.Append(1)
		if len(result) != len(scalar)+1 {
			t.Errorf("len = %d, ожидалось %d", len(result), len(scalar)+1)
		}

		result = scalar.Append(1, 2, 3)
		if len(result) != len(scalar)+3 {
			t.Errorf("len = %d, ожидалось %d", len(result), len(scalar)+3)
		}

		result = scalar.Append()
		if len(result) != len(scalar) {
			t.Errorf("len = %d, ожидалось %d", len(result), len(scalar))
		}
	})

	t.Run("Append сохраняет префикс", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 0}

		result := scalar.Append(1, 2, 3)

		if !result.StartsWith(OID(scalar)) {
			t.Error("Результат должен начинаться с оригинала")
		}
	})

	t.Run("Append пустого списка возвращает копию", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 0}

		result := scalar.Append()

		if !result.Equal(scalar) {
			t.Error("Append() должен вернуть тот же OID")
		}
	})
}

// Тест с подтестами
func TestScalarOIDAppendCategories(t *testing.T) {
	t.Run("Добавление к скалярному", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 0}

		tests := []struct {
			components []uint32
			expected   ScalarOID
		}{
			{[]uint32{1}, ScalarOID{1, 3, 6, 0, 1}},
			{[]uint32{1, 2}, ScalarOID{1, 3, 6, 0, 1, 2}},
			{[]uint32{}, ScalarOID{1, 3, 6, 0}},
		}

		for _, tt := range tests {
			result := scalar.Append(tt.components...)
			if !result.Equal(tt.expected) {
				t.Errorf("Append(%v) = %v, ожидалось %v",
					tt.components, result, tt.expected)
			}
		}
	})

	t.Run("Добавление к не скалярному", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6}

		result := scalar.Append(0)

		if !result.IsScalar() {
			t.Error("Append(0) должен создать скалярный OID")
		}
	})

	t.Run("Граничные случаи", func(t *testing.T) {
		tests := []struct {
			scalar     ScalarOID
			components []uint32
			expected   ScalarOID
		}{
			{ScalarOID{}, []uint32{1}, ScalarOID{1}},
			{nil, []uint32{1, 2}, ScalarOID{1, 2}},
			{ScalarOID{0}, []uint32{0}, ScalarOID{0, 0}},
			{ScalarOID{1, 0}, []uint32{}, ScalarOID{1, 0}},
		}

		for _, tt := range tests {
			result := tt.scalar.Append(tt.components...)
			if !result.Equal(tt.expected) {
				t.Errorf("Append(%v) = %v, ожидалось %v",
					tt.components, result, tt.expected)
			}
		}
	})
}

// Тест с round trip
func TestScalarOIDAppendRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			scalar := NewScalarOID(oid)

			// Append и проверяем
			appended := scalar.Append(99)

			// Проверяем, что StartsWith работает
			if !appended.StartsWith(OID(scalar)) {
				t.Error("Append должен сохранять префикс")
			}

			// Проверяем, что Base работает
			base := appended.Base()
			if !base.StartsWith(OID(scalar)) {
				t.Error("Base должен сохранять префикс")
			}
		})
	}
}

// Пример использования
func ExampleScalarOID_Append() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	extended := scalar.Append(1, 2, 3)

	fmt.Println(scalar)
	fmt.Println(extended)
	// Output:
	// 1.3.6.1.2.1.1.1.0
	// 1.3.6.1.2.1.1.1.0.1.2.3
}

// Пример с добавлением 0
func ExampleScalarOID_Append_zero() {
	scalar := ScalarOID{1, 3, 6}

	scalarWithZero := scalar.Append(0)

	fmt.Println(scalar.IsScalar())
	fmt.Println(scalarWithZero.IsScalar())
	// Output:
	// false
	// true
}

// Бенчмарк
func BenchmarkScalarOIDAppend(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_ = scalar.Append(1)
	}
}

func TestScalarOIDParent(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		expected OID
		wantErr  error
	}{
		{
			name:     "Скалярный с .0",
			scalar:   ScalarOID{1, 3, 6, 1, 0},
			expected: OID{1, 3, 6, 1},
			wantErr:  nil,
		},
		{
			name:     "Длинный скалярный",
			scalar:   ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: OID{1, 3, 6, 1, 2, 1, 1, 1},
			wantErr:  nil,
		},
		{
			name:     "Не скалярный",
			scalar:   ScalarOID{1, 3, 6, 1},
			expected: OID{1, 3, 6},
			wantErr:  nil,
		},
		{
			name:     "Два компонента",
			scalar:   ScalarOID{1, 3},
			expected: OID{1},
			wantErr:  nil,
		},
		{
			name:     "Один компонент",
			scalar:   ScalarOID{1},
			expected: nil,
			wantErr:  ErrNoParent,
		},
		{
			name:     "Пустой",
			scalar:   ScalarOID{},
			expected: nil,
			wantErr:  ErrNoParent,
		},
		{
			name:     "Nil",
			scalar:   nil,
			expected: nil,
			wantErr:  ErrNoParent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.scalar.Parent()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Parent(): ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Parent() = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Parent(): %v", err)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("Parent() = %v, ожидалось %v", result, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDParentProperties(t *testing.T) {
	t.Run("Parent всегда короче", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{1, 3, 6, 1},
			{1, 3},
		}

		for _, scalar := range scalars {
			parent, err := scalar.Parent()
			if err != nil {
				t.Errorf("Parent(%v): %v", scalar, err)
				continue
			}
			if len(parent) != len(scalar)-1 {
				t.Errorf("len(Parent(%v)) = %d, ожидалось %d",
					scalar, len(parent), len(scalar)-1)
			}
		}
	})

	t.Run("Parent сохраняет префикс", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}

		parent, err := scalar.Parent()
		if err != nil {
			t.Fatalf("Parent: %v", err)
		}

		if !scalar.StartsWith(parent) {
			t.Error("Scalar должен начинаться с Parent")
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}
		scalarCopy := make(ScalarOID, len(scalar))
		copy(scalarCopy, scalar)

		scalar.Parent()

		if !scalar.Equal(scalarCopy) {
			t.Error("Parent() не должен изменять OID")
		}
	})

	t.Run("Эквивалентна OID.Parent()", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		scalarParent, scalarErr := scalar.Parent()
		oidParent, oidErr := OID(scalar).Parent()

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && !scalarParent.Equal(oidParent) {
			t.Error("Результаты должны совпадать")
		}
	})
}

// Тест с round trip
func TestScalarOIDParentRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			scalar := NewScalarOID(oid)

			// Получаем parent
			parent, err := scalar.Parent()
			if err != nil {
				t.Fatalf("Parent: %v", err)
			}

			// Parent должен быть короче
			if len(parent) >= len(scalar) {
				t.Error("Parent должен быть короче")
			}

			// Scalar должен начинаться с parent
			if !scalar.StartsWith(parent) {
				t.Error("Scalar должен начинаться с parent")
			}
		})
	}
}

// Тест с подтестами
func TestScalarOIDParentCategories(t *testing.T) {
	t.Run("Успешные случаи", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected OID
		}{
			{ScalarOID{1, 3, 6, 1, 0}, OID{1, 3, 6, 1}},
			{ScalarOID{1, 3, 6, 1}, OID{1, 3, 6}},
			{ScalarOID{1, 3}, OID{1}},
		}

		for _, tt := range tests {
			result, err := tt.scalar.Parent()
			if err != nil {
				t.Errorf("Parent(%v): %v", tt.scalar, err)
				continue
			}
			if !result.Equal(tt.expected) {
				t.Errorf("Parent(%v) = %v, ожидалось %v", tt.scalar, result, tt.expected)
			}
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []ScalarOID{
			{1},
			{},
			nil,
		}

		for _, scalar := range tests {
			_, err := scalar.Parent()
			if !errors.Is(err, ErrNoParent) {
				t.Errorf("Parent(%v) = %v, ожидалось ErrNoParent", scalar, err)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_Parent() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	parent, err := scalar.Parent()
	if err != nil {
		panic(err)
	}

	fmt.Println(scalar)
	fmt.Println(parent)
	// Output:
	// 1.3.6.1.2.1.1.1.0
	// 1.3.6.1.2.1.1.1
}

// Пример с ошибкой
func ExampleScalarOID_Parent_error() {
	scalar := ScalarOID{1}

	_, err := scalar.Parent()
	fmt.Println(errors.Is(err, ErrNoParent))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDParent(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_, _ = scalar.Parent()
	}
}

func TestScalarOIDLast(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		expected uint32
		wantErr  error
	}{
		{
			name:     "Скалярный с .0",
			scalar:   ScalarOID{1, 3, 6, 1, 0},
			expected: 0,
			wantErr:  nil,
		},
		{
			name:     "Длинный скалярный",
			scalar:   ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: 0,
			wantErr:  nil,
		},
		{
			name:     "Не скалярный с 1",
			scalar:   ScalarOID{1, 3, 6, 1},
			expected: 1,
			wantErr:  nil,
		},
		{
			name:     "Не скалярный с 3",
			scalar:   ScalarOID{1, 3, 6, 3},
			expected: 3,
			wantErr:  nil,
		},
		{
			name:     "Один компонент",
			scalar:   ScalarOID{1},
			expected: 1,
			wantErr:  nil,
		},
		{
			name:     "Два компонента",
			scalar:   ScalarOID{1, 3},
			expected: 3,
			wantErr:  nil,
		},
		{
			name:     "С большим компонентом",
			scalar:   ScalarOID{1, 3, MaxOIDComponent},
			expected: MaxOIDComponent,
			wantErr:  nil,
		},
		{
			name:     "Пустой",
			scalar:   ScalarOID{},
			expected: 0,
			wantErr:  ErrEmptyOID,
		},
		{
			name:     "Nil",
			scalar:   nil,
			expected: 0,
			wantErr:  ErrEmptyOID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.scalar.Last()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Last(): ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Last() = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Last(): %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Last() = %d, ожидалось %d", result, tt.expected)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDLastProperties(t *testing.T) {
	t.Run("Last возвращает последний компонент", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1},
			{1, 3},
		}

		for _, scalar := range scalars {
			last, err := scalar.Last()
			if err != nil {
				t.Errorf("Last(%v): %v", scalar, err)
				continue
			}

			expected := scalar[len(scalar)-1]
			if last != expected {
				t.Errorf("Last(%v) = %d, ожидалось %d", scalar, last, expected)
			}
		}
	})

	t.Run("Эквивалентна OID.Last()", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		scalarLast, scalarErr := scalar.Last()
		oidLast, oidErr := OID(scalar).Last()

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && scalarLast != oidLast {
			t.Error("Результаты должны совпадать")
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}
		scalarCopy := make(ScalarOID, len(scalar))
		copy(scalarCopy, scalar)

		scalar.Last()

		if !scalar.Equal(scalarCopy) {
			t.Error("Last() не должен изменять OID")
		}
	})
}

// Тест с round trip
func TestScalarOIDLastRoundTrip(t *testing.T) {
	// Append добавляет компоненты, Last возвращает последний
	scalar := ScalarOID{1, 3, 6}

	appended := scalar.Append(99)

	last, err := appended.Last()
	if err != nil {
		t.Fatalf("Last: %v", err)
	}

	if last != 99 {
		t.Errorf("Last = %d, ожидалось 99", last)
	}
}

// Тест с подтестами
func TestScalarOIDLastCategories(t *testing.T) {
	t.Run("Успешные случаи", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected uint32
		}{
			{ScalarOID{1, 3, 6, 1, 0}, 0},
			{ScalarOID{1, 3, 6, 1}, 1},
			{ScalarOID{1, 3}, 3},
			{ScalarOID{1}, 1},
		}

		for _, tt := range tests {
			result, err := tt.scalar.Last()
			if err != nil {
				t.Errorf("Last(%v): %v", tt.scalar, err)
				continue
			}
			if result != tt.expected {
				t.Errorf("Last(%v) = %d, ожидалось %d", tt.scalar, result, tt.expected)
			}
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []ScalarOID{
			{},
			nil,
		}

		for _, scalar := range tests {
			_, err := scalar.Last()
			if !errors.Is(err, ErrEmptyOID) {
				t.Errorf("Last(%v) = %v, ожидалось ErrEmptyOID", scalar, err)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_Last() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	last, err := scalar.Last()
	if err != nil {
		panic(err)
	}

	fmt.Println(last)
	// Output: 0
}

// Пример с не скалярным
func ExampleScalarOID_Last_noZero() {
	scalar := ScalarOID{1, 3, 6, 1}

	last, err := scalar.Last()
	if err != nil {
		panic(err)
	}

	fmt.Println(last)
	// Output: 1
}

// Пример с ошибкой
func ExampleScalarOID_Last_error() {
	scalar := ScalarOID{}

	_, err := scalar.Last()
	fmt.Println(errors.Is(err, ErrEmptyOID))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDLast(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_, _ = scalar.Last()
	}
}

func TestScalarOIDMarshalBinary(t *testing.T) {
	tests := []struct {
		name    string
		scalar  ScalarOID
		wantErr error
	}{
		{
			name:    "Скалярный",
			scalar:  ScalarOID{1, 3, 6, 1, 0},
			wantErr: nil,
		},
		{
			name:    "Длинный скалярный",
			scalar:  ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr: nil,
		},
		{
			name:    "Не скалярный",
			scalar:  ScalarOID{1, 3, 6, 1},
			wantErr: nil,
		},
		{
			name:    "С первым 2",
			scalar:  ScalarOID{2, 100, 3, 0},
			wantErr: nil,
		},
		{
			name:    "С первым 0",
			scalar:  ScalarOID{0, 39, 1, 0},
			wantErr: nil,
		},
		{
			name:    "Пустой",
			scalar:  ScalarOID{},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Nil",
			scalar:  nil,
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Один компонент",
			scalar:  ScalarOID{1},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Невалидный",
			scalar:  ScalarOID{3, 1, 0},
			wantErr: ErrFirstComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.scalar.MarshalBinary()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("MarshalBinary: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("MarshalBinary = %v, ожидалось %v", err, tt.wantErr)
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

// Тест с сравнением со стандартной библиотекой
func TestScalarOIDMarshalBinaryCompareWithStd(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
		{0, 39, 1, 0},
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Наш MarshalBinary
			ourData, err := scalar.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// Стандартный asn1.Marshal
			stdData, err := asn1.Marshal(OID(scalar).ToASN1())
			if err != nil {
				t.Fatalf("asn1.Marshal: %v", err)
			}

			// Сравниваем
			if !bytes.Equal(ourData, stdData) {
				t.Errorf("MarshalBinary = %x, ожидалось %x", ourData, stdData)
			}
		})
	}
}

// Тест с round trip
func TestScalarOIDMarshalBinaryRoundTrip(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
		{0, 39, 1, 0},
		{1, 3, 6, 1}, // Не скалярный
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Кодируем
			data, err := scalar.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// Декодируем
			var decoded ScalarOID
			if err := decoded.UnmarshalBinary(data); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(scalar) {
				t.Errorf("Round trip: %v -> %x -> %v", scalar, data, decoded)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDMarshalBinaryProperties(t *testing.T) {
	t.Run("Результат начинается с тега OID (0x06)", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
		}

		for _, scalar := range scalars {
			data, err := scalar.MarshalBinary()
			if err != nil {
				t.Errorf("MarshalBinary(%v): %v", scalar, err)
				continue
			}

			if len(data) < 2 {
				t.Error("Данные слишком короткие")
				continue
			}

			if data[0] != 0x06 {
				t.Errorf("Первый байт = 0x%02x, ожидалось 0x06", data[0])
			}
		}
	})

	t.Run("Размер зависит от длины OID", func(t *testing.T) {
		shortOID := ScalarOID{1, 3, 6, 0}
		longOID := ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}

		shortData, _ := shortOID.MarshalBinary()
		longData, _ := longOID.MarshalBinary()

		if len(shortData) >= len(longData) {
			t.Error("Короткий OID должен давать меньше данных")
		}
	})

	t.Run("Эквивалентна OID.MarshalBinary()", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		scalarData, scalarErr := scalar.MarshalBinary()
		oidData, oidErr := OID(scalar).MarshalBinary()

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && !bytes.Equal(scalarData, oidData) {
			t.Error("Результаты должны совпадать")
		}
	})
}

// Тест с подтестами
func TestScalarOIDMarshalBinaryCategories(t *testing.T) {
	t.Run("Успешные случаи", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{2, 100, 3, 0},
		}

		for _, scalar := range scalars {
			data, err := scalar.MarshalBinary()
			if err != nil {
				t.Errorf("MarshalBinary(%v): %v", scalar, err)
				continue
			}
			if len(data) == 0 {
				t.Error("Пустой результат")
			}
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []struct {
			scalar  ScalarOID
			wantErr error
		}{
			{ScalarOID{}, ErrOIDTooShort},
			{ScalarOID{1}, ErrOIDTooShort},
			{ScalarOID{3, 1, 0}, ErrFirstComponentTooBig},
		}

		for _, tt := range tests {
			_, err := tt.scalar.MarshalBinary()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("MarshalBinary(%v) = %v, ожидалось %v", tt.scalar, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_MarshalBinary() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	data, err := scalar.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", data)
	// Output: 06082b06010201010100
}

// Пример с ошибкой
func ExampleScalarOID_MarshalBinary_error() {
	scalar := ScalarOID{}

	_, err := scalar.MarshalBinary()
	fmt.Println(errors.Is(err, ErrOIDTooShort))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDMarshalBinary(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_, _ = scalar.MarshalBinary()
	}
}

// Сравнение с OID.MarshalBinary
func BenchmarkScalarOIDVsOIDMarshalBinary(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")
	oid := OID(scalar)

	b.Run("ScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = scalar.MarshalBinary()
		}
	})

	b.Run("OID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = oid.MarshalBinary()
		}
	})
}

func TestScalarOIDUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected ScalarOID
		wantErr  error
	}{
		{
			name:     "Скалярный OID",
			data:     []byte{0x06, 0x05, 0x2B, 0x06, 0x01, 0x01, 0x00},
			expected: ScalarOID{1, 3, 6, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Длинный скалярный",
			data:     []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00},
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Не скалярный",
			data:     []byte{0x06, 0x04, 0x2B, 0x06, 0x01, 0x01},
			expected: ScalarOID{1, 3, 6, 1, 1},
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			data:     []byte{0x06, 0x04, 0x81, 0x34, 0x03, 0x00},
			expected: ScalarOID{2, 100, 3, 0},
			wantErr:  nil,
		},
		{
			name:     "Пустые данные",
			data:     []byte{},
			expected: nil,
			wantErr:  ErrDataTooShort,
		},
		{
			name:     "Короткие данные",
			data:     []byte{0x06},
			expected: nil,
			wantErr:  ErrDataTooShort,
		},
		{
			name:     "Неверный тег",
			data:     []byte{0x05, 0x00},
			expected: nil,
			wantErr:  ErrInvalidASN1Tag,
		},
		{
			name:     "Неверная длина",
			data:     []byte{0x06, 0x80},
			expected: nil,
			wantErr:  ErrInvalidASN1Length,
		},
		{
			name:     "Недостаточно данных",
			data:     []byte{0x06, 0x05, 0x01},
			expected: nil,
			wantErr:  ErrInsufficientData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var scalar ScalarOID
			err := scalar.UnmarshalBinary(tt.data)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UnmarshalBinary: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UnmarshalBinary = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalBinary: %v", err)
				return
			}

			if !scalar.Equal(tt.expected) {
				t.Errorf("UnmarshalBinary = %v, ожидалось %v", scalar, tt.expected)
			}
		})
	}
}

// Тест с round trip
func TestScalarOIDUnmarshalBinaryRoundTrip(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
		{0, 39, 1, 0},
		{1, 3, 6, 1}, // Не скалярный
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Кодируем
			data, err := scalar.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// Декодируем
			var decoded ScalarOID
			if err := decoded.UnmarshalBinary(data); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(scalar) {
				t.Errorf("Round trip: %v -> %x -> %v", scalar, data, decoded)
			}
		})
	}
}

// Тест с сравнением со стандартной библиотекой
func TestScalarOIDUnmarshalBinaryCompareWithStd(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Кодируем стандартным способом
			stdData, err := asn1.Marshal(OID(scalar).ToASN1())
			if err != nil {
				t.Fatalf("asn1.Marshal: %v", err)
			}

			// Декодируем нашим методом
			var decoded ScalarOID
			if err := decoded.UnmarshalBinary(stdData); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(scalar) {
				t.Errorf("UnmarshalBinary(std) = %v, ожидалось %v", decoded, scalar)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDUnmarshalBinaryProperties(t *testing.T) {
	t.Run("Декодированный OID валиден", func(t *testing.T) {
		original := ScalarOID{1, 3, 6, 1, 0}
		data, _ := original.MarshalBinary()

		var decoded ScalarOID
		if err := decoded.UnmarshalBinary(data); err != nil {
			t.Fatalf("UnmarshalBinary: %v", err)
		}

		if err := decoded.Validate(); err != nil {
			t.Errorf("Validate: %v", err)
		}
	})

	t.Run("Перезаписывает предыдущее значение", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		// Декодируем другой OID
		newData := []byte{0x06, 0x04, 0x2B, 0x06, 0x01, 0x01}
		if err := scalar.UnmarshalBinary(newData); err != nil {
			t.Fatalf("UnmarshalBinary: %v", err)
		}

		if !scalar.Equal(ScalarOID{1, 3, 6, 1, 1}) {
			t.Errorf("После UnmarshalBinary = %v, ожидалось 1.3.6.1.1", scalar)
		}
	})

	t.Run("Эквивалентна OID.UnmarshalBinary()", func(t *testing.T) {
		data := []byte{0x06, 0x05, 0x2B, 0x06, 0x01, 0x01, 0x00}

		var scalar ScalarOID
		var oid OID

		scalarErr := scalar.UnmarshalBinary(data)
		oidErr := oid.UnmarshalBinary(data)

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && !scalar.Equal(ScalarOID(oid)) {
			t.Error("Результаты должны совпадать")
		}
	})
}

// Тест с подтестами
func TestScalarOIDUnmarshalBinaryCategories(t *testing.T) {
	t.Run("Успешные случаи", func(t *testing.T) {
		tests := []struct {
			data     []byte
			expected ScalarOID
		}{
			{[]byte{0x06, 0x05, 0x2B, 0x06, 0x01, 0x01, 0x00}, ScalarOID{1, 3, 6, 1, 1, 0}},
			{[]byte{0x06, 0x04, 0x2B, 0x06, 0x01, 0x01}, ScalarOID{1, 3, 6, 1, 1}},
		}

		for _, tt := range tests {
			var scalar ScalarOID
			if err := scalar.UnmarshalBinary(tt.data); err != nil {
				t.Errorf("UnmarshalBinary(%x): %v", tt.data, err)
			}
			if !scalar.Equal(tt.expected) {
				t.Errorf("UnmarshalBinary(%x) = %v, ожидалось %v",
					tt.data, scalar, tt.expected)
			}
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []struct {
			data    []byte
			wantErr error
		}{
			{[]byte{}, ErrDataTooShort},
			{[]byte{0x06}, ErrDataTooShort},
			{[]byte{0x05, 0x00}, ErrInvalidASN1Tag},
			{[]byte{0x06, 0x80}, ErrInvalidASN1Length},
		}

		for _, tt := range tests {
			var scalar ScalarOID
			err := scalar.UnmarshalBinary(tt.data)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UnmarshalBinary(%x) = %v, ожидалось %v",
					tt.data, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_UnmarshalBinary() {
	data := []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00}

	var scalar ScalarOID
	if err := scalar.UnmarshalBinary(data); err != nil {
		panic(err)
	}

	fmt.Println(scalar)
	// Output: 1.3.6.1.2.1.1.1.0
}

// Пример с ошибкой
func ExampleScalarOID_UnmarshalBinary_error() {
	var scalar ScalarOID
	err := scalar.UnmarshalBinary([]byte{})
	fmt.Println(errors.Is(err, ErrDataTooShort))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDUnmarshalBinary(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")
	data, _ := scalar.MarshalBinary()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var decoded ScalarOID
		_ = decoded.UnmarshalBinary(data)
	}
}

// Сравнение с OID.UnmarshalBinary
func BenchmarkScalarOIDVsOIDUnmarshalBinary(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")
	data, _ := scalar.MarshalBinary()

	b.Run("ScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var decoded ScalarOID
			_ = decoded.UnmarshalBinary(data)
		}
	})

	b.Run("OID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var decoded OID
			_ = decoded.UnmarshalBinary(data)
		}
	})
}

func TestScalarOIDMarshalBER(t *testing.T) {
	tests := []struct {
		name    string
		scalar  ScalarOID
		wantErr error
	}{
		{
			name:    "Скалярный",
			scalar:  ScalarOID{1, 3, 6, 1, 0},
			wantErr: nil,
		},
		{
			name:    "Длинный скалярный",
			scalar:  ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr: nil,
		},
		{
			name:    "Не скалярный",
			scalar:  ScalarOID{1, 3, 6, 1},
			wantErr: nil,
		},
		{
			name:    "С первым 2",
			scalar:  ScalarOID{2, 100, 3, 0},
			wantErr: nil,
		},
		{
			name:    "С первым 0",
			scalar:  ScalarOID{0, 39, 1, 0},
			wantErr: nil,
		},
		{
			name:    "Пустой",
			scalar:  ScalarOID{},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Nil",
			scalar:  nil,
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Один компонент",
			scalar:  ScalarOID{1},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Невалидный",
			scalar:  ScalarOID{3, 1, 0},
			wantErr: ErrFirstComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.scalar.MarshalBER()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("MarshalBER: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("MarshalBER = %v, ожидалось %v", err, tt.wantErr)
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

// Тест с round trip
func TestScalarOIDMarshalBERRoundTrip(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
		{0, 39, 1, 0},
		{1, 3, 6, 1}, // Не скалярный
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Кодируем
			data, err := scalar.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}

			// Декодируем
			var decoded ScalarOID
			if err := decoded.UnmarshalBER(data); err != nil {
				t.Fatalf("UnmarshalBER: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(scalar) {
				t.Errorf("Round trip: %v -> %x -> %v", scalar, data, decoded)
			}
		})
	}
}

// Тест с сравнением MarshalBER и MarshalBinary
func TestScalarOIDMarshalBERCompareWithBinary(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			berData, berErr := scalar.MarshalBER()
			binData, binErr := scalar.MarshalBinary()

			if (berErr == nil) != (binErr == nil) {
				t.Error("Ошибки должны совпадать")
			}

			if berErr == nil && !bytes.Equal(berData, binData) {
				t.Errorf("BER = %x, Binary = %x", berData, binData)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDMarshalBERProperties(t *testing.T) {
	t.Run("Результат начинается с тега OID (0x06)", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
		}

		for _, scalar := range scalars {
			data, err := scalar.MarshalBER()
			if err != nil {
				t.Errorf("MarshalBER(%v): %v", scalar, err)
				continue
			}

			if len(data) < 2 {
				t.Error("Данные слишком короткие")
				continue
			}

			if data[0] != 0x06 {
				t.Errorf("Первый байт = 0x%02x, ожидалось 0x06", data[0])
			}
		}
	})

	t.Run("Эквивалентна OID.MarshalBER()", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		scalarData, scalarErr := scalar.MarshalBER()
		oidData, oidErr := OID(scalar).MarshalBER()

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && !bytes.Equal(scalarData, oidData) {
			t.Error("Результаты должны совпадать")
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}
		scalarCopy := make(ScalarOID, len(scalar))
		copy(scalarCopy, scalar)

		scalar.MarshalBER()

		if !scalar.Equal(scalarCopy) {
			t.Error("MarshalBER() не должен изменять OID")
		}
	})
}

// Тест с подтестами
func TestScalarOIDMarshalBERCategories(t *testing.T) {
	t.Run("Успешные случаи", func(t *testing.T) {
		scalars := []ScalarOID{
			{1, 3, 6, 1, 0},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{2, 100, 3, 0},
			{0, 39, 1, 0},
		}

		for _, scalar := range scalars {
			data, err := scalar.MarshalBER()
			if err != nil {
				t.Errorf("MarshalBER(%v): %v", scalar, err)
				continue
			}
			if len(data) == 0 {
				t.Error("Пустой результат")
			}
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []struct {
			scalar  ScalarOID
			wantErr error
		}{
			{ScalarOID{}, ErrOIDTooShort},
			{ScalarOID{1}, ErrOIDTooShort},
			{ScalarOID{3, 1, 0}, ErrFirstComponentTooBig},
		}

		for _, tt := range tests {
			_, err := tt.scalar.MarshalBER()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("MarshalBER(%v) = %v, ожидалось %v", tt.scalar, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_MarshalBER() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	data, err := scalar.MarshalBER()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", data)
	// Output: 06082b06010201010100
}

// Пример с ошибкой
func ExampleScalarOID_MarshalBER_error() {
	scalar := ScalarOID{}

	_, err := scalar.MarshalBER()
	fmt.Println(errors.Is(err, ErrOIDTooShort))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDMarshalBER(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_, _ = scalar.MarshalBER()
	}
}

// Сравнение с OID.MarshalBER
func BenchmarkScalarOIDVsOIDMarshalBER(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")
	oid := OID(scalar)

	b.Run("ScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = scalar.MarshalBER()
		}
	})

	b.Run("OID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = oid.MarshalBER()
		}
	})
}

func TestScalarOIDUnmarshalBER(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected ScalarOID
		wantErr  error
	}{
		{
			name:     "Скалярный OID",
			data:     []byte{0x06, 0x05, 0x2B, 0x06, 0x01, 0x01, 0x00},
			expected: ScalarOID{1, 3, 6, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Длинный скалярный",
			data:     []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00},
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Не скалярный",
			data:     []byte{0x06, 0x04, 0x2B, 0x06, 0x01, 0x01},
			expected: ScalarOID{1, 3, 6, 1, 1},
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			data:     []byte{0x06, 0x04, 0x81, 0x34, 0x03, 0x00},
			expected: ScalarOID{2, 100, 3, 0},
			wantErr:  nil,
		},
		{
			name:     "Пустые данные",
			data:     []byte{},
			expected: nil,
			wantErr:  ErrInsufficientData,
		},
		{
			name:     "Короткие данные",
			data:     []byte{0x06, 0x01},
			expected: nil,
			wantErr:  ErrInsufficientData,
		},
		{
			name:     "Неверный тег",
			data:     []byte{0x05, 0x01, 0x2B},
			expected: nil,
			wantErr:  ErrInvalidASN1Tag,
		},
		{
			name:     "Неверная длина",
			data:     []byte{0x06, 0x80, 0x00},
			expected: nil,
			wantErr:  ErrInvalidLength,
		},
		{
			name:     "Пустой контент",
			data:     []byte{0x06, 0x00},
			expected: nil,
			wantErr:  ErrEmptyContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var scalar ScalarOID
			err := scalar.UnmarshalBER(tt.data)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UnmarshalBER: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UnmarshalBER = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalBER: %v", err)
				return
			}

			if !scalar.Equal(tt.expected) {
				t.Errorf("UnmarshalBER = %v, ожидалось %v", scalar, tt.expected)
			}
		})
	}
}

// Тест с round trip
func TestScalarOIDUnmarshalBERRoundTrip(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
		{0, 39, 1, 0},
		{1, 3, 6, 1}, // Не скалярный
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Кодируем
			data, err := scalar.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}

			// Декодируем
			var decoded ScalarOID
			if err := decoded.UnmarshalBER(data); err != nil {
				t.Fatalf("UnmarshalBER: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(scalar) {
				t.Errorf("Round trip: %v -> %x -> %v", scalar, data, decoded)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDUnmarshalBERProperties(t *testing.T) {
	t.Run("Декодированный OID валиден", func(t *testing.T) {
		original := ScalarOID{1, 3, 6, 1, 0}
		data, _ := original.MarshalBER()

		var decoded ScalarOID
		if err := decoded.UnmarshalBER(data); err != nil {
			t.Fatalf("UnmarshalBER: %v", err)
		}

		if err := decoded.Validate(); err != nil {
			t.Errorf("Validate: %v", err)
		}
	})

	t.Run("Перезаписывает предыдущее значение", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		// Декодируем другой OID
		newData := []byte{0x06, 0x04, 0x2B, 0x06, 0x01, 0x01}
		if err := scalar.UnmarshalBER(newData); err != nil {
			t.Fatalf("UnmarshalBER: %v", err)
		}

		if !scalar.Equal(ScalarOID{1, 3, 6, 1, 1}) {
			t.Errorf("После UnmarshalBER = %v, ожидалось 1.3.6.1.1", scalar)
		}
	})

	t.Run("Эквивалентна OID.UnmarshalBER()", func(t *testing.T) {
		data := []byte{0x06, 0x05, 0x2B, 0x06, 0x01, 0x01, 0x00}

		var scalar ScalarOID
		var oid OID

		scalarErr := scalar.UnmarshalBER(data)
		oidErr := oid.UnmarshalBER(data)

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && !scalar.Equal(ScalarOID(oid)) {
			t.Error("Результаты должны совпадать")
		}
	})

	t.Run("UnmarshalBER и UnmarshalBinary дают одинаковый результат", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		derData, _ := scalar.MarshalBinary()
		berData, _ := scalar.MarshalBER()

		var decodedDER ScalarOID
		var decodedBER ScalarOID

		decodedDER.UnmarshalBinary(derData)
		decodedBER.UnmarshalBER(berData)

		if !decodedDER.Equal(decodedBER) {
			t.Error("DER и BER должны давать одинаковый результат")
		}
	})
}

// Тест с подтестами
func TestScalarOIDUnmarshalBERCategories(t *testing.T) {
	t.Run("Успешные случаи", func(t *testing.T) {
		tests := []struct {
			data     []byte
			expected ScalarOID
		}{
			{[]byte{0x06, 0x05, 0x2B, 0x06, 0x01, 0x01, 0x00}, ScalarOID{1, 3, 6, 1, 1, 0}},
			{[]byte{0x06, 0x04, 0x2B, 0x06, 0x01, 0x01}, ScalarOID{1, 3, 6, 1, 1}},
		}

		for _, tt := range tests {
			var scalar ScalarOID
			if err := scalar.UnmarshalBER(tt.data); err != nil {
				t.Errorf("UnmarshalBER(%x): %v", tt.data, err)
			}
			if !scalar.Equal(tt.expected) {
				t.Errorf("UnmarshalBER(%x) = %v, ожидалось %v",
					tt.data, scalar, tt.expected)
			}
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []struct {
			data    []byte
			wantErr error
		}{
			{[]byte{}, ErrInsufficientData},
			{[]byte{0x06}, ErrInsufficientData},
			{[]byte{0x06, 0x01}, ErrInsufficientData},
			{[]byte{0x05, 0x01, 0x2B}, ErrInvalidASN1Tag},
			{[]byte{0x06, 0x80, 0x00}, ErrInvalidLength},
			{[]byte{0x06, 0x00}, ErrEmptyContent},
		}

		for _, tt := range tests {
			var scalar ScalarOID
			err := scalar.UnmarshalBER(tt.data)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UnmarshalBER(%x) = %v, ожидалось %v",
					tt.data, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_UnmarshalBER() {
	data := []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00}

	var scalar ScalarOID
	if err := scalar.UnmarshalBER(data); err != nil {
		panic(err)
	}

	fmt.Println(scalar)
	// Output: 1.3.6.1.2.1.1.1.0
}

// Пример с ошибкой
func ExampleScalarOID_UnmarshalBER_error() {
	var scalar ScalarOID
	err := scalar.UnmarshalBER([]byte{})
	fmt.Println(errors.Is(err, ErrInsufficientData))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDUnmarshalBER(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")
	data, _ := scalar.MarshalBER()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var decoded ScalarOID
		_ = decoded.UnmarshalBER(data)
	}
}

// Сравнение с OID.UnmarshalBER
func BenchmarkScalarOIDVsOIDUnmarshalBER(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")
	data, _ := scalar.MarshalBER()

	b.Run("ScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var decoded ScalarOID
			_ = decoded.UnmarshalBER(data)
		}
	})

	b.Run("OID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var decoded OID
			_ = decoded.UnmarshalBER(data)
		}
	})
}

func TestScalarOIDMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		expected string
		wantErr  error
	}{
		{
			name:     "Скалярный",
			scalar:   ScalarOID{1, 3, 6, 1, 0},
			expected: `"1.3.6.1.0"`,
			wantErr:  nil,
		},
		{
			name:     "Длинный скалярный",
			scalar:   ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: `"1.3.6.1.2.1.1.1.0"`,
			wantErr:  nil,
		},
		{
			name:     "Не скалярный",
			scalar:   ScalarOID{1, 3, 6, 1},
			expected: `"1.3.6.1"`,
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			scalar:   ScalarOID{2, 100, 3, 0},
			expected: `"2.100.3.0"`,
			wantErr:  nil,
		},
		{
			name:     "С первым 0",
			scalar:   ScalarOID{0, 39, 1, 0},
			expected: `"0.39.1.0"`,
			wantErr:  nil,
		},
		{
			name:     "Пустой",
			scalar:   ScalarOID{},
			expected: `""`,
			wantErr:  nil,
		},
		{
			name:     "Nil",
			scalar:   nil,
			expected: `""`,
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.scalar.MarshalJSON()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("MarshalJSON: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("MarshalJSON = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("MarshalJSON: %v", err)
				return
			}

			if string(data) != tt.expected {
				t.Errorf("MarshalJSON = %s, ожидалось %s", data, tt.expected)
			}
		})
	}
}

// Тест с round trip через json.Marshal
func TestScalarOIDMarshalJSONRoundTrip(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
		{1, 3, 6, 1}, // Не скалярный
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Используем json.Marshal
			data, err := json.Marshal(scalar)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			// Декодируем
			var decoded ScalarOID
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(scalar) {
				t.Errorf("Round trip: %v -> %s -> %v", scalar, data, decoded)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDMarshalJSONProperties(t *testing.T) {
	t.Run("Возвращает валидный JSON", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		data, err := scalar.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}

		if !json.Valid(data) {
			t.Errorf("MarshalJSON = %s, невалидный JSON", data)
		}
	})

	t.Run("Эквивалентна OID.MarshalJSON()", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		scalarData, scalarErr := scalar.MarshalJSON()
		oidData, oidErr := OID(scalar).MarshalJSON()

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && !bytes.Equal(scalarData, oidData) {
			t.Error("Результаты должны совпадать")
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}
		scalarCopy := make(ScalarOID, len(scalar))
		copy(scalarCopy, scalar)

		scalar.MarshalJSON()

		if !scalar.Equal(scalarCopy) {
			t.Error("MarshalJSON() не должен изменять OID")
		}
	})
}

// Тест с подтестами
func TestScalarOIDMarshalJSONCategories(t *testing.T) {
	t.Run("Скалярные OID", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected string
		}{
			{ScalarOID{1, 3, 6, 1, 0}, `"1.3.6.1.0"`},
			{ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}, `"1.3.6.1.2.1.1.1.0"`},
		}

		for _, tt := range tests {
			data, err := tt.scalar.MarshalJSON()
			if err != nil {
				t.Errorf("MarshalJSON(%v): %v", tt.scalar, err)
				continue
			}
			if string(data) != tt.expected {
				t.Errorf("MarshalJSON(%v) = %s, ожидалось %s",
					tt.scalar, data, tt.expected)
			}
		}
	})

	t.Run("Не скалярные OID", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected string
		}{
			{ScalarOID{1, 3, 6, 1}, `"1.3.6.1"`},
			{ScalarOID{2, 100, 3}, `"2.100.3"`},
		}

		for _, tt := range tests {
			data, err := tt.scalar.MarshalJSON()
			if err != nil {
				t.Errorf("MarshalJSON(%v): %v", tt.scalar, err)
				continue
			}
			if string(data) != tt.expected {
				t.Errorf("MarshalJSON(%v) = %s, ожидалось %s",
					tt.scalar, data, tt.expected)
			}
		}
	})

	t.Run("Пустые OID", func(t *testing.T) {
		tests := []ScalarOID{
			{},
			nil,
		}

		for _, scalar := range tests {
			data, err := scalar.MarshalJSON()
			if err != nil {
				t.Errorf("MarshalJSON(%v): %v", scalar, err)
				continue
			}
			if string(data) != `""` {
				t.Errorf("MarshalJSON(%v) = %s, ожидалось '\"\"'", scalar, data)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_MarshalJSON() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	data, err := scalar.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output: "1.3.6.1.2.1.1.1.0"
}

// Пример с пустым OID
func ExampleScalarOID_MarshalJSON_empty() {
	scalar := ScalarOID{}

	data, _ := scalar.MarshalJSON()

	fmt.Println(string(data))
	// Output: ""
}

// Бенчмарк
func BenchmarkScalarOIDMarshalJSON(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_, _ = scalar.MarshalJSON()
	}
}

// Сравнение с OID.MarshalJSON
func BenchmarkScalarOIDVsOIDMarshalJSON(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")
	oid := OID(scalar)

	b.Run("ScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = scalar.MarshalJSON()
		}
	})

	b.Run("OID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = oid.MarshalJSON()
		}
	})
}

func TestScalarOIDUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected ScalarOID
		wantErr  error
	}{
		{
			name:     "Скалярный OID",
			data:     []byte(`"1.3.6.1.2.1.1.1.0"`),
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Не скалярный OID",
			data:     []byte(`"1.3.6.1.2.1.1.1"`),
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1},
			wantErr:  nil,
		},
		{
			name:     "Короткий OID",
			data:     []byte(`"1.3.6.1"`),
			expected: ScalarOID{1, 3, 6, 1},
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			data:     []byte(`"2.100.3.0"`),
			expected: ScalarOID{2, 100, 3, 0},
			wantErr:  nil,
		},
		{
			name:     "Пустая строка JSON",
			data:     []byte(`""`),
			expected: nil,
			wantErr:  ErrEmptyOID,
		},
		{
			name:     "Null",
			data:     []byte(`null`),
			expected: nil,
			wantErr:  ErrInvalidJSONType,
		},
		{
			name:     "Число",
			data:     []byte(`123`),
			expected: nil,
			wantErr:  ErrInvalidJSONType,
		},
		{
			name:     "Объект",
			data:     []byte(`{"oid":"1.3.6.1"}`),
			expected: nil,
			wantErr:  ErrInvalidJSONType,
		},
		{
			name:     "Массив",
			data:     []byte(`["1.3.6.1"]`),
			expected: nil,
			wantErr:  ErrInvalidJSONType,
		},
		{
			name:     "Невалидный JSON",
			data:     []byte(`invalid`),
			expected: nil,
			wantErr:  ErrInvalidJSONType,
		},
		{
			name:     "Невалидный OID",
			data:     []byte(`"invalid"`),
			expected: nil,
			wantErr:  nil, // Любая ошибка парсинга
		},
		{
			name:     "Один компонент",
			data:     []byte(`"1"`),
			expected: nil,
			wantErr:  ErrOIDTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var scalar ScalarOID
			err := scalar.UnmarshalJSON(tt.data)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UnmarshalJSON: ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UnmarshalJSON = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if tt.expected == nil && tt.data != nil {
				// Ожидаем ошибку для невалидного ввода
				if err == nil {
					t.Errorf("UnmarshalJSON(%s): ожидалась ошибка", tt.data)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalJSON: %v", err)
				return
			}

			if !scalar.Equal(tt.expected) {
				t.Errorf("UnmarshalJSON = %v, ожидалось %v", scalar, tt.expected)
			}
		})
	}
}

// Тест с round trip через json.Marshal/Unmarshal
func TestScalarOIDUnmarshalJSONRoundTrip(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
		{0, 39, 1, 0},
		{1, 3, 6, 1}, // Не скалярный
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Кодируем через MarshalJSON
			data, err := scalar.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}

			// Декодируем через UnmarshalJSON
			var decoded ScalarOID
			if err := decoded.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(scalar) {
				t.Errorf("Round trip: %v -> %s -> %v", scalar, data, decoded)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDUnmarshalJSONProperties(t *testing.T) {
	t.Run("Декодированный OID валиден", func(t *testing.T) {
		var scalar ScalarOID
		if err := scalar.UnmarshalJSON([]byte(`"1.3.6.1.0"`)); err != nil {
			t.Fatalf("UnmarshalJSON: %v", err)
		}

		if err := scalar.Validate(); err != nil {
			t.Errorf("Validate: %v", err)
		}
	})

	t.Run("Перезаписывает предыдущее значение", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		// Декодируем другой OID
		if err := scalar.UnmarshalJSON([]byte(`"2.100.3.0"`)); err != nil {
			t.Fatalf("UnmarshalJSON: %v", err)
		}

		if !scalar.Equal(ScalarOID{2, 100, 3, 0}) {
			t.Errorf("После UnmarshalJSON = %v, ожидалось 2.100.3.0", scalar)
		}
	})

	t.Run("Эквивалентна OID.UnmarshalJSON()", func(t *testing.T) {
		data := []byte(`"1.3.6.1.0"`)

		var scalar ScalarOID
		var oid OID

		scalarErr := scalar.UnmarshalJSON(data)
		oidErr := oid.UnmarshalJSON(data)

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && !scalar.Equal(ScalarOID(oid)) {
			t.Error("Результаты должны совпадать")
		}
	})
}

// Тест с подтестами
func TestScalarOIDUnmarshalJSONCategories(t *testing.T) {
	t.Run("Успешные случаи", func(t *testing.T) {
		tests := []struct {
			data     []byte
			expected ScalarOID
		}{
			{[]byte(`"1.3.6.1.0"`), ScalarOID{1, 3, 6, 1, 0}},
			{[]byte(`"1.3.6.1.2.1.1.1.0"`), ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}},
			{[]byte(`"2.100.3.0"`), ScalarOID{2, 100, 3, 0}},
		}

		for _, tt := range tests {
			var scalar ScalarOID
			if err := scalar.UnmarshalJSON(tt.data); err != nil {
				t.Errorf("UnmarshalJSON(%s): %v", tt.data, err)
			}
			if !scalar.Equal(tt.expected) {
				t.Errorf("UnmarshalJSON(%s) = %v, ожидалось %v",
					tt.data, scalar, tt.expected)
			}
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []struct {
			data    []byte
			wantErr error
		}{
			{[]byte(`""`), ErrEmptyOID},
			{[]byte(`null`), ErrInvalidJSONType},
			{[]byte(`123`), ErrInvalidJSONType},
			{[]byte(`{"oid":"1.3.6.1"}`), ErrInvalidJSONType},
			{[]byte(`invalid`), ErrInvalidJSONType},
			{[]byte(`"1"`), ErrOIDTooShort},
		}

		for _, tt := range tests {
			var scalar ScalarOID
			err := scalar.UnmarshalJSON(tt.data)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UnmarshalJSON(%s) = %v, ожидалось %v",
					tt.data, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_UnmarshalJSON() {
	data := []byte(`"1.3.6.1.2.1.1.1.0"`)

	var scalar ScalarOID
	if err := scalar.UnmarshalJSON(data); err != nil {
		panic(err)
	}

	fmt.Println(scalar)
	// Output: 1.3.6.1.2.1.1.1.0
}

// Пример с ошибкой
func ExampleScalarOID_UnmarshalJSON_error() {
	var scalar ScalarOID
	err := scalar.UnmarshalJSON([]byte(`null`))
	fmt.Println(errors.Is(err, ErrInvalidJSONType))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDUnmarshalJSON(b *testing.B) {
	data := []byte(`"1.3.6.1.2.1.1.1.0"`)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var scalar ScalarOID
		_ = scalar.UnmarshalJSON(data)
	}
}

// Сравнение с OID.UnmarshalJSON
func BenchmarkScalarOIDVsOIDUnmarshalJSON(b *testing.B) {
	data := []byte(`"1.3.6.1.2.1.1.1.0"`)

	b.Run("ScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var scalar ScalarOID
			_ = scalar.UnmarshalJSON(data)
		}
	})

	b.Run("OID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var oid OID
			_ = oid.UnmarshalJSON(data)
		}
	})
}

func TestScalarOIDValue(t *testing.T) {
	tests := []struct {
		name     string
		scalar   ScalarOID
		expected driver.Value
		wantErr  error
	}{
		{
			name:     "Скалярный OID",
			scalar:   ScalarOID{1, 3, 6, 1, 0},
			expected: "1.3.6.1.0",
			wantErr:  nil,
		},
		{
			name:     "Длинный скалярный",
			scalar:   ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: "1.3.6.1.2.1.1.1.0",
			wantErr:  nil,
		},
		{
			name:     "Не скалярный",
			scalar:   ScalarOID{1, 3, 6, 1},
			expected: "1.3.6.1",
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			scalar:   ScalarOID{2, 100, 3, 0},
			expected: "2.100.3.0",
			wantErr:  nil,
		},
		{
			name:     "Пустой",
			scalar:   ScalarOID{},
			expected: nil,
			wantErr:  nil,
		},
		{
			name:     "Nil",
			scalar:   nil,
			expected: nil,
			wantErr:  nil,
		},
		{
			name:     "Один компонент",
			scalar:   ScalarOID{1},
			expected: nil,
			wantErr:  ErrOIDTooShort,
		},
		{
			name:     "Невалидный",
			scalar:   ScalarOID{3, 1, 0},
			expected: nil,
			wantErr:  ErrFirstComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.scalar.Value()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Value(): ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Value() = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Value(): %v", err)
				return
			}

			if value != tt.expected {
				t.Errorf("Value() = %v, ожидалось %v", value, tt.expected)
			}
		})
	}
}

// Тест с проверкой типов
func TestScalarOIDValueTypes(t *testing.T) {
	t.Run("Возвращает string", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		value, err := scalar.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if _, ok := value.(string); !ok {
			t.Errorf("Value() тип = %T, ожидался string", value)
		}
	})

	t.Run("Возвращает nil для пустого", func(t *testing.T) {
		scalar := ScalarOID{}

		value, err := scalar.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if value != nil {
			t.Errorf("Value() = %v, ожидался nil", value)
		}
	})
}

// Тест с проверкой driver.Valuer интерфейса
func TestScalarOIDImplementsValuer(t *testing.T) {
	var _ driver.Valuer = ScalarOID{1, 3, 6, 1, 0}

	scalar := ScalarOID{1, 3, 6, 1, 0}

	value, err := scalar.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	if value == nil {
		t.Error("Value не должен быть nil")
	}
}

// Тест с round trip (Value + Scan)
func TestScalarOIDValueRoundTrip(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Получаем значение для БД
			value, err := scalar.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			// Сканируем обратно
			var decoded ScalarOID
			if err := decoded.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(scalar) {
				t.Errorf("Round trip: %v -> %v -> %v", scalar, value, decoded)
			}
		})
	}
}

// Тест с проверкой свойств
func TestScalarOIDValueProperties(t *testing.T) {
	t.Run("Эквивалентна OID.Value()", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		scalarValue, scalarErr := scalar.Value()
		oidValue, oidErr := OID(scalar).Value()

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && scalarValue != oidValue {
			t.Error("Результаты должны совпадать")
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}
		scalarCopy := make(ScalarOID, len(scalar))
		copy(scalarCopy, scalar)

		scalar.Value()

		if !scalar.Equal(scalarCopy) {
			t.Error("Value() не должен изменять OID")
		}
	})
}

// Тест с подтестами
func TestScalarOIDValueCategories(t *testing.T) {
	t.Run("Валидные OID", func(t *testing.T) {
		tests := []struct {
			scalar   ScalarOID
			expected string
		}{
			{ScalarOID{1, 3, 6, 1, 0}, "1.3.6.1.0"},
			{ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}, "1.3.6.1.2.1.1.1.0"},
			{ScalarOID{2, 100, 3, 0}, "2.100.3.0"},
		}

		for _, tt := range tests {
			value, err := tt.scalar.Value()
			if err != nil {
				t.Errorf("Value(%v): %v", tt.scalar, err)
				continue
			}
			if value != tt.expected {
				t.Errorf("Value(%v) = %v, ожидалось %v", tt.scalar, value, tt.expected)
			}
		}
	})

	t.Run("Пустые OID", func(t *testing.T) {
		tests := []ScalarOID{
			{},
			nil,
		}

		for _, scalar := range tests {
			value, err := scalar.Value()
			if err != nil {
				t.Errorf("Value(%v): %v", scalar, err)
				continue
			}
			if value != nil {
				t.Errorf("Value(%v) = %v, ожидался nil", scalar, value)
			}
		}
	})

	t.Run("Невалидные OID", func(t *testing.T) {
		tests := []struct {
			scalar  ScalarOID
			wantErr error
		}{
			{ScalarOID{1}, ErrOIDTooShort},
			{ScalarOID{3, 1, 0}, ErrFirstComponentTooBig},
		}

		for _, tt := range tests {
			_, err := tt.scalar.Value()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Value(%v) = %v, ожидалось %v", tt.scalar, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_Value() {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	value, err := scalar.Value()
	if err != nil {
		panic(err)
	}

	fmt.Println(value)
	// Output: 1.3.6.1.2.1.1.1.0
}

// Пример с пустым OID
func ExampleScalarOID_Value_empty() {
	scalar := ScalarOID{}

	value, err := scalar.Value()
	if err != nil {
		panic(err)
	}

	fmt.Println(value == nil)
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDValue(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")

	b.ReportAllocs()
	for b.Loop() {
		_, _ = scalar.Value()
	}
}

// Сравнение с OID.Value
func BenchmarkScalarOIDVsOIDValue(b *testing.B) {
	scalar := MustScalarOID("1.3.6.1.2.1.1.1.0")
	oid := OID(scalar)

	b.Run("ScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = scalar.Value()
		}
	})

	b.Run("OID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = oid.Value()
		}
	})
}

func TestScalarOIDScan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected ScalarOID
		wantErr  error
	}{
		{
			name:     "Строка",
			input:    "1.3.6.1.2.1.1.1.0",
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Байты",
			input:    []byte("1.3.6.1.2.1.1.1.0"),
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "Не скалярный",
			input:    "1.3.6.1.2.1.1.1",
			expected: ScalarOID{1, 3, 6, 1, 2, 1, 1, 1},
			wantErr:  nil,
		},
		{
			name:     "Короткий",
			input:    "1.3.6.1",
			expected: ScalarOID{1, 3, 6, 1},
			wantErr:  nil,
		},
		{
			name:     "NULL",
			input:    nil,
			expected: nil,
			wantErr:  nil,
		},
		{
			name:     "Число",
			input:    123,
			expected: nil,
			wantErr:  ErrUnsupportedScanType,
		},
		{
			name:     "Булево",
			input:    true,
			expected: nil,
			wantErr:  ErrUnsupportedScanType,
		},
		{
			name:     "Слайс int",
			input:    []int{1, 3, 6},
			expected: nil,
			wantErr:  ErrUnsupportedScanType,
		},
		{
			name:     "Невалидная строка",
			input:    "invalid",
			expected: nil,
			wantErr:  nil, // Любая ошибка парсинга
		},
		{
			name:     "Пустая строка",
			input:    "",
			expected: nil,
			wantErr:  ErrEmptyOID,
		},
		{
			name:     "Один компонент",
			input:    "1",
			expected: nil,
			wantErr:  ErrOIDTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var scalar ScalarOID
			err := scalar.Scan(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Scan(): ожидалась ошибка %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Scan() = %v, ожидалось %v", err, tt.wantErr)
				}
				return
			}

			if tt.expected == nil && tt.input != nil {
				// Ожидаем ошибку для невалидного ввода
				if err == nil {
					t.Errorf("Scan(%v): ожидалась ошибка", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Scan(): %v", err)
				return
			}

			if !scalar.Equal(tt.expected) {
				t.Errorf("Scan() = %v, ожидалось %v", scalar, tt.expected)
			}
		})
	}
}

// Тест с round trip (Value + Scan)
func TestScalarOIDScanRoundTrip(t *testing.T) {
	tests := []ScalarOID{
		{1, 3, 6, 1, 0},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3, 0},
		{0, 39, 1, 0},
	}

	for _, scalar := range tests {
		t.Run(scalar.String(), func(t *testing.T) {
			// Получаем значение для БД
			value, err := scalar.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			// Сканируем обратно
			var decoded ScalarOID
			if err := decoded.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(scalar) {
				t.Errorf("Round trip: %v -> %v -> %v", scalar, value, decoded)
			}
		})
	}
}

// Тест с проверкой sql.Scanner интерфейса
func TestScalarOIDImplementsScanner(t *testing.T) {
	var _ sql.Scanner = (*ScalarOID)(nil)

	var scalar ScalarOID
	if err := scalar.Scan("1.3.6.1.0"); err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if !scalar.Equal(ScalarOID{1, 3, 6, 1, 0}) {
		t.Errorf("Scan = %v, ожидалось 1.3.6.1.0", scalar)
	}
}

// Тест с проверкой свойств
func TestScalarOIDScanProperties(t *testing.T) {
	t.Run("Scan очищает предыдущее значение при NULL", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		if err := scalar.Scan(nil); err != nil {
			t.Fatalf("Scan(nil): %v", err)
		}

		if len(scalar) != 0 {
			t.Errorf("После Scan(nil) длина = %d, ожидалось 0", len(scalar))
		}
	})

	t.Run("Scan перезаписывает предыдущее значение", func(t *testing.T) {
		scalar := ScalarOID{1, 3, 6, 1, 0}

		if err := scalar.Scan("2.100.3.0"); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		if !scalar.Equal(ScalarOID{2, 100, 3, 0}) {
			t.Errorf("Scan = %v, ожидалось 2.100.3.0", scalar)
		}
	})

	t.Run("Эквивалентна OID.Scan()", func(t *testing.T) {
		input := "1.3.6.1.0"

		var scalar ScalarOID
		var oid OID

		scalarErr := scalar.Scan(input)
		oidErr := oid.Scan(input)

		if (scalarErr == nil) != (oidErr == nil) {
			t.Error("Ошибки должны совпадать")
		}

		if scalarErr == nil && !scalar.Equal(ScalarOID(oid)) {
			t.Error("Результаты должны совпадать")
		}
	})
}

// Тест с подтестами
func TestScalarOIDScanCategories(t *testing.T) {
	t.Run("Успешные случаи", func(t *testing.T) {
		tests := []struct {
			input    any
			expected ScalarOID
		}{
			{"1.3.6.1.0", ScalarOID{1, 3, 6, 1, 0}},
			{[]byte("1.3.6.1.0"), ScalarOID{1, 3, 6, 1, 0}},
			{"1.3.6.1.2.1.1.1.0", ScalarOID{1, 3, 6, 1, 2, 1, 1, 1, 0}},
		}

		for _, tt := range tests {
			var scalar ScalarOID
			if err := scalar.Scan(tt.input); err != nil {
				t.Errorf("Scan(%v): %v", tt.input, err)
			}
			if !scalar.Equal(tt.expected) {
				t.Errorf("Scan(%v) = %v, ожидалось %v", tt.input, scalar, tt.expected)
			}
		}
	})

	t.Run("NULL", func(t *testing.T) {
		var scalar ScalarOID
		if err := scalar.Scan(nil); err != nil {
			t.Errorf("Scan(nil): %v", err)
		}
		if len(scalar) != 0 {
			t.Errorf("После Scan(nil) длина = %d, ожидалось 0", len(scalar))
		}
	})

	t.Run("Ошибки", func(t *testing.T) {
		tests := []struct {
			input   any
			wantErr error
		}{
			{123, ErrUnsupportedScanType},
			{true, ErrUnsupportedScanType},
			{[]int{1, 2}, ErrUnsupportedScanType},
			{"", ErrEmptyOID},
			{"1", ErrOIDTooShort},
		}

		for _, tt := range tests {
			var scalar ScalarOID
			err := scalar.Scan(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Scan(%v) = %v, ожидалось %v", tt.input, err, tt.wantErr)
			}
		}
	})
}

// Пример использования
func ExampleScalarOID_Scan() {
	var scalar ScalarOID

	// Симуляция чтения из БД
	if err := scalar.Scan("1.3.6.1.2.1.1.1.0"); err != nil {
		panic(err)
	}

	fmt.Println(scalar)
	// Output: 1.3.6.1.2.1.1.1.0
}

// Пример с NULL
func ExampleScalarOID_Scan_null() {
	var scalar ScalarOID

	// Симуляция чтения NULL из БД
	if err := scalar.Scan(nil); err != nil {
		panic(err)
	}

	fmt.Println(len(scalar) == 0)
	// Output: true
}

// Пример с ошибкой
func ExampleScalarOID_Scan_error() {
	var scalar ScalarOID
	err := scalar.Scan(123)
	fmt.Println(errors.Is(err, ErrUnsupportedScanType))
	// Output: true
}

// Бенчмарк
func BenchmarkScalarOIDScan(b *testing.B) {
	oidStr := "1.3.6.1.2.1.1.1.0"

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var scalar ScalarOID
		_ = scalar.Scan(oidStr)
	}
}

// Сравнение с OID.Scan
func BenchmarkScalarOIDVsOIDScan(b *testing.B) {
	oidStr := "1.3.6.1.2.1.1.1.0"

	b.Run("ScalarOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var scalar ScalarOID
			_ = scalar.Scan(oidStr)
		}
	})

	b.Run("OID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var oid OID
			_ = oid.Scan(oidStr)
		}
	})
}
