package oid

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestOIDScan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected OID
		wantErr  error
	}{
		{
			name:     "Строка",
			input:    "1.3.6.1.4.1",
			expected: OID{1, 3, 6, 1, 4, 1},
			wantErr:  nil,
		},
		{
			name:     "Байты",
			input:    []byte("1.3.6.1.4.1"),
			expected: OID{1, 3, 6, 1, 4, 1},
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
			input:    []int{1, 2, 3},
			expected: nil,
			wantErr:  ErrUnsupportedScanType,
		},
		{
			name:     "Невалидная строка",
			input:    "invalid",
			expected: nil,
			wantErr:  nil, // Любая ошибка
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
			var oid OID
			err := oid.Scan(tt.input)

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

			if tt.expected == nil && tt.input != nil {
				if err == nil {
					t.Errorf("Scan(%v): expected error", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Scan: %v", err)
				return
			}

			if !oid.Equal(tt.expected) {
				t.Errorf("Scan = %v, want %v", oid, tt.expected)
			}
		})
	}
}

func TestOIDScanRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			// Получаем значение
			value, err := oid.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			// Сканируем обратно
			var decoded OID
			if err := decoded.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %v -> %v", oid, value, decoded)
			}
		})
	}
}

func TestOIDScanImplementsScanner(t *testing.T) {
	var _ sql.Scanner = (*OID)(nil)

	var oid OID
	if err := oid.Scan("1.3.6.1"); err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if !oid.Equal(OID{1, 3, 6, 1}) {
		t.Errorf("Scan = %v, want 1.3.6.1", oid)
	}
}

func TestOIDScanProperties(t *testing.T) {
	t.Run("Scan очищает при NULL", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		if err := oid.Scan(nil); err != nil {
			t.Fatalf("Scan(nil): %v", err)
		}

		if len(oid) != 0 {
			t.Errorf("len = %d, want 0", len(oid))
		}
	})

	t.Run("Scan перезаписывает значение", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		if err := oid.Scan("2.100.3"); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		if !oid.Equal(OID{2, 100, 3}) {
			t.Errorf("Scan = %v, want 2.100.3", oid)
		}
	})
}

// Пример использования
func ExampleOID_Scan() {
	var oid OID

	if err := oid.Scan("1.3.6.1.4.1"); err != nil {
		panic(err)
	}

	fmt.Println(oid)
	// Output: 1.3.6.1.4.1
}

// Пример с NULL
func ExampleOID_Scan_null() {
	var oid OID

	if err := oid.Scan(nil); err != nil {
		panic(err)
	}

	fmt.Println(len(oid) == 0)
	// Output: true
}

// Пример с ошибкой
func ExampleOID_Scan_error() {
	var oid OID
	err := oid.Scan(123)
	fmt.Println(errors.Is(err, ErrUnsupportedScanType))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDScan(b *testing.B) {
	oidStr := "1.3.6.1.4.1.99999.1.1"

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var oid OID
		_ = oid.Scan(oidStr)
	}
}

func TestOIDValue(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected driver.Value
		wantErr  error
	}{
		{
			name:     "Стандартный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: "1.3.6.1.4.1",
			wantErr:  nil,
		},
		{
			name:     "Короткий OID",
			oid:      OID{1, 3, 6},
			expected: "1.3.6",
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			oid:      OID{2, 100, 3},
			expected: "2.100.3",
			wantErr:  nil,
		},
		{
			name:     "С первым 0",
			oid:      OID{0, 39, 1},
			expected: "0.39.1",
			wantErr:  nil,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: nil,
			wantErr:  nil,
		},
		{
			name:     "Nil OID",
			oid:      nil,
			expected: nil,
			wantErr:  nil,
		},
		{
			name:     "Один компонент",
			oid:      OID{1},
			expected: nil,
			wantErr:  ErrOIDTooShort,
		},
		{
			name:     "Невалидный",
			oid:      OID{3, 1},
			expected: nil,
			wantErr:  ErrFirstComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.oid.Value()

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

func TestOIDValueTypes(t *testing.T) {
	t.Run("Возвращает string", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		value, err := oid.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if _, ok := value.(string); !ok {
			t.Errorf("Value тип = %T, want string", value)
		}
	})

	t.Run("Возвращает nil для пустого", func(t *testing.T) {
		oid := OID{}

		value, err := oid.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if value != nil {
			t.Errorf("Value = %v, want nil", value)
		}
	})
}

func TestOIDValueImplementsValuer(t *testing.T) {
	var _ driver.Valuer = OID{1, 3, 6, 1}

	oid := OID{1, 3, 6, 1}

	value, err := oid.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	if value == nil {
		t.Error("Value не должен быть nil")
	}
}

func TestOIDValueRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			value, err := oid.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			var decoded OID
			if err := decoded.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %v -> %v", oid, value, decoded)
			}
		})
	}
}

func TestOIDValueNotModifyOID(t *testing.T) {
	oid := OID{1, 3, 6, 1}
	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)

	oid.Value()

	if !oid.Equal(oidCopy) {
		t.Error("Value() не должен изменять OID")
	}
}

// Пример использования
func ExampleOID_Value() {
	oid := OID{1, 3, 6, 1, 4, 1}

	value, err := oid.Value()
	if err != nil {
		panic(err)
	}

	fmt.Println(value)
	// Output: 1.3.6.1.4.1
}

// Пример с пустым OID
func ExampleOID_Value_empty() {
	oid := OID{}

	value, err := oid.Value()
	if err != nil {
		panic(err)
	}

	fmt.Println(value == nil)
	// Output: true
}

// Пример с ошибкой
func ExampleOID_Value_error() {
	oid := OID{3, 1}

	_, err := oid.Value()
	fmt.Println(errors.Is(err, ErrFirstComponentTooBig))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDValue(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = oid.Value()
	}
}

func TestFromOID(t *testing.T) {
	tests := []struct {
		name          string
		oid           OID
		expectedValid bool
		expectedOID   OID
	}{
		{
			name:          "Валидный OID",
			oid:           OID{1, 3, 6, 1},
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1},
		},
		{
			name:          "Длинный OID",
			oid:           OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
		},
		{
			name:          "Пустой OID",
			oid:           OID{},
			expectedValid: false,
			expectedOID:   nil,
		},
		{
			name:          "Nil OID",
			oid:           nil,
			expectedValid: false,
			expectedOID:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromOID(tt.oid)

			if result.Valid != tt.expectedValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.expectedValid)
			}

			if result.Valid {
				if !result.OID.Equal(tt.expectedOID) {
					t.Errorf("OID = %v, want %v", result.OID, tt.expectedOID)
				}
			} else {
				if result.OID != nil {
					t.Errorf("OID = %v, want nil", result.OID)
				}
			}
		})
	}
}

func TestFromOIDStoresReference(t *testing.T) {
	oid := MustParseOID("1.3.6.1")

	result := FromOID(oid)

	// Изменяем оригинал
	oid[0] = 99

	// NullOID должен хранить ссылку (не копию)
	if result.OID[0] != 99 {
		t.Error("FromOID должен хранить ссылку на OID")
	}

	// Восстанавливаем
	oid[0] = 1
}

func TestFromOIDRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			nullOID := FromOID(oid)

			if !nullOID.Valid {
				t.Error("Valid должно быть true")
			}

			if !nullOID.OID.Equal(oid) {
				t.Errorf("OID = %v, want %v", nullOID.OID, oid)
			}
		})
	}
}

func TestFromOIDConsistency(t *testing.T) {
	// FromOID с пустым OID должен давать NullOID{Valid: false}
	empty := FromOID(OID{})

	if empty.Valid {
		t.Error("Valid должно быть false для пустого OID")
	}

	if empty.OID != nil {
		t.Error("OID должно быть nil для пустого OID")
	}

	// FromOID с nil должен давать NullOID{Valid: false}
	nilOID := FromOID(nil)

	if nilOID.Valid {
		t.Error("Valid должно быть false для nil OID")
	}

	if nilOID.OID != nil {
		t.Error("OID должно быть nil для nil OID")
	}
}

// Пример использования
func ExampleFromOID() {
	oid := MustParseOID("1.3.6.1")

	nullOID := FromOID(oid)

	fmt.Println(nullOID.Valid)
	fmt.Println(nullOID.OID)
	// Output:
	// true
	// 1.3.6.1
}

// Пример с пустым OID
func ExampleFromOID_empty() {
	nullOID := FromOID(OID{})

	fmt.Println(nullOID.Valid)
	// Output: false
}

// Бенчмарк
func BenchmarkFromOID(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = FromOID(oid)
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValid bool
		expectedOID   OID
		wantErr       error
	}{
		{
			name:          "Валидный OID",
			input:         "1.3.6.1",
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1},
			wantErr:       nil,
		},
		{
			name:          "Длинный OID",
			input:         "1.3.6.1.2.1.1.1.0",
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:       nil,
		},
		{
			name:          "Пустая строка",
			input:         "",
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       nil,
		},
		{
			name:          "Невалидный OID",
			input:         "invalid",
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       nil, // Любая ошибка
		},
		{
			name:          "Один компонент",
			input:         "1",
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       ErrOIDTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromString(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("FromString: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("FromString = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if tt.expectedValid {
				// Ожидаем успех
				if err != nil {
					t.Errorf("FromString: %v", err)
					return
				}

				if result.Valid != tt.expectedValid {
					t.Errorf("Valid = %v, want %v", result.Valid, tt.expectedValid)
				}

				if !result.OID.Equal(tt.expectedOID) {
					t.Errorf("OID = %v, want %v", result.OID, tt.expectedOID)
				}
			} else {
				// Пустая строка - Valid = false, нет ошибки
				if tt.input == "" {
					if err != nil {
						t.Errorf("FromString(\"\"): %v", err)
					}
					if result.Valid {
						t.Error("Valid должно быть false для пустой строки")
					}
				} else {
					// Невалидный ввод - ожидаем ошибку
					if err == nil {
						t.Errorf("FromString(%q): expected error", tt.input)
					}
				}
			}
		})
	}
}

func TestFromStringEmpty(t *testing.T) {
	result, err := FromString("")

	if err != nil {
		t.Errorf("FromString(\"\"): %v", err)
	}

	if result.Valid {
		t.Error("Valid должно быть false")
	}

	if result.OID != nil {
		t.Error("OID должно быть nil")
	}
}

func TestFromStringRoundTrip(t *testing.T) {
	oids := []string{
		"1.3.6.1",
		"1.3.6.1.2.1.1.1.0",
		"2.100.3",
		"0.39.1",
	}

	for _, input := range oids {
		t.Run(input, func(t *testing.T) {
			nullOID, err := FromString(input)
			if err != nil {
				t.Fatalf("FromString: %v", err)
			}

			if !nullOID.Valid {
				t.Error("Valid должно быть true")
			}

			if nullOID.OID.String() != input {
				t.Errorf("OID.String() = %q, want %q", nullOID.OID.String(), input)
			}
		})
	}
}

func TestFromStringConsistency(t *testing.T) {
	// FromString с валидным OID должен давать тот же результат, что и FromOID + ParseOID
	input := "1.3.6.1"

	result1, err1 := FromString(input)
	if err1 != nil {
		t.Fatalf("FromString: %v", err1)
	}

	oid, err2 := ParseOID(input)
	if err2 != nil {
		t.Fatalf("ParseOID: %v", err2)
	}
	result2 := FromOID(oid)

	if result1.Valid != result2.Valid {
		t.Error("Valid должны совпадать")
	}

	if result1.Valid && !result1.OID.Equal(result2.OID) {
		t.Error("OID должны совпадать")
	}
}

// Пример использования
func ExampleFromString() {
	nullOID, err := FromString("1.3.6.1")
	if err != nil {
		panic(err)
	}

	fmt.Println(nullOID.Valid)
	fmt.Println(nullOID.OID)
	// Output:
	// true
	// 1.3.6.1
}

// Пример с пустой строкой
func ExampleFromString_empty() {
	nullOID, err := FromString("")
	if err != nil {
		panic(err)
	}

	fmt.Println(nullOID.Valid)
	// Output: false
}

// Пример с ошибкой
func ExampleFromString_error() {
	_, err := FromString("invalid")
	fmt.Println(err != nil)
	// Output: true
}

// Бенчмарк
func BenchmarkFromString(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = FromString("1.3.6.1.4.1.99999")
	}
}

func TestMustFromString(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValid bool
		expectedOID   OID
	}{
		{
			name:          "Валидный OID",
			input:         "1.3.6.1",
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1},
		},
		{
			name:          "Длинный OID",
			input:         "1.3.6.1.2.1.1.1.0",
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
		},
		{
			name:          "Пустая строка",
			input:         "",
			expectedValid: false,
			expectedOID:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MustFromString(tt.input)

			if result.Valid != tt.expectedValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.expectedValid)
			}

			if result.Valid {
				if !result.OID.Equal(tt.expectedOID) {
					t.Errorf("OID = %v, want %v", result.OID, tt.expectedOID)
				}
			} else {
				if result.OID != nil {
					t.Errorf("OID = %v, want nil", result.OID)
				}
			}
		})
	}
}

func TestMustFromStringPanic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Невалидный", "invalid"},
		{"Один компонент", "1"},
		{"Первый > 2", "3.1"},
		{"Второй > 39", "1.40"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("MustFromString(%q): expected panic", tt.input)
				}
			}()

			MustFromString(tt.input)
		})
	}
}

func TestMustFromStringPanicMessage(t *testing.T) {
	t.Run("Паника содержит ошибку", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic")
				return
			}

			if _, ok := r.(error); !ok {
				t.Errorf("panic = %v, want error", r)
			}
		}()

		MustFromString("invalid")
	})
}

func TestMustFromStringConsistency(t *testing.T) {
	// MustFromString должен давать тот же результат, что и FromString
	input := "1.3.6.1"

	mustResult := MustFromString(input)
	fromResult, err := FromString(input)

	if err != nil {
		t.Fatalf("FromString: %v", err)
	}

	if mustResult.Valid != fromResult.Valid {
		t.Error("Valid должны совпадать")
	}

	if mustResult.Valid && !mustResult.OID.Equal(fromResult.OID) {
		t.Error("OID должны совпадать")
	}
}

// Пример использования
func ExampleMustFromString() {
	nullOID := MustFromString("1.3.6.1")

	fmt.Println(nullOID.Valid)
	fmt.Println(nullOID.OID)
	// Output:
	// true
	// 1.3.6.1
}

// Пример с пустой строкой
func ExampleMustFromString_empty() {
	nullOID := MustFromString("")

	fmt.Println(nullOID.Valid)
	// Output: false
}

// Пример с паникой
func ExampleMustFromString_panic() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Паника поймана")
		}
	}()

	MustFromString("invalid")
	// Output: Паника поймана
}

// Бенчмарк
func BenchmarkMustFromString(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = MustFromString("1.3.6.1.4.1.99999")
	}
}

func TestNullOIDValue(t *testing.T) {
	tests := []struct {
		name     string
		nullOID  NullOID
		expected driver.Value
		wantErr  error
	}{
		{
			name: "Валидный NullOID",
			nullOID: NullOID{
				OID:   MustParseOID("1.3.6.1"),
				Valid: true,
			},
			expected: "1.3.6.1",
			wantErr:  nil,
		},
		{
			name: "Длинный валидный",
			nullOID: NullOID{
				OID:   MustParseOID("1.3.6.1.2.1.1.1.0"),
				Valid: true,
			},
			expected: "1.3.6.1.2.1.1.1.0",
			wantErr:  nil,
		},
		{
			name:     "Невалидный (Valid = false)",
			nullOID:  NullOID{Valid: false},
			expected: nil,
			wantErr:  nil,
		},
		{
			name: "Пустой OID с Valid = true",
			nullOID: NullOID{
				OID:   OID{},
				Valid: true,
			},
			expected: nil,
			wantErr:  nil,
		},
		{
			name: "Nil OID с Valid = true",
			nullOID: NullOID{
				OID:   nil,
				Valid: true,
			},
			expected: nil,
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.nullOID.Value()

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

func TestNullOIDValueTypes(t *testing.T) {
	t.Run("Возвращает string", func(t *testing.T) {
		nullOID := NullOID{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		}

		value, err := nullOID.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if _, ok := value.(string); !ok {
			t.Errorf("Value тип = %T, want string", value)
		}
	})

	t.Run("Возвращает nil для невалидного", func(t *testing.T) {
		nullOID := NullOID{Valid: false}

		value, err := nullOID.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if value != nil {
			t.Errorf("Value = %v, want nil", value)
		}
	})
}

func TestNullOIDValueImplementsValuer(t *testing.T) {
	var _ driver.Valuer = NullOID{
		OID:   MustParseOID("1.3.6.1"),
		Valid: true,
	}
}

func TestNullOIDValueRoundTrip(t *testing.T) {
	tests := []NullOID{
		{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		},
		{
			OID:   MustParseOID("1.3.6.1.2.1.1.1.0"),
			Valid: true,
		},
		{
			Valid: false,
		},
	}

	for _, nullOID := range tests {
		t.Run(fmt.Sprintf("Valid=%v", nullOID.Valid), func(t *testing.T) {
			value, err := nullOID.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			var decoded NullOID
			if err := decoded.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			if decoded.Valid != nullOID.Valid {
				t.Errorf("Valid = %v, want %v", decoded.Valid, nullOID.Valid)
			}

			if decoded.Valid {
				if !decoded.OID.Equal(nullOID.OID) {
					t.Errorf("OID = %v, want %v", decoded.OID, nullOID.OID)
				}
			}
		})
	}
}

func TestNullOIDValueNotModify(t *testing.T) {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1"),
		Valid: true,
	}

	nullOIDCopy := NullOID{
		OID:   make(OID, len(nullOID.OID)),
		Valid: nullOID.Valid,
	}
	copy(nullOIDCopy.OID, nullOID.OID)

	nullOID.Value()

	if nullOID.Valid != nullOIDCopy.Valid {
		t.Error("Value() не должен изменять Valid")
	}

	if !nullOID.OID.Equal(nullOIDCopy.OID) {
		t.Error("Value() не должен изменять OID")
	}
}

// Пример использования
func ExampleNullOID_Value() {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1"),
		Valid: true,
	}

	value, err := nullOID.Value()
	if err != nil {
		panic(err)
	}

	fmt.Println(value)
	// Output: 1.3.6.1
}

// Пример с NULL
func ExampleNullOID_Value_null() {
	nullOID := NullOID{Valid: false}

	value, err := nullOID.Value()
	if err != nil {
		panic(err)
	}

	fmt.Println(value == nil)
	// Output: true
}

// Бенчмарк
func BenchmarkNullOIDValue(b *testing.B) {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1.4.1.99999"),
		Valid: true,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = nullOID.Value()
	}
}

func TestNullOIDScan(t *testing.T) {
	tests := []struct {
		name          string
		input         any
		expectedValid bool
		expectedOID   OID
		wantErr       error
	}{
		{
			name:          "Строка",
			input:         "1.3.6.1",
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1},
			wantErr:       nil,
		},
		{
			name:          "Байты",
			input:         []byte("1.3.6.1"),
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1},
			wantErr:       nil,
		},
		{
			name:          "NULL",
			input:         nil,
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       nil,
		},
		{
			name:          "Число",
			input:         123,
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       ErrUnsupportedScanType,
		},
		{
			name:          "Невалидная строка",
			input:         "invalid",
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       nil, // Любая ошибка
		},
		{
			name:          "Пустая строка",
			input:         "",
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       ErrEmptyOID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nullOID NullOID
			err := nullOID.Scan(tt.input)

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

			if tt.expectedValid {
				if err != nil {
					t.Errorf("Scan: %v", err)
					return
				}

				if !nullOID.Valid {
					t.Error("Valid должно быть true")
				}

				if !nullOID.OID.Equal(tt.expectedOID) {
					t.Errorf("OID = %v, want %v", nullOID.OID, tt.expectedOID)
				}
			} else {
				// NULL или ошибка
				if tt.input == nil {
					if err != nil {
						t.Errorf("Scan(nil): %v", err)
					}
					if nullOID.Valid {
						t.Error("Valid должно быть false для NULL")
					}
					if nullOID.OID != nil {
						t.Error("OID должно быть nil для NULL")
					}
				} else {
					// Невалидный ввод
					if err == nil {
						t.Errorf("Scan(%v): expected error", tt.input)
					}
				}
			}
		})
	}
}

func TestNullOIDScanRoundTrip(t *testing.T) {
	tests := []NullOID{
		{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		},
		{
			OID:   MustParseOID("1.3.6.1.2.1.1.1.0"),
			Valid: true,
		},
		{
			Valid: false,
		},
	}

	for _, nullOID := range tests {
		t.Run(fmt.Sprintf("Valid=%v", nullOID.Valid), func(t *testing.T) {
			value, err := nullOID.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			var decoded NullOID
			if err := decoded.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			if decoded.Valid != nullOID.Valid {
				t.Errorf("Valid = %v, want %v", decoded.Valid, nullOID.Valid)
			}

			if decoded.Valid {
				if !decoded.OID.Equal(nullOID.OID) {
					t.Errorf("OID = %v, want %v", decoded.OID, nullOID.OID)
				}
			}
		})
	}
}

func TestNullOIDScanImplementsScanner(t *testing.T) {
	var _ sql.Scanner = (*NullOID)(nil)
}

func TestNullOIDScanProperties(t *testing.T) {
	t.Run("Scan NULL очищает", func(t *testing.T) {
		nullOID := NullOID{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		}

		if err := nullOID.Scan(nil); err != nil {
			t.Fatalf("Scan(nil): %v", err)
		}

		if nullOID.Valid {
			t.Error("Valid должно быть false после NULL")
		}
		if nullOID.OID != nil {
			t.Error("OID должно быть nil после NULL")
		}
	})

	t.Run("Scan перезаписывает", func(t *testing.T) {
		nullOID := NullOID{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		}

		if err := nullOID.Scan("2.100.3"); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		if !nullOID.OID.Equal(MustParseOID("2.100.3")) {
			t.Error("OID должен перезаписаться")
		}
	})
}

// Пример использования
func ExampleNullOID_Scan() {
	var nullOID NullOID

	if err := nullOID.Scan("1.3.6.1"); err != nil {
		panic(err)
	}

	fmt.Println(nullOID.Valid)
	fmt.Println(nullOID.OID)
	// Output:
	// true
	// 1.3.6.1
}

// Пример с NULL
func ExampleNullOID_Scan_null() {
	var nullOID NullOID

	if err := nullOID.Scan(nil); err != nil {
		panic(err)
	}

	fmt.Println(nullOID.Valid)
	// Output: false
}

// Пример с ошибкой
func ExampleNullOID_Scan_error() {
	var nullOID NullOID
	err := nullOID.Scan(123)
	fmt.Println(errors.Is(err, ErrUnsupportedScanType))
	// Output: true
}

// Бенчмарк
func BenchmarkNullOIDScan(b *testing.B) {
	oidStr := "1.3.6.1.4.1.99999"

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var nullOID NullOID
		_ = nullOID.Scan(oidStr)
	}
}

func TestNullOIDString(t *testing.T) {
	tests := []struct {
		name     string
		nullOID  NullOID
		expected string
	}{
		{
			name: "Валидный NullOID",
			nullOID: NullOID{
				OID:   MustParseOID("1.3.6.1"),
				Valid: true,
			},
			expected: "1.3.6.1",
		},
		{
			name: "Длинный валидный",
			nullOID: NullOID{
				OID:   MustParseOID("1.3.6.1.2.1.1.1.0"),
				Valid: true,
			},
			expected: "1.3.6.1.2.1.1.1.0",
		},
		{
			name:     "Невалидный (Valid = false)",
			nullOID:  NullOID{Valid: false},
			expected: "",
		},
		{
			name: "Пустой OID с Valid = true",
			nullOID: NullOID{
				OID:   OID{},
				Valid: true,
			},
			expected: "",
		},
		{
			name: "Nil OID с Valid = true",
			nullOID: NullOID{
				OID:   nil,
				Valid: true,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.nullOID.String()

			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNullOIDStringConsistency(t *testing.T) {
	t.Run("Совпадает с OID.String()", func(t *testing.T) {
		nullOID := NullOID{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		}

		if nullOID.String() != nullOID.OID.String() {
			t.Error("String() должен совпадать с OID.String()")
		}
	})

	t.Run("Пустая строка для невалидного", func(t *testing.T) {
		nullOID := NullOID{Valid: false}

		if nullOID.String() != "" {
			t.Error("String() должен вернуть пустую строку")
		}
	})
}

func TestNullOIDStringNotModify(t *testing.T) {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1"),
		Valid: true,
	}

	nullOIDCopy := NullOID{
		OID:   make(OID, len(nullOID.OID)),
		Valid: nullOID.Valid,
	}
	copy(nullOIDCopy.OID, nullOID.OID)

	nullOID.String()

	if nullOID.Valid != nullOIDCopy.Valid {
		t.Error("String() не должен изменять Valid")
	}

	if !nullOID.OID.Equal(nullOIDCopy.OID) {
		t.Error("String() не должен изменять OID")
	}
}

// Пример использования
func ExampleNullOID_String() {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1"),
		Valid: true,
	}

	fmt.Println(nullOID.String())
	// Output: 1.3.6.1
}

// Пример с NULL
func ExampleNullOID_String_null() {
	nullOID := NullOID{Valid: false}

	fmt.Println(nullOID.String() == "")
	// Output: true
}

// Бенчмарк
func BenchmarkNullOIDString(b *testing.B) {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1.4.1.99999"),
		Valid: true,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = nullOID.String()
	}
}

func TestNullOIDEqual(t *testing.T) {
	tests := []struct {
		name     string
		n        NullOID
		other    NullOID
		expected bool
	}{
		{
			name:     "Оба NULL",
			n:        NullOID{Valid: false},
			other:    NullOID{Valid: false},
			expected: true,
		},
		{
			name: "Оба валидные с одинаковым OID",
			n: NullOID{
				OID:   MustParseOID("1.3.6.1"),
				Valid: true,
			},
			other: NullOID{
				OID:   MustParseOID("1.3.6.1"),
				Valid: true,
			},
			expected: true,
		},
		{
			name: "Оба валидные с разными OID",
			n: NullOID{
				OID:   MustParseOID("1.3.6.1"),
				Valid: true,
			},
			other: NullOID{
				OID:   MustParseOID("2.100.3"),
				Valid: true,
			},
			expected: false,
		},
		{
			name: "NULL и валидный",
			n:    NullOID{Valid: false},
			other: NullOID{
				OID:   MustParseOID("1.3.6.1"),
				Valid: true,
			},
			expected: false,
		},
		{
			name: "Валидный и NULL",
			n: NullOID{
				OID:   MustParseOID("1.3.6.1"),
				Valid: true,
			},
			other:    NullOID{Valid: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.n.Equal(tt.other)

			if result != tt.expected {
				t.Errorf("Equal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNullOIDEqualProperties(t *testing.T) {
	t.Run("Рефлексивность", func(t *testing.T) {
		nullOIDs := []NullOID{
			{Valid: false},
			{OID: MustParseOID("1.3.6.1"), Valid: true},
		}

		for _, nullOID := range nullOIDs {
			if !nullOID.Equal(nullOID) {
				t.Errorf("Equal(%v, %v) = false, want true", nullOID, nullOID)
			}
		}
	})

	t.Run("Симметричность", func(t *testing.T) {
		n1 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
		n2 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}

		if n1.Equal(n2) != n2.Equal(n1) {
			t.Error("Equal должен быть симметричным")
		}
	})

	t.Run("Транзитивность", func(t *testing.T) {
		n1 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
		n2 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
		n3 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}

		if n1.Equal(n2) && n2.Equal(n3) {
			if !n1.Equal(n3) {
				t.Error("Equal должен быть транзитивным")
			}
		}
	})

	t.Run("Не изменяет NullOID", func(t *testing.T) {
		n1 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
		n2 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}

		n1Copy := NullOID{
			OID:   make(OID, len(n1.OID)),
			Valid: n1.Valid,
		}
		copy(n1Copy.OID, n1.OID)

		n1.Equal(n2)

		if n1.Valid != n1Copy.Valid {
			t.Error("Equal() не должен изменять Valid")
		}
		if !n1.OID.Equal(n1Copy.OID) {
			t.Error("Equal() не должен изменять OID")
		}
	})
}

// Пример использования
func ExampleNullOID_Equal() {
	null1 := NullOID{Valid: false}
	null2 := NullOID{Valid: false}

	valid1 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
	valid2 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}

	fmt.Println(null1.Equal(null2))
	fmt.Println(valid1.Equal(valid2))
	fmt.Println(null1.Equal(valid1))
	// Output:
	// true
	// true
	// false
}

// Бенчмарк
func BenchmarkNullOIDEqual(b *testing.B) {
	n1 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
	n2 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = n1.Equal(n2)
	}
}

func TestNullOIDMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		nullOID  NullOID
		expected string
	}{
		{
			name: "Валидный NullOID",
			nullOID: NullOID{
				OID:   MustParseOID("1.3.6.1"),
				Valid: true,
			},
			expected: `"1.3.6.1"`,
		},
		{
			name: "Длинный валидный",
			nullOID: NullOID{
				OID:   MustParseOID("1.3.6.1.2.1.1.1.0"),
				Valid: true,
			},
			expected: `"1.3.6.1.2.1.1.1.0"`,
		},
		{
			name:     "Невалидный (Valid = false)",
			nullOID:  NullOID{Valid: false},
			expected: "null",
		},
		{
			name: "Пустой OID с Valid = true",
			nullOID: NullOID{
				OID:   OID{},
				Valid: true,
			},
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.nullOID.MarshalJSON()

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

func TestNullOIDMarshalJSONValid(t *testing.T) {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1"),
		Valid: true,
	}

	data, err := nullOID.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if !json.Valid(data) {
		t.Errorf("MarshalJSON = %s, невалидный JSON", data)
	}
}

func TestNullOIDMarshalJSONRoundTrip(t *testing.T) {
	tests := []NullOID{
		{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		},
		{
			OID:   MustParseOID("1.3.6.1.2.1.1.1.0"),
			Valid: true,
		},
		{
			Valid: false,
		},
	}

	for _, nullOID := range tests {
		t.Run(fmt.Sprintf("Valid=%v", nullOID.Valid), func(t *testing.T) {
			// Кодируем
			data, err := nullOID.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}

			// Декодируем
			var decoded NullOID
			if err := decoded.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}

			if decoded.Valid != nullOID.Valid {
				t.Errorf("Valid = %v, want %v", decoded.Valid, nullOID.Valid)
			}

			if decoded.Valid {
				if !decoded.OID.Equal(nullOID.OID) {
					t.Errorf("OID = %v, want %v", decoded.OID, nullOID.OID)
				}
			}
		})
	}
}

func TestNullOIDMarshalJSONNotModify(t *testing.T) {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1"),
		Valid: true,
	}

	nullOIDCopy := NullOID{
		OID:   make(OID, len(nullOID.OID)),
		Valid: nullOID.Valid,
	}
	copy(nullOIDCopy.OID, nullOID.OID)

	nullOID.MarshalJSON()

	if nullOID.Valid != nullOIDCopy.Valid {
		t.Error("MarshalJSON() не должен изменять Valid")
	}
	if !nullOID.OID.Equal(nullOIDCopy.OID) {
		t.Error("MarshalJSON() не должен изменять OID")
	}
}

// Пример использования
func ExampleNullOID_MarshalJSON() {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1"),
		Valid: true,
	}

	data, err := nullOID.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output: "1.3.6.1"
}

// Пример с NULL
func ExampleNullOID_MarshalJSON_null() {
	nullOID := NullOID{Valid: false}

	data, _ := nullOID.MarshalJSON()

	fmt.Println(string(data))
	// Output: null
}

// Бенчмарк
func BenchmarkNullOIDMarshalJSON(b *testing.B) {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1.4.1.99999"),
		Valid: true,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = nullOID.MarshalJSON()
	}
}

func TestNullOIDUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		expectedValid bool
		expectedOID   OID
		wantErr       error
	}{
		{
			name:          "Валидный OID",
			data:          []byte(`"1.3.6.1"`),
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1},
			wantErr:       nil,
		},
		{
			name:          "Длинный OID",
			data:          []byte(`"1.3.6.1.2.1.1.1.0"`),
			expectedValid: true,
			expectedOID:   OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:       nil,
		},
		{
			name:          "null",
			data:          []byte(`null`),
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       nil,
		},
		{
			name:          "Пустая строка",
			data:          []byte(`""`),
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       nil,
		},
		{
			name:          "Невалидный JSON",
			data:          []byte(`invalid`),
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       ErrInvalidJSONType,
		},
		{
			name:          "Число",
			data:          []byte(`123`),
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       ErrInvalidJSONType,
		},
		{
			name:          "Объект",
			data:          []byte(`{"oid":"1.3.6.1"}`),
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       ErrInvalidJSONType,
		},
		{
			name:          "Невалидный OID",
			data:          []byte(`"invalid"`),
			expectedValid: false,
			expectedOID:   nil,
			wantErr:       nil, // Любая ошибка
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nullOID NullOID
			err := nullOID.UnmarshalJSON(tt.data)

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

			if tt.expectedValid {
				if err != nil {
					t.Errorf("UnmarshalJSON: %v", err)
					return
				}

				if !nullOID.Valid {
					t.Error("Valid должно быть true")
				}

				if !nullOID.OID.Equal(tt.expectedOID) {
					t.Errorf("OID = %v, want %v", nullOID.OID, tt.expectedOID)
				}
			} else {
				// null или пустая строка
				if string(tt.data) == "null" || string(tt.data) == `""` {
					if err != nil {
						t.Errorf("UnmarshalJSON(%s): %v", tt.data, err)
					}
					if nullOID.Valid {
						t.Error("Valid должно быть false")
					}
					if nullOID.OID != nil {
						t.Error("OID должно быть nil")
					}
				} else {
					// Невалидный ввод
					if err == nil {
						t.Errorf("UnmarshalJSON(%s): expected error", tt.data)
					}
				}
			}
		})
	}
}

func TestNullOIDUnmarshalJSONRoundTrip(t *testing.T) {
	tests := []NullOID{
		{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		},
		{
			OID:   MustParseOID("1.3.6.1.2.1.1.1.0"),
			Valid: true,
		},
		{
			Valid: false,
		},
	}

	for _, nullOID := range tests {
		t.Run(fmt.Sprintf("Valid=%v", nullOID.Valid), func(t *testing.T) {
			// Кодируем
			data, err := nullOID.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}

			// Декодируем
			var decoded NullOID
			if err := decoded.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}

			if decoded.Valid != nullOID.Valid {
				t.Errorf("Valid = %v, want %v", decoded.Valid, nullOID.Valid)
			}

			if decoded.Valid {
				if !decoded.OID.Equal(nullOID.OID) {
					t.Errorf("OID = %v, want %v", decoded.OID, nullOID.OID)
				}
			}
		})
	}
}

func TestNullOIDUnmarshalJSONProperties(t *testing.T) {
	t.Run("Unmarshal null очищает", func(t *testing.T) {
		nullOID := NullOID{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		}

		if err := nullOID.UnmarshalJSON([]byte(`null`)); err != nil {
			t.Fatalf("UnmarshalJSON: %v", err)
		}

		if nullOID.Valid {
			t.Error("Valid должно быть false после null")
		}
		if nullOID.OID != nil {
			t.Error("OID должно быть nil после null")
		}
	})

	t.Run("Unmarshal перезаписывает", func(t *testing.T) {
		nullOID := NullOID{
			OID:   MustParseOID("1.3.6.1"),
			Valid: true,
		}

		if err := nullOID.UnmarshalJSON([]byte(`"2.100.3"`)); err != nil {
			t.Fatalf("UnmarshalJSON: %v", err)
		}

		if !nullOID.OID.Equal(MustParseOID("2.100.3")) {
			t.Error("OID должен перезаписаться")
		}
	})
}

// Пример использования
func ExampleNullOID_UnmarshalJSON() {
	var nullOID NullOID

	if err := nullOID.UnmarshalJSON([]byte(`"1.3.6.1"`)); err != nil {
		panic(err)
	}

	fmt.Println(nullOID.Valid)
	fmt.Println(nullOID.OID)
	// Output:
	// true
	// 1.3.6.1
}

// Пример с null
func ExampleNullOID_UnmarshalJSON_null() {
	var nullOID NullOID

	if err := nullOID.UnmarshalJSON([]byte(`null`)); err != nil {
		panic(err)
	}

	fmt.Println(nullOID.Valid)
	// Output: false
}

// Бенчмарк
func BenchmarkNullOIDUnmarshalJSON(b *testing.B) {
	data := []byte(`"1.3.6.1.4.1.99999"`)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var nullOID NullOID
		_ = nullOID.UnmarshalJSON(data)
	}
}

func TestArrayValue(t *testing.T) {
	tests := []struct {
		name     string
		array    Array
		expected driver.Value
		wantErr  error
	}{
		{
			name:     "Пустой массив",
			array:    Array{},
			expected: "{}",
			wantErr:  nil,
		},
		{
			name:     "Nil массив",
			array:    nil,
			expected: "{}",
			wantErr:  nil,
		},
		{
			name: "Один OID",
			array: Array{
				MustParseOID("1.3.6.1"),
			},
			expected: "{1.3.6.1}",
			wantErr:  nil,
		},
		{
			name: "Два OID",
			array: Array{
				MustParseOID("1.3.6.1"),
				MustParseOID("2.100.3"),
			},
			expected: "{1.3.6.1,2.100.3}",
			wantErr:  nil,
		},
		{
			name: "Три OID",
			array: Array{
				MustParseOID("1.3.6.1"),
				MustParseOID("2.100.3"),
				MustParseOID("0.39.1"),
			},
			expected: "{1.3.6.1,2.100.3,0.39.1}",
			wantErr:  nil,
		},
		{
			name: "Невалидный OID",
			array: Array{
				MustParseOID("1.3.6.1"),
				OID{3, 1},
			},
			expected: nil,
			wantErr:  ErrInvalidArrayOID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.array.Value()

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

func TestArrayValueTypes(t *testing.T) {
	t.Run("Возвращает string", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1")}

		value, err := array.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if _, ok := value.(string); !ok {
			t.Errorf("Value тип = %T, want string", value)
		}
	})
}

func TestArrayValueImplementsValuer(t *testing.T) {
	var _ driver.Valuer = Array{MustParseOID("1.3.6.1")}
}

func TestArrayValueRoundTrip(t *testing.T) {
	tests := []Array{
		{},
		{MustParseOID("1.3.6.1")},
		{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
		{MustParseOID("1.3.6.1"), MustParseOID("2.100.3"), MustParseOID("0.39.1")},
	}

	for _, array := range tests {
		t.Run(fmt.Sprintf("len=%d", len(array)), func(t *testing.T) {
			value, err := array.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			var decoded Array
			if err := decoded.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			if !array.Equal(decoded) {
				t.Errorf("Round trip: %v -> %v -> %v", array, value, decoded)
			}
		})
	}
}

func TestArrayValueNotModifyArray(t *testing.T) {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	arrayCopy := make(Array, len(array))
	copy(arrayCopy, array)

	array.Value()

	if !array.Equal(arrayCopy) {
		t.Error("Value() не должен изменять массив")
	}
}

// Пример использования
func ExampleArray_Value() {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	value, err := array.Value()
	if err != nil {
		panic(err)
	}

	fmt.Println(value)
	// Output: {1.3.6.1,2.100.3}
}

// Пример с пустым массивом
func ExampleArray_Value_empty() {
	array := Array{}

	value, _ := array.Value()

	fmt.Println(value)
	// Output: {}
}

// Бенчмарк
func BenchmarkArrayValue(b *testing.B) {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = array.Value()
	}
}

func TestArrayScan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected Array
		wantErr  error
	}{
		{
			name:     "Строка с одним OID",
			input:    "{1.3.6.1}",
			expected: Array{MustParseOID("1.3.6.1")},
			wantErr:  nil,
		},
		{
			name:     "Строка с двумя OID",
			input:    "{1.3.6.1,2.100.3}",
			expected: Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
			wantErr:  nil,
		},
		{
			name:     "Байты",
			input:    []byte("{1.3.6.1}"),
			expected: Array{MustParseOID("1.3.6.1")},
			wantErr:  nil,
		},
		{
			name:     "NULL",
			input:    nil,
			expected: nil,
			wantErr:  nil,
		},
		{
			name:     "Пустой массив",
			input:    "{}",
			expected: Array{},
			wantErr:  nil,
		},
		{
			name:     "Число",
			input:    123,
			expected: nil,
			wantErr:  ErrUnsupportedScanType,
		},
		{
			name:     "Невалидный формат",
			input:    "not-array",
			expected: nil,
			wantErr:  ErrInvalidArrayFormat,
		},
		{
			name:     "Нет открывающей скобки",
			input:    "1.3.6.1}",
			expected: nil,
			wantErr:  ErrInvalidArrayFormat,
		},
		{
			name:     "Нет закрывающей скобки",
			input:    "{1.3.6.1",
			expected: nil,
			wantErr:  ErrInvalidArrayFormat,
		},
		{
			name:     "Невалидный OID",
			input:    "{invalid}",
			expected: nil,
			wantErr:  nil, // Любая ошибка
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var array Array
			err := array.Scan(tt.input)

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

			if tt.expected != nil {
				if err != nil {
					t.Errorf("Scan: %v", err)
					return
				}

				if !array.Equal(tt.expected) {
					t.Errorf("Scan = %v, want %v", array, tt.expected)
				}
			} else {
				// NULL или ошибка
				if tt.input == nil {
					if err != nil {
						t.Errorf("Scan(nil): %v", err)
					}
					if array != nil {
						t.Error("array должно быть nil")
					}
				} else {
					// Невалидный ввод
					if err == nil {
						t.Errorf("Scan(%v): expected error", tt.input)
					}
				}
			}
		})
	}
}

func TestArrayScanRoundTrip(t *testing.T) {
	tests := []Array{
		{},
		{MustParseOID("1.3.6.1")},
		{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
		{MustParseOID("1.3.6.1"), MustParseOID("2.100.3"), MustParseOID("0.39.1")},
	}

	for _, array := range tests {
		t.Run(fmt.Sprintf("len=%d", len(array)), func(t *testing.T) {
			value, err := array.Value()
			if err != nil {
				t.Fatalf("Value: %v", err)
			}

			var decoded Array
			if err := decoded.Scan(value); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			if !array.Equal(decoded) {
				t.Errorf("Round trip: %v -> %v -> %v", array, value, decoded)
			}
		})
	}
}

func TestArrayScanImplementsScanner(t *testing.T) {
	var _ sql.Scanner = (*Array)(nil)
}

func TestArrayScanProperties(t *testing.T) {
	t.Run("Scan NULL очищает", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1")}

		if err := array.Scan(nil); err != nil {
			t.Fatalf("Scan(nil): %v", err)
		}

		if array != nil {
			t.Error("array должно быть nil после NULL")
		}
	})

	t.Run("Scan перезаписывает", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1")}

		if err := array.Scan("{2.100.3}"); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		if !array.Equal(Array{MustParseOID("2.100.3")}) {
			t.Error("array должен перезаписаться")
		}
	})
}

// Пример использования
func ExampleArray_Scan() {
	var array Array

	if err := array.Scan("{1.3.6.1,2.100.3}"); err != nil {
		panic(err)
	}

	fmt.Println(array)
	// Output: [1.3.6.1, 2.100.3]
}

// Пример с NULL
func ExampleArray_Scan_null() {
	var array Array

	if err := array.Scan(nil); err != nil {
		panic(err)
	}

	fmt.Println(array == nil)
	// Output: true
}

// Пример с ошибкой
func ExampleArray_Scan_error() {
	var array Array
	err := array.Scan(123)
	fmt.Println(errors.Is(err, ErrUnsupportedScanType))
	// Output: true
}

// Бенчмарк
func BenchmarkArrayScan(b *testing.B) {
	arrayStr := "{1.3.6.1,2.100.3}"

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var array Array
		_ = array.Scan(arrayStr)
	}
}

func TestArrayMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		array    Array
		expected string
	}{
		{
			name:     "Пустой массив",
			array:    Array{},
			expected: "[]",
		},
		{
			name:     "Nil массив",
			array:    nil,
			expected: "[]",
		},
		{
			name:     "Один OID",
			array:    Array{MustParseOID("1.3.6.1")},
			expected: `["1.3.6.1"]`,
		},
		{
			name:     "Два OID",
			array:    Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
			expected: `["1.3.6.1","2.100.3"]`,
		},
		{
			name:     "Три OID",
			array:    Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3"), MustParseOID("0.39.1")},
			expected: `["1.3.6.1","2.100.3","0.39.1"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.array.MarshalJSON()

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

func TestArrayMarshalJSONValid(t *testing.T) {
	array := Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")}

	data, err := array.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if !json.Valid(data) {
		t.Errorf("MarshalJSON = %s, невалидный JSON", data)
	}
}

func TestArrayMarshalJSONRoundTrip(t *testing.T) {
	tests := []Array{
		{},
		{MustParseOID("1.3.6.1")},
		{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
		{MustParseOID("1.3.6.1"), MustParseOID("2.100.3"), MustParseOID("0.39.1")},
	}

	for _, array := range tests {
		t.Run(fmt.Sprintf("len=%d", len(array)), func(t *testing.T) {
			// Кодируем
			data, err := array.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}

			// Декодируем
			var decoded Array
			if err := decoded.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}

			if !array.Equal(decoded) {
				t.Errorf("Round trip: %v -> %s -> %v", array, data, decoded)
			}
		})
	}
}

func TestArrayMarshalJSONNotModifyArray(t *testing.T) {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	arrayCopy := make(Array, len(array))
	copy(arrayCopy, array)

	array.MarshalJSON()

	if !array.Equal(arrayCopy) {
		t.Error("MarshalJSON() не должен изменять массив")
	}
}

// Пример использования
func ExampleArray_MarshalJSON() {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	data, err := array.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output: ["1.3.6.1","2.100.3"]
}

// Пример с пустым массивом
func ExampleArray_MarshalJSON_empty() {
	array := Array{}

	data, _ := array.MarshalJSON()

	fmt.Println(string(data))
	// Output: []
}

// Бенчмарк
func BenchmarkArrayMarshalJSON(b *testing.B) {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = array.MarshalJSON()
	}
}

func TestArrayUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected Array
		wantErr  error
	}{
		{
			name:     "Пустой массив",
			data:     []byte(`[]`),
			expected: Array{},
			wantErr:  nil,
		},
		{
			name:     "null",
			data:     []byte(`null`),
			expected: Array{},
			wantErr:  nil,
		},
		{
			name:     "Один OID",
			data:     []byte(`["1.3.6.1"]`),
			expected: Array{MustParseOID("1.3.6.1")},
			wantErr:  nil,
		},
		{
			name:     "Два OID",
			data:     []byte(`["1.3.6.1","2.100.3"]`),
			expected: Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
			wantErr:  nil,
		},
		{
			name:     "Три OID",
			data:     []byte(`["1.3.6.1","2.100.3","0.39.1"]`),
			expected: Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3"), MustParseOID("0.39.1")},
			wantErr:  nil,
		},
		{
			name:     "Невалидный JSON",
			data:     []byte(`invalid`),
			expected: nil,
			wantErr:  ErrJSONDecodeArray,
		},
		{
			name:     "Число",
			data:     []byte(`123`),
			expected: nil,
			wantErr:  ErrJSONDecodeArray,
		},
		{
			name:     "Объект",
			data:     []byte(`{"oid":"1.3.6.1"}`),
			expected: nil,
			wantErr:  ErrJSONDecodeArray,
		},
		{
			name:     "Невалидный OID",
			data:     []byte(`["invalid"]`),
			expected: nil,
			wantErr:  nil, // Любая ошибка
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var array Array
			err := array.UnmarshalJSON(tt.data)

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

			if tt.expected != nil {
				if err != nil {
					t.Errorf("UnmarshalJSON: %v", err)
					return
				}

				if !array.Equal(tt.expected) {
					t.Errorf("UnmarshalJSON = %v, want %v", array, tt.expected)
				}
			} else {
				// Невалидный ввод
				if err == nil {
					t.Errorf("UnmarshalJSON(%s): expected error", tt.data)
				}
			}
		})
	}
}

func TestArrayUnmarshalJSONRoundTrip(t *testing.T) {
	tests := []Array{
		{},
		{MustParseOID("1.3.6.1")},
		{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
		{MustParseOID("1.3.6.1"), MustParseOID("2.100.3"), MustParseOID("0.39.1")},
	}

	for _, array := range tests {
		t.Run(fmt.Sprintf("len=%d", len(array)), func(t *testing.T) {
			// Кодируем
			data, err := array.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}

			// Декодируем
			var decoded Array
			if err := decoded.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}

			if !array.Equal(decoded) {
				t.Errorf("Round trip: %v -> %s -> %v", array, data, decoded)
			}
		})
	}
}

func TestArrayUnmarshalJSONProperties(t *testing.T) {
	t.Run("Unmarshal [] очищает", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1")}

		if err := array.UnmarshalJSON([]byte(`[]`)); err != nil {
			t.Fatalf("UnmarshalJSON: %v", err)
		}

		if len(array) != 0 {
			t.Error("array должно быть пустым после []")
		}
	})

	t.Run("Unmarshal перезаписывает", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1")}

		if err := array.UnmarshalJSON([]byte(`["2.100.3"]`)); err != nil {
			t.Fatalf("UnmarshalJSON: %v", err)
		}

		if !array.Equal(Array{MustParseOID("2.100.3")}) {
			t.Error("array должен перезаписаться")
		}
	})
}

// Пример использования
func ExampleArray_UnmarshalJSON() {
	var array Array

	if err := array.UnmarshalJSON([]byte(`["1.3.6.1","2.100.3"]`)); err != nil {
		panic(err)
	}

	fmt.Println(array)
	// Output: [1.3.6.1, 2.100.3]
}

// Пример с пустым массивом
func ExampleArray_UnmarshalJSON_empty() {
	var array Array

	if err := array.UnmarshalJSON([]byte(`[]`)); err != nil {
		panic(err)
	}

	fmt.Println(len(array))
	// Output: 0
}

// Пример с ошибкой
func ExampleArray_UnmarshalJSON_error() {
	var array Array
	err := array.UnmarshalJSON([]byte(`invalid`))
	fmt.Println(errors.Is(err, ErrJSONDecodeArray))
	// Output: true
}

// Бенчмарк
func BenchmarkArrayUnmarshalJSON(b *testing.B) {
	data := []byte(`["1.3.6.1","2.100.3"]`)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var array Array
		_ = array.UnmarshalJSON(data)
	}
}

func TestArrayString(t *testing.T) {
	tests := []struct {
		name     string
		array    Array
		expected string
	}{
		{
			name:     "Пустой массив",
			array:    Array{},
			expected: "[]",
		},
		{
			name:     "Nil массив",
			array:    nil,
			expected: "[]",
		},
		{
			name:     "Один OID",
			array:    Array{MustParseOID("1.3.6.1")},
			expected: "[1.3.6.1]",
		},
		{
			name:     "Два OID",
			array:    Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
			expected: "[1.3.6.1, 2.100.3]",
		},
		{
			name:     "Три OID",
			array:    Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3"), MustParseOID("0.39.1")},
			expected: "[1.3.6.1, 2.100.3, 0.39.1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.array.String()

			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestArrayStringProperties(t *testing.T) {
	t.Run("Не изменяет массив", func(t *testing.T) {
		array := Array{
			MustParseOID("1.3.6.1"),
			MustParseOID("2.100.3"),
		}

		arrayCopy := make(Array, len(array))
		copy(arrayCopy, array)

		array.String()

		if !array.Equal(arrayCopy) {
			t.Error("String() не должен изменять массив")
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")}

		str1 := array.String()
		str2 := array.String()

		if str1 != str2 {
			t.Error("String() должен быть детерминированным")
		}
	})

	t.Run("Содержит все OID", func(t *testing.T) {
		array := Array{
			MustParseOID("1.3.6.1"),
			MustParseOID("2.100.3"),
		}

		str := array.String()

		if !strings.Contains(str, "1.3.6.1") {
			t.Error("String не содержит 1.3.6.1")
		}
		if !strings.Contains(str, "2.100.3") {
			t.Error("String не содержит 2.100.3")
		}
	})
}

func TestArrayStringRoundTrip(t *testing.T) {
	tests := []Array{
		{},
		{MustParseOID("1.3.6.1")},
		{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
	}

	for _, array := range tests {
		t.Run(fmt.Sprintf("len=%d", len(array)), func(t *testing.T) {
			str := array.String()

			// String не должен быть пустым для непустого массива
			if len(array) > 0 && str == "" {
				t.Error("String не должен быть пустым")
			}
		})
	}
}

// Пример использования
func ExampleArray_String() {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	fmt.Println(array.String())
	// Output: [1.3.6.1, 2.100.3]
}

// Пример с пустым массивом
func ExampleArray_String_empty() {
	array := Array{}

	fmt.Println(array.String())
	// Output: []
}

// Бенчмарк
func BenchmarkArrayString(b *testing.B) {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = array.String()
	}
}

func TestArrayEqual(t *testing.T) {
	tests := []struct {
		name     string
		array1   Array
		array2   Array
		expected bool
	}{
		{
			name:     "Оба пустые",
			array1:   Array{},
			array2:   Array{},
			expected: true,
		},
		{
			name:     "Оба nil",
			array1:   nil,
			array2:   nil,
			expected: true,
		},
		{
			name:     "Пустой и nil",
			array1:   Array{},
			array2:   nil,
			expected: true,
		},
		{
			name:     "Одинаковые с одним OID",
			array1:   Array{MustParseOID("1.3.6.1")},
			array2:   Array{MustParseOID("1.3.6.1")},
			expected: true,
		},
		{
			name:     "Одинаковые с двумя OID",
			array1:   Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
			array2:   Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
			expected: true,
		},
		{
			name:     "Разные OID",
			array1:   Array{MustParseOID("1.3.6.1")},
			array2:   Array{MustParseOID("2.100.3")},
			expected: false,
		},
		{
			name:     "Разная длина",
			array1:   Array{MustParseOID("1.3.6.1")},
			array2:   Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
			expected: false,
		},
		{
			name:     "Пустой и непустой",
			array1:   Array{},
			array2:   Array{MustParseOID("1.3.6.1")},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.array1.Equal(tt.array2)

			if result != tt.expected {
				t.Errorf("Equal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestArrayEqualProperties(t *testing.T) {
	t.Run("Рефлексивность", func(t *testing.T) {
		arrays := []Array{
			{},
			{MustParseOID("1.3.6.1")},
			{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
		}

		for _, array := range arrays {
			if !array.Equal(array) {
				t.Errorf("Equal(%v, %v) = false, want true", array, array)
			}
		}
	})

	t.Run("Симметричность", func(t *testing.T) {
		array1 := Array{MustParseOID("1.3.6.1")}
		array2 := Array{MustParseOID("1.3.6.1")}

		if array1.Equal(array2) != array2.Equal(array1) {
			t.Error("Equal должен быть симметричным")
		}
	})

	t.Run("Транзитивность", func(t *testing.T) {
		array1 := Array{MustParseOID("1.3.6.1")}
		array2 := Array{MustParseOID("1.3.6.1")}
		array3 := Array{MustParseOID("1.3.6.1")}

		if array1.Equal(array2) && array2.Equal(array3) {
			if !array1.Equal(array3) {
				t.Error("Equal должен быть транзитивным")
			}
		}
	})

	t.Run("Не изменяет массивы", func(t *testing.T) {
		array1 := Array{MustParseOID("1.3.6.1")}
		array2 := Array{MustParseOID("1.3.6.1")}

		array1Copy := make(Array, len(array1))
		copy(array1Copy, array1)

		array1.Equal(array2)

		if !array1.Equal(array1Copy) {
			t.Error("Equal() не должен изменять массив")
		}
	})
}

// Пример использования
func ExampleArray_Equal() {
	array1 := Array{MustParseOID("1.3.6.1")}
	array2 := Array{MustParseOID("1.3.6.1")}
	array3 := Array{MustParseOID("2.100.3")}

	fmt.Println(array1.Equal(array2))
	fmt.Println(array1.Equal(array3))
	// Output:
	// true
	// false
}

// Пример с пустыми массивами
func ExampleArray_Equal_empty() {
	empty1 := Array{}
	empty2 := Array{}

	fmt.Println(empty1.Equal(empty2))
	// Output: true
}

// Бенчмарк
func BenchmarkArrayEqual(b *testing.B) {
	array1 := Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")}
	array2 := Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = array1.Equal(array2)
	}
}

func TestArrayContains(t *testing.T) {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
		MustParseOID("0.39.1"),
	}

	tests := []struct {
		name     string
		target   OID
		expected bool
	}{
		{
			name:     "Существующий OID",
			target:   MustParseOID("1.3.6.1"),
			expected: true,
		},
		{
			name:     "Существующий OID (второй)",
			target:   MustParseOID("2.100.3"),
			expected: true,
		},
		{
			name:     "Существующий OID (третий)",
			target:   MustParseOID("0.39.1"),
			expected: true,
		},
		{
			name:     "Несуществующий OID",
			target:   MustParseOID("1.3.6.99"),
			expected: false,
		},
		{
			name:     "Пустой OID",
			target:   OID{},
			expected: false,
		},
		{
			name:     "Nil OID",
			target:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := array.Contains(tt.target)

			if result != tt.expected {
				t.Errorf("Contains(%v) = %v, want %v",
					tt.target, result, tt.expected)
			}
		})
	}
}

func TestArrayContainsEmpty(t *testing.T) {
	emptyArray := Array{}

	if emptyArray.Contains(MustParseOID("1.3.6.1")) {
		t.Error("Пустой массив не должен содержать OID")
	}
}

func TestArrayContainsNil(t *testing.T) {
	var nilArray Array

	if nilArray.Contains(MustParseOID("1.3.6.1")) {
		t.Error("Nil массив не должен содержать OID")
	}
}

func TestArrayContainsProperties(t *testing.T) {
	t.Run("Не изменяет массив", func(t *testing.T) {
		array := Array{
			MustParseOID("1.3.6.1"),
			MustParseOID("2.100.3"),
		}

		arrayCopy := make(Array, len(array))
		copy(arrayCopy, array)

		array.Contains(MustParseOID("1.3.6.1"))

		if !array.Equal(arrayCopy) {
			t.Error("Contains() не должен изменять массив")
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1")}
		target := MustParseOID("1.3.6.1")

		result1 := array.Contains(target)
		result2 := array.Contains(target)

		if result1 != result2 {
			t.Error("Contains должен быть детерминированным")
		}
	})
}

// Пример использования
func ExampleArray_Contains() {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	fmt.Println(array.Contains(MustParseOID("1.3.6.1")))
	fmt.Println(array.Contains(MustParseOID("1.3.6.99")))
	// Output:
	// true
	// false
}

// Пример с пустым массивом
func ExampleArray_Contains_empty() {
	array := Array{}

	fmt.Println(array.Contains(MustParseOID("1.3.6.1")))
	// Output: false
}

// Бенчмарк
func BenchmarkArrayContains(b *testing.B) {
	array := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}
	target := MustParseOID("1.3.6.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = array.Contains(target)
	}
}

func TestArrayAppend(t *testing.T) {
	tests := []struct {
		name     string
		array    Array
		oids     []OID
		expected Array
	}{
		{
			name:     "Добавление одного OID",
			array:    Array{},
			oids:     []OID{MustParseOID("1.3.6.1")},
			expected: Array{MustParseOID("1.3.6.1")},
		},
		{
			name:     "Добавление к существующему",
			array:    Array{MustParseOID("1.3.6.1")},
			oids:     []OID{MustParseOID("2.100.3")},
			expected: Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3")},
		},
		{
			name:     "Добавление нескольких OID",
			array:    Array{MustParseOID("1.3.6.1")},
			oids:     []OID{MustParseOID("2.100.3"), MustParseOID("0.39.1")},
			expected: Array{MustParseOID("1.3.6.1"), MustParseOID("2.100.3"), MustParseOID("0.39.1")},
		},
		{
			name:     "Без OID",
			array:    Array{MustParseOID("1.3.6.1")},
			oids:     []OID{},
			expected: Array{MustParseOID("1.3.6.1")},
		},
		{
			name:     "Nil OIDs",
			array:    Array{MustParseOID("1.3.6.1")},
			oids:     nil,
			expected: Array{MustParseOID("1.3.6.1")},
		},
		{
			name:     "Добавление к nil",
			array:    nil,
			oids:     []OID{MustParseOID("1.3.6.1")},
			expected: Array{MustParseOID("1.3.6.1")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.array.Append(tt.oids...)

			if !result.Equal(tt.expected) {
				t.Errorf("Append() = %v, want %v", result, tt.expected)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("len = %d, want %d", len(result), len(tt.expected))
			}
		})
	}
}

func TestArrayAppendImmutability(t *testing.T) {
	original := Array{MustParseOID("1.3.6.1")}
	originalCopy := make(Array, len(original))
	copy(originalCopy, original)

	result := original.Append(MustParseOID("2.100.3"))

	// Оригинал не должен измениться
	if !original.Equal(originalCopy) {
		t.Errorf("Оригинал изменился: %v -> %v", originalCopy, original)
	}

	// Результат должен отличаться
	if result.Equal(original) {
		t.Error("Результат должен отличаться от оригинала")
	}
}

func TestArrayAppendProperties(t *testing.T) {
	t.Run("Append увеличивает длину", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1")}

		result := array.Append(MustParseOID("2.100.3"))
		if len(result) != len(array)+1 {
			t.Errorf("len = %d, want %d", len(result), len(array)+1)
		}

		result = array.Append(MustParseOID("2.100.3"), MustParseOID("0.39.1"))
		if len(result) != len(array)+2 {
			t.Errorf("len = %d, want %d", len(result), len(array)+2)
		}
	})

	t.Run("Append сохраняет префикс", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1")}

		result := array.Append(MustParseOID("2.100.3"))

		if !result[0].Equal(array[0]) {
			t.Error("Первый элемент должен сохраниться")
		}
	})

	t.Run("Append пустого списка возвращает копию", func(t *testing.T) {
		array := Array{MustParseOID("1.3.6.1")}

		result := array.Append()

		if !result.Equal(array) {
			t.Error("Append() должен вернуть тот же массив")
		}
	})
}

// Пример использования
func ExampleArray_Append() {
	array := Array{MustParseOID("1.3.6.1")}

	extended := array.Append(MustParseOID("2.100.3"))

	fmt.Println(array)
	fmt.Println(extended)
	// Output:
	// [1.3.6.1]
	// [1.3.6.1, 2.100.3]
}

// Пример с пустым массивом
func ExampleArray_Append_empty() {
	array := Array{}

	result := array.Append(MustParseOID("1.3.6.1"))

	fmt.Println(result)
	// Output: [1.3.6.1]
}

// Бенчмарк
func BenchmarkArrayAppend(b *testing.B) {
	array := Array{MustParseOID("1.3.6.1")}
	oid := MustParseOID("2.100.3")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = array.Append(oid)
	}
}

func TestSplitPostgresArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Пустая строка",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Один элемент",
			input:    "1.3.6.1",
			expected: []string{"1.3.6.1"},
		},
		{
			name:     "Два элемента",
			input:    "1.3.6.1,2.100.3",
			expected: []string{"1.3.6.1", "2.100.3"},
		},
		{
			name:     "Три элемента",
			input:    "1.3.6.1,2.100.3,0.39.1",
			expected: []string{"1.3.6.1", "2.100.3", "0.39.1"},
		},
		{
			name:     "С кавычками",
			input:    `"1.3.6.1","2.100.3"`,
			expected: []string{"1.3.6.1", "2.100.3"},
		},
		{
			name:     "С пробелами",
			input:    "1.3.6.1, 2.100.3",
			expected: []string{"1.3.6.1", " 2.100.3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitPostgresArray(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("len = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("result[%d] = %q, want %q",
						i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSplitPostgresArrayQuotes(t *testing.T) {
	t.Run("Одинарные кавычки", func(t *testing.T) {
		result := splitPostgresArray(`"1.3.6.1"`)

		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		if result[0] != "1.3.6.1" {
			t.Errorf("result[0] = %q, want '1.3.6.1'", result[0])
		}
	})

	t.Run("Кавычки с запятой внутри", func(t *testing.T) {
		result := splitPostgresArray(`"1.3.6.1,2.100.3"`)

		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		if result[0] != "1.3.6.1,2.100.3" {
			t.Errorf("result[0] = %q, want '1.3.6.1,2.100.3'", result[0])
		}
	})
}

func TestSplitPostgresArrayEscaping(t *testing.T) {
	t.Run("Экранированные кавычки", func(t *testing.T) {
		result := splitPostgresArray(`"1.3.6.1\"extra"`)

		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		t.Logf("result[0] = %q", result[0])
	})

	t.Run("Экранированная запятая", func(t *testing.T) {
		result := splitPostgresArray(`1.3.6.1\,2.100.3`)

		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
		t.Logf("result[0] = %q", result[0])
	})
}

func TestSplitPostgresArrayProperties(t *testing.T) {
	t.Run("Не изменяет входную строку", func(t *testing.T) {
		input := "1.3.6.1,2.100.3"

		splitPostgresArray(input)

		if input != "1.3.6.1,2.100.3" {
			t.Error("splitPostgresArray не должен изменять входную строку")
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		input := "1.3.6.1,2.100.3"

		result1 := splitPostgresArray(input)
		result2 := splitPostgresArray(input)

		if len(result1) != len(result2) {
			t.Error("splitPostgresArray должен быть детерминированным")
		}

		for i := range result1 {
			if result1[i] != result2[i] {
				t.Error("splitPostgresArray должен быть детерминированным")
			}
		}
	})
}

func TestSplitPostgresArrayRoundTrip(t *testing.T) {
	// Проверяем через Array.Scan
	tests := []string{
		"{1.3.6.1}",
		"{1.3.6.1,2.100.3}",
		"{1.3.6.1,2.100.3,0.39.1}",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			var array Array
			if err := array.Scan(input); err != nil {
				t.Fatalf("Scan: %v", err)
			}

			if len(array) == 0 {
				t.Error("array не должен быть пустым")
			}
		})
	}
}

// Бенчмарк
func BenchmarkSplitPostgresArray(b *testing.B) {
	input := "1.3.6.1,2.100.3,0.39.1"

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = splitPostgresArray(input)
	}
}
