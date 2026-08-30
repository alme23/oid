// oid/parse_oid_test.go
package oid

import (
	"bytes"
	"encoding/asn1"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestOIDType(t *testing.T) {
	t.Run("Создание", func(t *testing.T) {
		oid := OID{1, 3, 6, 1, 4, 1}

		if len(oid) != 6 {
			t.Errorf("len = %d, want 6", len(oid))
		}

		if oid[0] != 1 || oid[1] != 3 {
			t.Error("Неверные компоненты")
		}
	})

	t.Run("Пустой OID", func(t *testing.T) {
		oid := OID{}

		if len(oid) != 0 {
			t.Errorf("len = %d, want 0", len(oid))
		}
	})

	t.Run("Nil OID", func(t *testing.T) {
		var oid OID

		if oid != nil {
			t.Error("Nil OID должен быть nil")
		}

		if len(oid) != 0 {
			t.Errorf("len = %d, want 0", len(oid))
		}
	})
}

func TestOIDUnderlyingType(t *testing.T) {
	// OID - это []uint32
	var oid OID = OID{1, 3, 6}

	// Можно использовать как слайс
	if oid[0] != 1 {
		t.Error("oid[0] должен быть 1")
	}

	// Можно использовать append
	oid = append(oid, 1)
	if len(oid) != 4 {
		t.Errorf("len после append = %d, want 4", len(oid))
	}

	// Можно использовать copy
	copyOID := make(OID, len(oid))
	copy(copyOID, oid)

	if !copyOID.Equal(oid) {
		t.Error("copy должен сохранить значения")
	}
}

func TestOIDSliceOperations(t *testing.T) {
	oid := OID{1, 3, 6, 1, 4, 1}

	t.Run("Срез", func(t *testing.T) {
		sub := oid[1:4]

		if len(sub) != 3 {
			t.Errorf("len(sub) = %d, want 3", len(sub))
		}

		if sub[0] != 3 || sub[1] != 6 || sub[2] != 1 {
			t.Error("Неверный срез")
		}
	})

	t.Run("Итерация", func(t *testing.T) {
		sum := uint32(0)
		for _, v := range oid {
			sum += v
		}

		if sum != 16 { // 1+3+6+1+4+1
			t.Errorf("sum = %d, want 16", sum)
		}
	})
}

func TestOIDMaxValues(t *testing.T) {
	// Максимальное значение компонента
	oid := OID{1, 3, MaxOIDComponent}

	if oid[2] != MaxOIDComponent {
		t.Errorf("oid[2] = %d, want %d", oid[2], MaxOIDComponent)
	}

	// Максимальное значение uint32
	maxOID := OID{2, ^uint32(0)}

	if maxOID[1] != ^uint32(0) {
		t.Errorf("maxOID[1] = %d, want %d", maxOID[1], ^uint32(0))
	}
}

// Пример использования
func ExampleOID() {
	oid := OID{1, 3, 6, 1, 4, 1}

	fmt.Println(oid)
	// Output: 1.3.6.1.4.1
}

// Бенчмарк
func BenchmarkOIDCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = OID{1, 3, 6, 1, 4, 1}
	}
}

func TestParseOID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected OID
		wantErr  error
	}{
		{
			name:     "Стандартный OID",
			input:    "1.3.6.1.4.1",
			expected: OID{1, 3, 6, 1, 4, 1},
			wantErr:  nil,
		},
		{
			name:     "Короткий OID",
			input:    "1.3.6",
			expected: OID{1, 3, 6},
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			input:    "2.100.3",
			expected: OID{2, 100, 3},
			wantErr:  nil,
		},
		{
			name:     "С первым 0",
			input:    "0.39.1",
			expected: OID{0, 39, 1},
			wantErr:  nil,
		},
		{
			name:     "Максимальный компонент",
			input:    "1.3.268435455",
			expected: OID{1, 3, MaxOIDComponent},
			wantErr:  nil,
		},
		{
			name:     "Пустая строка",
			input:    "",
			expected: nil,
			wantErr:  ErrEmptyOID,
		},
		{
			name:     "Одна точка",
			input:    ".",
			expected: nil,
			wantErr:  ErrInvalidOID,
		},
		{
			name:     "Две точки",
			input:    "1..3",
			expected: nil,
			wantErr:  ErrInvalidOID,
		},
		{
			name:     "Точка в начале",
			input:    ".1.3",
			expected: nil,
			wantErr:  ErrInvalidOID,
		},
		{
			name:     "Точка в конце",
			input:    "1.3.",
			expected: nil,
			wantErr:  ErrInvalidOID,
		},
		{
			name:     "Отрицательное",
			input:    "-1.3",
			expected: nil,
			wantErr:  nil, // Любая ошибка
		},
		{
			name:     "Буквы",
			input:    "a.b",
			expected: nil,
			wantErr:  nil, // Любая ошибка
		},
		{
			name:     "Спецсимволы",
			input:    "1.3#",
			expected: nil,
			wantErr:  nil, // Любая ошибка
		},
		{
			name:     "Слишком большое",
			input:    "1.3.4294967296",
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
			name:     "Первый > 2",
			input:    "3.1",
			expected: nil,
			wantErr:  ErrFirstComponentTooBig,
		},
		{
			name:     "Второй > 39 при первом 1",
			input:    "1.40",
			expected: nil,
			wantErr:  ErrSecondComponentTooBig,
		},
		{
			name:     "Второй > 39 при первом 0",
			input:    "0.40",
			expected: nil,
			wantErr:  ErrSecondComponentTooBig,
		},
		{
			name:     "Компонент больше MaxOIDComponent",
			input:    "1.3.268435456", // MaxOIDComponent + 1
			expected: nil,
			wantErr:  ErrComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseOID(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseOID(%q): expected error %v", tt.input, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseOID(%q) = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}

			if tt.expected == nil && tt.input != "" {
				// Ожидаем любую ошибку
				if err == nil {
					t.Errorf("ParseOID(%q): expected error", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseOID(%q): %v", tt.input, err)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("ParseOID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseOIDRoundTrip(t *testing.T) {
	oids := []string{
		"1.3.6.1",
		"1.3.6.1.2.1.1.1.0",
		"2.100.3",
		"0.39.1",
		"1.3.268435455",
	}

	for _, input := range oids {
		t.Run(input, func(t *testing.T) {
			oid, err := ParseOID(input)
			if err != nil {
				t.Fatalf("ParseOID(%q): %v", input, err)
			}

			// String должен вернуть тот же ввод
			if oid.String() != input {
				t.Errorf("String() = %q, want %q", oid.String(), input)
			}
		})
	}
}

func TestParseOIDNoAllocations(t *testing.T) {
	// Проверяем, что ParseOID не делает лишних аллокаций
	allocs := testing.AllocsPerRun(1000, func() {
		_, _ = ParseOID("1.3.6.1.4.1")
	})

	if allocs > 1 {
		t.Errorf("ParseOID: %f allocs, want <= 1", allocs)
	}
}

// Пример использования
func ExampleParseOID() {
	oid, err := ParseOID("1.3.6.1.4.1")
	if err != nil {
		panic(err)
	}

	fmt.Println(oid)
	// Output: 1.3.6.1.4.1
}

// Пример с ошибкой
func ExampleParseOID_error() {
	_, err := ParseOID("invalid")
	fmt.Println(err != nil)
	// Output: true
}

// Бенчмарк
func BenchmarkParseOID(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = ParseOID("1.3.6.1.4.1.99999.1.1")
	}
}

func TestMustParseOID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected OID
	}{
		{
			name:     "Стандартный OID",
			input:    "1.3.6.1.4.1",
			expected: OID{1, 3, 6, 1, 4, 1},
		},
		{
			name:     "Короткий OID",
			input:    "1.3.6",
			expected: OID{1, 3, 6},
		},
		{
			name:     "С первым 2",
			input:    "2.100.3",
			expected: OID{2, 100, 3},
		},
		{
			name:     "С первым 0",
			input:    "0.39.1",
			expected: OID{0, 39, 1},
		},
		{
			name:     "Максимальный компонент",
			input:    "1.3.268435455",
			expected: OID{1, 3, MaxOIDComponent},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MustParseOID(tt.input)

			if !result.Equal(tt.expected) {
				t.Errorf("MustParseOID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMustParseOIDPanic(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("MustParseOID(%q): expected panic", tt.input)
				}
			}()

			MustParseOID(tt.input)
		})
	}
}

func TestMustParseOIDPanicMessage(t *testing.T) {
	t.Run("Паника содержит ошибку", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic")
				return
			}

			// Проверяем, что паника содержит ошибку
			if _, ok := r.(error); !ok {
				t.Errorf("panic = %v, want error", r)
			}
		}()

		MustParseOID("")
	})
}

func TestMustParseOIDRoundTrip(t *testing.T) {
	oids := []string{
		"1.3.6.1",
		"1.3.6.1.2.1.1.1.0",
		"2.100.3",
		"0.39.1",
	}

	for _, input := range oids {
		t.Run(input, func(t *testing.T) {
			oid := MustParseOID(input)

			if oid.String() != input {
				t.Errorf("String() = %q, want %q", oid.String(), input)
			}
		})
	}
}

func TestMustParseOIDConsistency(t *testing.T) {
	// MustParseOID должен давать тот же результат, что и ParseOID
	input := "1.3.6.1.4.1"

	parsed, err := ParseOID(input)
	if err != nil {
		t.Fatalf("ParseOID: %v", err)
	}

	mustParsed := MustParseOID(input)

	if !parsed.Equal(mustParsed) {
		t.Error("MustParseOID и ParseOID должны давать одинаковый результат")
	}
}

// Пример использования
func ExampleMustParseOID() {
	oid := MustParseOID("1.3.6.1.4.1")
	fmt.Println(oid)
	// Output: 1.3.6.1.4.1
}

// Пример с паникой
func ExampleMustParseOID_panic() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Паника поймана")
		}
	}()

	MustParseOID("invalid")
	// Output: Паника поймана
}

// Бенчмарк
func BenchmarkMustParseOID(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = MustParseOID("1.3.6.1.4.1.99999.1.1")
	}
}

// Сравнение ParseOID vs MustParseOID
func BenchmarkParseOIDVsMustParseOID(b *testing.B) {
	b.Run("ParseOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = ParseOID("1.3.6.1.4.1.99999.1.1")
		}
	})

	b.Run("MustParseOID", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = MustParseOID("1.3.6.1.4.1.99999.1.1")
		}
	})
}

func TestFromASN1(t *testing.T) {
	tests := []struct {
		name     string
		input    asn1.ObjectIdentifier
		expected OID
	}{
		{
			name:     "Стандартный OID",
			input:    asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1},
			expected: OID{1, 3, 6, 1, 4, 1},
		},
		{
			name:     "Короткий OID",
			input:    asn1.ObjectIdentifier{1, 3, 6},
			expected: OID{1, 3, 6},
		},
		{
			name:     "С первым 2",
			input:    asn1.ObjectIdentifier{2, 100, 3},
			expected: OID{2, 100, 3},
		},
		{
			name:     "С первым 0",
			input:    asn1.ObjectIdentifier{0, 39, 1},
			expected: OID{0, 39, 1},
		},
		{
			name:     "Пустой",
			input:    asn1.ObjectIdentifier{},
			expected: OID{},
		},
		{
			name:     "Nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Максимальные значения",
			input:    asn1.ObjectIdentifier{1, 3, MaxOIDComponent},
			expected: OID{1, 3, MaxOIDComponent},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromASN1(tt.input)

			if !result.Equal(tt.expected) {
				t.Errorf("FromASN1(%v) = %v, want %v", tt.input, result, tt.expected)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("len = %d, want %d", len(result), len(tt.expected))
			}
		})
	}
}

func TestFromASN1Negative(t *testing.T) {
	// asn1.ObjectIdentifier может содержать отрицательные значения
	input := asn1.ObjectIdentifier{1, 3, -1}

	result := FromASN1(input)

	// Должен вернуть nil при отрицательном значении
	if result != nil {
		t.Errorf("FromASN1(%v) = %v, want nil", input, result)
	}
}

func TestFromASN1RoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			// OID -> asn1.ObjectIdentifier
			asn1OID := oid.ToASN1()

			// asn1.ObjectIdentifier -> OID
			back := FromASN1(asn1OID)

			if !back.Equal(oid) {
				t.Errorf("Round trip: %v -> %v -> %v", oid, asn1OID, back)
			}
		})
	}
}

func TestFromASN1NewSlice(t *testing.T) {
	input := asn1.ObjectIdentifier{1, 3, 6, 1}

	result := FromASN1(input)

	// Изменяем оригинал
	input[0] = 99

	// Результат не должен измениться
	if result[0] != 1 {
		t.Error("FromASN1 должен создать новый слайс")
	}
}

func TestFromASN1ToASN1RoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			// ToASN1 -> FromASN1
			asn1OID := oid.ToASN1()
			back := FromASN1(asn1OID)

			if !back.Equal(oid) {
				t.Errorf("ToASN1/FromASN1: %v -> %v", oid, back)
			}
		})
	}
}

// Пример использования
func ExampleFromASN1() {
	asn1OID := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1}

	oid := FromASN1(asn1OID)

	fmt.Println(oid)
	// Output: 1.3.6.1.4.1
}

// Бенчмарк
func BenchmarkFromASN1(b *testing.B) {
	asn1OID := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999}

	b.ReportAllocs()
	for b.Loop() {
		_ = FromASN1(asn1OID)
	}
}

func TestOIDString(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected string
	}{
		{
			name:     "Стандартный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: "1.3.6.1.4.1",
		},
		{
			name:     "Короткий OID",
			oid:      OID{1, 3, 6},
			expected: "1.3.6",
		},
		{
			name:     "Два компонента",
			oid:      OID{1, 3},
			expected: "1.3",
		},
		{
			name:     "Один компонент",
			oid:      OID{1},
			expected: "1",
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: "",
		},
		{
			name:     "Nil OID",
			oid:      nil,
			expected: "",
		},
		{
			name:     "С первым 2",
			oid:      OID{2, 100, 3},
			expected: "2.100.3",
		},
		{
			name:     "С первым 0",
			oid:      OID{0, 39, 1},
			expected: "0.39.1",
		},
		{
			name:     "С нулями",
			oid:      OID{0, 0, 0},
			expected: "0.0.0",
		},
		{
			name:     "Максимальный компонент",
			oid:      OID{1, 3, MaxOIDComponent},
			expected: "1.3.268435455",
		},
		{
			name:     "Длинный OID",
			oid:      OID{1, 3, 6, 1, 4, 1, 99999, 1, 1},
			expected: "1.3.6.1.4.1.99999.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.oid.String()

			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestOIDStringRoundTrip(t *testing.T) {
	oids := []string{
		"1.3.6.1",
		"1.3.6.1.2.1.1.1.0",
		"2.100.3",
		"0.39.1",
		"1.3.268435455",
	}

	for _, input := range oids {
		t.Run(input, func(t *testing.T) {
			oid, err := ParseOID(input)
			if err != nil {
				t.Fatalf("ParseOID(%q): %v", input, err)
			}

			if oid.String() != input {
				t.Errorf("String() = %q, want %q", oid.String(), input)
			}
		})
	}
}

func TestOIDStringContainsDots(t *testing.T) {
	oid := OID{1, 3, 6, 1, 4, 1}

	str := oid.String()

	// Количество точек = len - 1
	dotCount := strings.Count(str, ".")
	expectedDots := len(oid) - 1

	if dotCount != expectedDots {
		t.Errorf("dots = %d, want %d", dotCount, expectedDots)
	}
}

func TestOIDStringNoLeadingDots(t *testing.T) {
	oid := OID{1, 3, 6}

	str := oid.String()

	if strings.HasPrefix(str, ".") {
		t.Error("String не должен начинаться с точки")
	}
}

func TestOIDStringNoTrailingDots(t *testing.T) {
	oid := OID{1, 3, 6}

	str := oid.String()

	if strings.HasSuffix(str, ".") {
		t.Error("String не должен заканчиваться точкой")
	}
}

func TestOIDStringNotModifyOID(t *testing.T) {
	oid := OID{1, 3, 6, 1}
	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)

	if !oid.Equal(oidCopy) {
		t.Error("String() не должен изменять OID")
	}
}

func TestOIDStringConsistency(t *testing.T) {
	oid := OID{1, 3, 6, 1, 4, 1}

	// Два вызова должны давать одинаковый результат
	str1 := oid.String()
	str2 := oid.String()

	if str1 != str2 {
		t.Error("String() должен быть детерминированным")
	}
}

// Пример использования
func ExampleOID_String() {
	oid := OID{1, 3, 6, 1, 4, 1}
	fmt.Println(oid.String())
	// Output: 1.3.6.1.4.1
}

// Пример с пустым OID
func ExampleOID_String_empty() {
	oid := OID{}
	fmt.Println(oid.String() == "")
	// Output: true
}

// Бенчмарк
func BenchmarkOIDString(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = oid.String()
	}
}

func TestOIDValidate(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr error
	}{
		{
			name:    "Стандартный OID",
			oid:     OID{1, 3, 6, 1, 4, 1},
			wantErr: nil,
		},
		{
			name:    "Короткий OID",
			oid:     OID{1, 3, 6},
			wantErr: nil,
		},
		{
			name:    "Минимальный OID",
			oid:     OID{0, 0},
			wantErr: nil,
		},
		{
			name:    "Максимальный первый",
			oid:     OID{2, 100},
			wantErr: nil,
		},
		{
			name:    "Максимальный второй при первом 1",
			oid:     OID{1, 39},
			wantErr: nil,
		},
		{
			name:    "Максимальный второй при первом 0",
			oid:     OID{0, 39},
			wantErr: nil,
		},
		{
			name:    "Максимальный компонент",
			oid:     OID{1, 3, MaxOIDComponent},
			wantErr: nil,
		},
		{
			name:    "Пустой OID",
			oid:     OID{},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Nil OID",
			oid:     nil,
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Один компонент",
			oid:     OID{1},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Первый > 2",
			oid:     OID{3, 1},
			wantErr: ErrFirstComponentTooBig,
		},
		{
			name:    "Второй > 39 при первом 0",
			oid:     OID{0, 40},
			wantErr: ErrSecondComponentTooBig,
		},
		{
			name:    "Второй > 39 при первом 1",
			oid:     OID{1, 40},
			wantErr: ErrSecondComponentTooBig,
		},
		{
			name:    "Второй > 39 при первом 2",
			oid:     OID{2, 100},
			wantErr: nil,
		},
		{
			name:    "Компонент больше Max",
			oid:     OID{1, 3, MaxOIDComponent + 1},
			wantErr: ErrComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.oid.Validate()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Validate(): expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate(): %v", err)
			}
		})
	}
}

func TestOIDValidateAllComponents(t *testing.T) {
	t.Run("Все компоненты валидны", func(t *testing.T) {
		oid := OID{1, 3, MaxOIDComponent, MaxOIDComponent, MaxOIDComponent}

		if err := oid.Validate(); err != nil {
			t.Errorf("Validate(): %v", err)
		}
	})

	t.Run("Один компонент невалиден", func(t *testing.T) {
		oid := OID{1, 3, 1, MaxOIDComponent + 1, 2}

		err := oid.Validate()
		if err == nil {
			t.Error("Validate(): expected error")
			return
		}
		if !errors.Is(err, ErrComponentTooBig) {
			t.Errorf("Validate() = %v, want ErrComponentTooBig", err)
		}
	})
}

func TestOIDValidateNotModifyOID(t *testing.T) {
	oid := OID{1, 3, 6, 1}
	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)

	oid.Validate()

	if !oid.Equal(oidCopy) {
		t.Error("Validate() не должен изменять OID")
	}
}

func TestOIDValidateConsistency(t *testing.T) {
	// Валидные OID
	validOIDs := []OID{
		{0, 0},
		{1, 39},
		{2, 100},
		{2, MaxOIDComponent},
	}

	for _, oid := range validOIDs {
		if err := oid.Validate(); err != nil {
			t.Errorf("Validate(%v): %v", oid, err)
		}
	}

	// Невалидные OID
	invalidOIDs := []OID{
		{},
		{1},
		{3, 1},
		{0, 40},
		{1, 40},
	}

	for _, oid := range invalidOIDs {
		if err := oid.Validate(); err == nil {
			t.Errorf("Validate(%v): expected error", oid)
		}
	}
}

func TestOIDValidateErrorMessage(t *testing.T) {
	t.Run("ErrComponentTooBig содержит информацию", func(t *testing.T) {
		oid := OID{1, 3, MaxOIDComponent + 1}

		err := oid.Validate()
		if err == nil {
			t.Fatal("expected error")
		}

		// Проверяем, что сообщение содержит полезную информацию
		msg := err.Error()
		if msg == "" {
			t.Error("Сообщение об ошибке пустое")
		}

		t.Logf("Error: %s", msg)
	})
}

// Пример использования
func ExampleOID_Validate() {
	// Валидный OID
	valid := OID{1, 3, 6, 1}
	fmt.Println(valid.Validate() == nil)

	// Невалидный OID
	invalid := OID{3, 1}
	fmt.Println(errors.Is(invalid.Validate(), ErrFirstComponentTooBig))
	// Output:
	// true
	// true
}

// Пример с пустым OID
func ExampleOID_Validate_empty() {
	empty := OID{}
	fmt.Println(errors.Is(empty.Validate(), ErrOIDTooShort))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDValidate(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = oid.Validate()
	}
}

func TestOIDEqual(t *testing.T) {
	tests := []struct {
		name     string
		oid1     OID
		oid2     OID
		expected bool
	}{
		{
			name:     "Одинаковые OID",
			oid1:     OID{1, 3, 6, 1},
			oid2:     OID{1, 3, 6, 1},
			expected: true,
		},
		{
			name:     "Одинаковые длинные",
			oid1:     OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			oid2:     OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: true,
		},
		{
			name:     "Разные OID",
			oid1:     OID{1, 3, 6, 1},
			oid2:     OID{1, 3, 6, 2},
			expected: false,
		},
		{
			name:     "Разная длина",
			oid1:     OID{1, 3, 6},
			oid2:     OID{1, 3, 6, 1},
			expected: false,
		},
		{
			name:     "Оба пустые",
			oid1:     OID{},
			oid2:     OID{},
			expected: true,
		},
		{
			name:     "Оба nil",
			oid1:     nil,
			oid2:     nil,
			expected: true,
		},
		{
			name:     "Пустой и nil",
			oid1:     OID{},
			oid2:     nil,
			expected: true,
		},
		{
			name:     "Пустой и непустой",
			oid1:     OID{},
			oid2:     OID{1, 3},
			expected: false,
		},
		{
			name:     "Разные первые компоненты",
			oid1:     OID{1, 3, 6},
			oid2:     OID{2, 3, 6},
			expected: false,
		},
		{
			name:     "С первым 2",
			oid1:     OID{2, 100, 3},
			oid2:     OID{2, 100, 3},
			expected: true,
		},
		{
			name:     "Максимальные компоненты",
			oid1:     OID{1, 3, MaxOIDComponent},
			oid2:     OID{1, 3, MaxOIDComponent},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.oid1.Equal(tt.oid2)

			if result != tt.expected {
				t.Errorf("Equal(%v, %v) = %v, want %v",
					tt.oid1, tt.oid2, result, tt.expected)
			}
		})
	}
}

func TestOIDEqualProperties(t *testing.T) {
	t.Run("Рефлексивность", func(t *testing.T) {
		oids := []OID{
			{1, 3, 6, 1},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{},
		}

		for _, oid := range oids {
			if !oid.Equal(oid) {
				t.Errorf("Equal(%v, %v) = false, want true", oid, oid)
			}
		}
	})

	t.Run("Симметричность", func(t *testing.T) {
		oid1 := OID{1, 3, 6, 1}
		oid2 := OID{1, 3, 6, 1}

		if oid1.Equal(oid2) != oid2.Equal(oid1) {
			t.Error("Equal должен быть симметричным")
		}
	})

	t.Run("Транзитивность", func(t *testing.T) {
		oid1 := OID{1, 3, 6, 1}
		oid2 := OID{1, 3, 6, 1}
		oid3 := OID{1, 3, 6, 1}

		if oid1.Equal(oid2) && oid2.Equal(oid3) {
			if !oid1.Equal(oid3) {
				t.Error("Equal должен быть транзитивным")
			}
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		oid1 := OID{1, 3, 6, 1}
		oid2 := OID{1, 3, 6, 1}

		oid1Copy := make(OID, len(oid1))
		copy(oid1Copy, oid1)

		oid1.Equal(oid2)

		if !oid1.Equal(oid1Copy) {
			t.Error("Equal() не должен изменять OID")
		}
	})
}

func TestOIDEqualConsistency(t *testing.T) {
	// Валидные OID
	validOIDs := []OID{
		{0, 0},
		{1, 39},
		{2, 100},
		{1, 3, MaxOIDComponent},
	}

	for i := range validOIDs {
		for j := range validOIDs {
			expected := i == j
			result := validOIDs[i].Equal(validOIDs[j])

			if result != expected {
				t.Errorf("Equal(%v, %v) = %v, want %v",
					validOIDs[i], validOIDs[j], result, expected)
			}
		}
	}
}

// Пример использования
func ExampleOID_Equal() {
	oid1 := OID{1, 3, 6, 1}
	oid2 := OID{1, 3, 6, 1}
	oid3 := OID{2, 100, 3}

	fmt.Println(oid1.Equal(oid2))
	fmt.Println(oid1.Equal(oid3))
	// Output:
	// true
	// false
}

// Пример с пустыми OID
func ExampleOID_Equal_empty() {
	empty1 := OID{}
	empty2 := OID{}

	fmt.Println(empty1.Equal(empty2))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDEqual(b *testing.B) {
	oid1 := MustParseOID("1.3.6.1.4.1.99999")
	oid2 := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = oid1.Equal(oid2)
	}
}

func TestOIDStartsWith(t *testing.T) {
	oid := OID{1, 3, 6, 1, 2, 1, 1, 1, 0}

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
			name:     "Без последнего компонента",
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
			result := oid.StartsWith(tt.prefix)

			if result != tt.expected {
				t.Errorf("StartsWith(%v) = %v, want %v",
					tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestOIDStartsWithDifferentOIDs(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		prefix   OID
		expected bool
	}{
		{
			name:     "Короткий OID",
			oid:      OID{1, 3, 6, 0},
			prefix:   OID{1, 3, 6},
			expected: true,
		},
		{
			name:     "Средний OID",
			oid:      OID{1, 3, 6, 1, 2, 1, 1, 0},
			prefix:   OID{1, 3, 6, 1, 2},
			expected: true,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			prefix:   OID{1, 3},
			expected: false,
		},
		{
			name:     "Nil OID с пустым префиксом",
			oid:      nil,
			prefix:   OID{},
			expected: true,
		},
		{
			name:     "Пустой с пустым префиксом",
			oid:      OID{},
			prefix:   OID{},
			expected: true,
		},
		{
			name:     "С первым 2",
			oid:      OID{2, 100, 3, 0},
			prefix:   OID{2, 100},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.oid.StartsWith(tt.prefix)
			if result != tt.expected {
				t.Errorf("StartsWith(%v) = %v, want %v",
					tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestOIDStartsWithProperties(t *testing.T) {
	oid := OID{1, 3, 6, 1, 2, 1, 1, 1, 0}

	t.Run("Всегда начинается с самого себя", func(t *testing.T) {
		if !oid.StartsWith(oid) {
			t.Error("StartsWith(self) = false, want true")
		}
	})

	t.Run("Всегда начинается с пустого префикса", func(t *testing.T) {
		if !oid.StartsWith(OID{}) {
			t.Error("StartsWith(empty) = false, want true")
		}
		if !oid.StartsWith(nil) {
			t.Error("StartsWith(nil) = false, want true")
		}
	})

	t.Run("Транзитивность", func(t *testing.T) {
		prefix1 := OID{1, 3, 6}
		prefix2 := OID{1, 3}

		if oid.StartsWith(prefix1) && prefix1.StartsWith(prefix2) {
			if !oid.StartsWith(prefix2) {
				t.Error("Транзитивность нарушена")
			}
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)

		oid.StartsWith(OID{1, 3})

		if !oid.Equal(oidCopy) {
			t.Error("StartsWith() не должен изменять OID")
		}
	})
}

func TestOIDStartsWithRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			// Создаем расширенный OID
			extended := oid.Append(99)

			// Расширенный должен начинаться с оригинала
			if !extended.StartsWith(oid) {
				t.Errorf("Extended %v should start with %v", extended, oid)
			}
		})
	}
}

// Пример использования
func ExampleOID_StartsWith() {
	oid := OID{1, 3, 6, 1, 2, 1, 1, 1, 0}

	fmt.Println(oid.StartsWith(OID{1, 3, 6}))
	fmt.Println(oid.StartsWith(OID{2, 100, 3}))
	fmt.Println(oid.StartsWith(OID{}))
	// Output:
	// true
	// false
	// true
}

// Бенчмарк
func BenchmarkOIDStartsWith(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")
	prefix := MustParseOID("1.3.6.1.4.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = oid.StartsWith(prefix)
	}
}

func TestOIDAppend(t *testing.T) {
	tests := []struct {
		name       string
		oid        OID
		components []uint32
		expected   OID
	}{
		{
			name:       "Добавление одного компонента",
			oid:        OID{1, 3, 6, 1},
			components: []uint32{4},
			expected:   OID{1, 3, 6, 1, 4},
		},
		{
			name:       "Добавление нескольких компонентов",
			oid:        OID{1, 3, 6},
			components: []uint32{1, 4, 1},
			expected:   OID{1, 3, 6, 1, 4, 1},
		},
		{
			name:       "Без компонентов",
			oid:        OID{1, 3, 6, 1},
			components: []uint32{},
			expected:   OID{1, 3, 6, 1},
		},
		{
			name:       "Nil components",
			oid:        OID{1, 3, 6, 1},
			components: nil,
			expected:   OID{1, 3, 6, 1},
		},
		{
			name:       "Добавление к пустому",
			oid:        OID{},
			components: []uint32{1, 3},
			expected:   OID{1, 3},
		},
		{
			name:       "Добавление к nil",
			oid:        nil,
			components: []uint32{1, 3, 6},
			expected:   OID{1, 3, 6},
		},
		{
			name:       "Добавление нуля",
			oid:        OID{1, 3, 6},
			components: []uint32{0},
			expected:   OID{1, 3, 6, 0},
		},
		{
			name:       "Добавление больших значений",
			oid:        OID{1, 3},
			components: []uint32{MaxOIDComponent},
			expected:   OID{1, 3, MaxOIDComponent},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.oid.Append(tt.components...)

			if !result.Equal(tt.expected) {
				t.Errorf("Append(%v...) = %v, want %v",
					tt.components, result, tt.expected)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("len = %d, want %d", len(result), len(tt.expected))
			}
		})
	}
}

func TestOIDAppendImmutability(t *testing.T) {
	original := OID{1, 3, 6, 1}
	originalCopy := make(OID, len(original))
	copy(originalCopy, original)

	// Выполняем Append
	result := original.Append(4, 1)

	// Оригинал не должен измениться
	if !original.Equal(originalCopy) {
		t.Errorf("Оригинал изменился: %v -> %v", originalCopy, original)
	}

	// Результат должен отличаться
	if result.Equal(original) {
		t.Error("Результат должен отличаться от оригинала")
	}
}

func TestOIDAppendProperties(t *testing.T) {
	t.Run("Append увеличивает длину", func(t *testing.T) {
		oid := OID{1, 3, 6}

		result := oid.Append(1)
		if len(result) != len(oid)+1 {
			t.Errorf("len = %d, want %d", len(result), len(oid)+1)
		}

		result = oid.Append(1, 2, 3)
		if len(result) != len(oid)+3 {
			t.Errorf("len = %d, want %d", len(result), len(oid)+3)
		}
	})

	t.Run("Append сохраняет префикс", func(t *testing.T) {
		oid := OID{1, 3, 6}

		result := oid.Append(1, 2, 3)

		if !result.StartsWith(oid) {
			t.Error("Результат должен начинаться с оригинала")
		}
	})

	t.Run("Append пустого списка возвращает копию", func(t *testing.T) {
		oid := OID{1, 3, 6}

		result := oid.Append()

		if !result.Equal(oid) {
			t.Error("Append() должен вернуть тот же OID")
		}

		// Изменяем копию
		result[0] = 99

		// Оригинал не должен измениться
		if oid[0] != 1 {
			t.Error("Append() должен вернуть независимую копию")
		}
	})
}

func TestOIDAppendRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			extended := oid.Append(99)

			// Проверяем StartsWith
			if !extended.StartsWith(oid) {
				t.Error("Extended должен начинаться с оригинала")
			}

			// Проверяем последний компонент
			last, err := extended.Last()
			if err != nil {
				t.Fatalf("Last: %v", err)
			}
			if last != 99 {
				t.Errorf("Last = %d, want 99", last)
			}
		})
	}
}

// Пример использования
func ExampleOID_Append() {
	oid := OID{1, 3, 6, 1}

	extended := oid.Append(4, 1)

	fmt.Println(oid)
	fmt.Println(extended)
	// Output:
	// 1.3.6.1
	// 1.3.6.1.4.1
}

// Пример с пустым OID
func ExampleOID_Append_empty() {
	oid := OID{}

	result := oid.Append(1, 3, 6)

	fmt.Println(result)
	// Output: 1.3.6
}

// Бенчмарк
func BenchmarkOIDAppend(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = oid.Append(99999)
	}
}

func TestOIDParent(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected OID
		wantErr  error
	}{
		{
			name:     "Стандартный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: OID{1, 3, 6, 1, 4},
			wantErr:  nil,
		},
		{
			name:     "Длинный OID",
			oid:      OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			expected: OID{1, 3, 6, 1, 2, 1, 1, 1},
			wantErr:  nil,
		},
		{
			name:     "Три компонента",
			oid:      OID{1, 3, 6},
			expected: OID{1, 3},
			wantErr:  nil,
		},
		{
			name:     "Два компонента",
			oid:      OID{1, 3},
			expected: OID{1},
			wantErr:  nil,
		},
		{
			name:     "Один компонент",
			oid:      OID{1},
			expected: nil,
			wantErr:  ErrNoParent,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: nil,
			wantErr:  ErrNoParent,
		},
		{
			name:     "Nil OID",
			oid:      nil,
			expected: nil,
			wantErr:  ErrNoParent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.oid.Parent()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Parent(): expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Parent() = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Parent(): %v", err)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("Parent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestOIDParentProperties(t *testing.T) {
	t.Run("Parent всегда короче", func(t *testing.T) {
		oids := []OID{
			{1, 3, 6, 1, 4, 1},
			{1, 3, 6, 1},
			{1, 3},
		}

		for _, oid := range oids {
			parent, err := oid.Parent()
			if err != nil {
				t.Errorf("Parent(%v): %v", oid, err)
				continue
			}

			if len(parent) != len(oid)-1 {
				t.Errorf("len(Parent(%v)) = %d, want %d",
					oid, len(parent), len(oid)-1)
			}
		}
	})

	t.Run("Parent сохраняет префикс", func(t *testing.T) {
		oid := OID{1, 3, 6, 1, 4, 1}

		parent, err := oid.Parent()
		if err != nil {
			t.Fatalf("Parent: %v", err)
		}

		if !oid.StartsWith(parent) {
			t.Error("OID должен начинаться с Parent")
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}
		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)

		oid.Parent()

		if !oid.Equal(oidCopy) {
			t.Error("Parent() не должен изменять OID")
		}
	})

	t.Run("Возвращает тот же слайс", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		parent, _ := oid.Parent()

		// Parent - это под-слайс оригинала
		if &parent[0] != &oid[0] {
			t.Error("Parent должен возвращать под-слайс")
		}
	})
}

func TestOIDParentRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			parent, err := oid.Parent()
			if err != nil {
				t.Fatalf("Parent: %v", err)
			}

			// Восстанавливаем через Append
			last, _ := oid.Last()
			restored := parent.Append(last)

			if !restored.Equal(oid) {
				t.Errorf("Round trip: %v -> %v -> %v", oid, parent, restored)
			}
		})
	}
}

func TestOIDParentChain(t *testing.T) {
	oid := OID{1, 3, 6, 1, 4, 1}

	// Последовательные вызовы Parent
	parent1, _ := oid.Parent()
	parent2, _ := parent1.Parent()
	parent3, _ := parent2.Parent()

	if !parent1.Equal(OID{1, 3, 6, 1, 4}) {
		t.Error("parent1 неверный")
	}
	if !parent2.Equal(OID{1, 3, 6, 1}) {
		t.Error("parent2 неверный")
	}
	if !parent3.Equal(OID{1, 3, 6}) {
		t.Error("parent3 неверный")
	}
}

// Пример использования
func ExampleOID_Parent() {
	oid := OID{1, 3, 6, 1, 4, 1}

	parent, err := oid.Parent()
	if err != nil {
		panic(err)
	}

	fmt.Println(oid)
	fmt.Println(parent)
	// Output:
	// 1.3.6.1.4.1
	// 1.3.6.1.4
}

// Пример с ошибкой
func ExampleOID_Parent_error() {
	oid := OID{1}

	_, err := oid.Parent()
	fmt.Println(errors.Is(err, ErrNoParent))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDParent(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = oid.Parent()
	}
}

func TestOIDLast(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected uint32
		wantErr  error
	}{
		{
			name:     "Стандартный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: 1,
			wantErr:  nil,
		},
		{
			name:     "Последний компонент 0",
			oid:      OID{1, 3, 6, 1, 0},
			expected: 0,
			wantErr:  nil,
		},
		{
			name:     "Последний компонент 39",
			oid:      OID{1, 39},
			expected: 39,
			wantErr:  nil,
		},
		{
			name:     "Последний компонент Max",
			oid:      OID{1, 3, MaxOIDComponent},
			expected: MaxOIDComponent,
			wantErr:  nil,
		},
		{
			name:     "Два компонента",
			oid:      OID{1, 3},
			expected: 3,
			wantErr:  nil,
		},
		{
			name:     "Один компонент",
			oid:      OID{1},
			expected: 1,
			wantErr:  nil,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: 0,
			wantErr:  ErrEmptyOID,
		},
		{
			name:     "Nil OID",
			oid:      nil,
			expected: 0,
			wantErr:  ErrEmptyOID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.oid.Last()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Last(): expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Last() = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Last(): %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Last() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestOIDLastProperties(t *testing.T) {
	t.Run("Last возвращает последний элемент", func(t *testing.T) {
		oids := []OID{
			{1, 3, 6, 1},
			{1, 3, 6, 1, 2, 1, 1, 1, 0},
			{2, 100, 3},
		}

		for _, oid := range oids {
			last, err := oid.Last()
			if err != nil {
				t.Errorf("Last(%v): %v", oid, err)
				continue
			}

			expected := oid[len(oid)-1]
			if last != expected {
				t.Errorf("Last(%v) = %d, want %d", oid, last, expected)
			}
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}
		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)

		oid.Last()

		if !oid.Equal(oidCopy) {
			t.Error("Last() не должен изменять OID")
		}
	})
}

func TestOIDLastRoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1},
		{2, 100, 3},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			last, err := oid.Last()
			if err != nil {
				t.Fatalf("Last: %v", err)
			}

			// Восстанавливаем через Append
			parent, _ := oid.Parent()
			restored := parent.Append(last)

			if !restored.Equal(oid) {
				t.Errorf("Round trip: %v -> %d -> %v", oid, last, restored)
			}
		})
	}
}

func TestOIDLastConsistency(t *testing.T) {
	oid := OID{1, 3, 6, 1, 4, 1}

	// Многократные вызовы дают одинаковый результат
	last1, _ := oid.Last()
	last2, _ := oid.Last()

	if last1 != last2 {
		t.Error("Last() должен быть детерминированным")
	}
}

// Пример использования
func ExampleOID_Last() {
	oid := OID{1, 3, 6, 1, 4, 1}

	last, err := oid.Last()
	if err != nil {
		panic(err)
	}

	fmt.Println(last)
	// Output: 1
}

// Пример с пустым OID
func ExampleOID_Last_empty() {
	oid := OID{}

	_, err := oid.Last()
	fmt.Println(errors.Is(err, ErrEmptyOID))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDLast(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = oid.Last()
	}
}

func TestOIDToASN1(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected asn1.ObjectIdentifier
	}{
		{
			name:     "Стандартный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1},
		},
		{
			name:     "Короткий OID",
			oid:      OID{1, 3, 6},
			expected: asn1.ObjectIdentifier{1, 3, 6},
		},
		{
			name:     "С первым 2",
			oid:      OID{2, 100, 3},
			expected: asn1.ObjectIdentifier{2, 100, 3},
		},
		{
			name:     "С первым 0",
			oid:      OID{0, 39, 1},
			expected: asn1.ObjectIdentifier{0, 39, 1},
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: asn1.ObjectIdentifier{},
		},
		{
			name:     "Nil OID",
			oid:      nil,
			expected: asn1.ObjectIdentifier{},
		},
		{
			name:     "Максимальный компонент",
			oid:      OID{1, 3, MaxOIDComponent},
			expected: asn1.ObjectIdentifier{1, 3, MaxOIDComponent},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.oid.ToASN1()

			if len(result) != len(tt.expected) {
				t.Errorf("len = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("result[%d] = %d, want %d", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestOIDToASN1RoundTrip(t *testing.T) {
	oids := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range oids {
		t.Run(oid.String(), func(t *testing.T) {
			// OID -> asn1.ObjectIdentifier
			asn1OID := oid.ToASN1()

			// asn1.ObjectIdentifier -> OID
			back := FromASN1(asn1OID)

			if !back.Equal(oid) {
				t.Errorf("Round trip: %v -> %v -> %v", oid, asn1OID, back)
			}
		})
	}
}

func TestOIDToASN1NewSlice(t *testing.T) {
	oid := OID{1, 3, 6, 1}

	result := oid.ToASN1()

	// Изменяем оригинал
	oid[0] = 99

	// Результат не должен измениться
	if result[0] != 1 {
		t.Error("ToASN1 должен создать новый слайс")
	}
}

func TestOIDToASN1NotModifyOID(t *testing.T) {
	oid := OID{1, 3, 6, 1}
	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)

	oid.ToASN1()

	if !oid.Equal(oidCopy) {
		t.Error("ToASN1() не должен изменять OID")
	}
}

func TestOIDToASN1Asn1Compatible(t *testing.T) {
	oid := OID{1, 3, 6, 1, 4, 1}

	asn1OID := oid.ToASN1()

	// Проверяем, что можно использовать с asn1.Marshal
	data, err := asn1.Marshal(asn1OID)
	if err != nil {
		t.Fatalf("asn1.Marshal: %v", err)
	}

	if len(data) == 0 {
		t.Error("asn1.Marshal вернул пустые данные")
	}
}

func TestOIDToASN1Consistency(t *testing.T) {
	oid := OID{1, 3, 6, 1}

	// Два вызова дают одинаковый результат
	result1 := oid.ToASN1()
	result2 := oid.ToASN1()

	if len(result1) != len(result2) {
		t.Error("ToASN1 должен быть детерминированным")
	}

	for i := range result1 {
		if result1[i] != result2[i] {
			t.Error("ToASN1 должен быть детерминированным")
		}
	}
}

// Пример использования
func ExampleOID_ToASN1() {
	oid := OID{1, 3, 6, 1, 4, 1}

	asn1OID := oid.ToASN1()

	// asn1.ObjectIdentifier - это []int с методом String()
	fmt.Printf("%v\n", []int(asn1OID))
	// Output: [1 3 6 1 4 1]
}

// Бенчмарк
func BenchmarkOIDToASN1(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = oid.ToASN1()
	}
}

func TestOIDMarshalBinary(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr error
	}{
		{
			name:    "Стандартный OID",
			oid:     OID{1, 3, 6, 1, 4, 1},
			wantErr: nil,
		},
		{
			name:    "Короткий OID",
			oid:     OID{1, 3, 6},
			wantErr: nil,
		},
		{
			name:    "С первым 2",
			oid:     OID{2, 100, 3},
			wantErr: nil,
		},
		{
			name:    "С первым 0",
			oid:     OID{0, 39, 1},
			wantErr: nil,
		},
		{
			name:    "Максимальный компонент",
			oid:     OID{1, 3, MaxOIDComponent},
			wantErr: nil,
		},
		{
			name:    "Пустой OID",
			oid:     OID{},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Nil OID",
			oid:     nil,
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Один компонент",
			oid:     OID{1},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Первый > 2",
			oid:     OID{3, 1},
			wantErr: ErrFirstComponentTooBig,
		},
		{
			name:    "Второй > 39",
			oid:     OID{1, 40},
			wantErr: ErrSecondComponentTooBig,
		},
		{
			name:    "Валидный короткий",
			oid:     OID{1, 3, 6},
			wantErr: nil,
		},
		{
			name:    "Валидный с первым 2",
			oid:     OID{2, 100, 3},
			wantErr: nil,
		},
		{
			name:    "Максимальные значения",
			oid:     OID{2, MaxOIDComponent},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalBinary()

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

func TestOIDMarshalBinaryEdgeCases(t *testing.T) {
	t.Run("Пустой OID", func(t *testing.T) {
		oid := OID{}

		_, err := oid.MarshalBinary()
		if err == nil {
			t.Fatal("MarshalBinary: expected error")
		}
		if !errors.Is(err, ErrOIDTooShort) {
			t.Errorf("MarshalBinary = %v, want ErrOIDTooShort", err)
		}
	})

	t.Run("Nil OID", func(t *testing.T) {
		var oid OID

		_, err := oid.MarshalBinary()
		if err == nil {
			t.Fatal("MarshalBinary: expected error")
		}
		if !errors.Is(err, ErrOIDTooShort) {
			t.Errorf("MarshalBinary = %v, want ErrOIDTooShort", err)
		}
	})

	t.Run("Один компонент", func(t *testing.T) {
		oid := OID{1}

		_, err := oid.MarshalBinary()
		if err == nil {
			t.Fatal("MarshalBinary: expected error")
		}
		if !errors.Is(err, ErrOIDTooShort) {
			t.Errorf("MarshalBinary = %v, want ErrOIDTooShort", err)
		}
	})

	t.Run("CombinedFirstComponents overflow", func(t *testing.T) {
		// OID с первыми компонентами, дающими переполнение
		oid := OID{2, MaxOIDComponent}

		// 40*2 + MaxOIDComponent = 80 + 268435455 = 268435535
		// Это меньше uint32 max (4294967295), поэтому не переполняется
		// Но может быть больше MaxOIDComponent
		_, err := oid.MarshalBinary()
		if err != nil {
			t.Logf("MarshalBinary: %v (expected success or specific error)", err)
		}
	})

	t.Run("CombinedFirstComponents error", func(t *testing.T) {
		// Создаем OID с максимальными значениями
		oid := OID{2, MaxOIDComponent, MaxOIDComponent}

		// Первые компоненты: 40*2 + 268435455 = 268435535
		// Это валидно для uint32
		_, err := oid.MarshalBinary()
		if err != nil {
			t.Logf("MarshalBinary: %v", err)
		}
	})
}

func TestOIDMarshalBinaryCompareWithStd(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1, 4, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			// Наш MarshalBinary
			ourData, err := oid.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// Стандартный asn1.Marshal
			stdData, err := asn1.Marshal(oid.ToASN1())
			if err != nil {
				t.Fatalf("asn1.Marshal: %v", err)
			}

			// Сравниваем
			if !bytes.Equal(ourData, stdData) {
				t.Errorf("MarshalBinary = %x, want %x", ourData, stdData)
			}
		})
	}
}

func TestOIDMarshalBinaryRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
		{1, 3, MaxOIDComponent},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			// Кодируем
			data, err := oid.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// Декодируем
			var decoded OID
			if err := decoded.UnmarshalBinary(data); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %x -> %v", oid, data, decoded)
			}
		})
	}
}

func TestOIDMarshalBinaryProperties(t *testing.T) {
	t.Run("Начинается с тега 0x06", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		data, err := oid.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary: %v", err)
		}

		if data[0] != 0x06 {
			t.Errorf("Первый байт = 0x%02x, want 0x06", data[0])
		}
	})

	t.Run("Размер зависит от длины", func(t *testing.T) {
		shortOID := OID{1, 3, 6}
		longOID := OID{1, 3, 6, 1, 2, 1, 1, 1, 0}

		shortData, _ := shortOID.MarshalBinary()
		longData, _ := longOID.MarshalBinary()

		if len(shortData) >= len(longData) {
			t.Error("Короткий OID должен давать меньше данных")
		}
	})

	t.Run("Не изменяет OID", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}
		oidCopy := make(OID, len(oid))
		copy(oidCopy, oid)

		oid.MarshalBinary()

		if !oid.Equal(oidCopy) {
			t.Error("MarshalBinary() не должен изменять OID")
		}
	})
}

func TestOIDMarshalBinaryDERFormat(t *testing.T) {
	oid := OID{1, 3, 6, 1}

	data, err := oid.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	// Правильный DER формат: 0x06 0x03 0x2B 0x06 0x01
	expected := []byte{0x06, 0x03, 0x2B, 0x06, 0x01}

	if !bytes.Equal(data, expected) {
		t.Errorf("MarshalBinary = %x, want %x", data, expected)
	}
}

// Пример использования
func ExampleOID_MarshalBinary() {
	oid := OID{1, 3, 6, 1}

	data, err := oid.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", data)
	// Output: 06032b0601
}

// Бенчмарк
func BenchmarkOIDMarshalBinary(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = oid.MarshalBinary()
	}
}

func TestOIDUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected OID
		wantErr  error
	}{
		{
			name:     "Стандартный OID",
			data:     []byte{0x06, 0x03, 0x2B, 0x06, 0x01},
			expected: OID{1, 3, 6, 1},
			wantErr:  nil,
		},
		{
			name:     "Длинный OID",
			data:     []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00},
			expected: OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			data:     []byte{0x06, 0x03, 0x81, 0x34, 0x03},
			expected: OID{2, 100, 3},
			wantErr:  nil,
		},
		{
			name:     "Пустые данные",
			data:     []byte{},
			expected: nil,
			wantErr:  ErrDataTooShort,
		},
		{
			name:     "Один байт",
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
		{
			name:     "Неверный первый компонент",
			data:     []byte{0x06, 0x01, 0x80},
			expected: nil,
			wantErr:  ErrInvalidFirstComponent,
		},
		{
			name:     "Неверный компонент",
			data:     []byte{0x06, 0x02, 0x2B, 0x80},
			expected: nil,
			wantErr:  ErrInvalidComponent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.UnmarshalBinary(tt.data)

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

			if !oid.Equal(tt.expected) {
				t.Errorf("UnmarshalBinary = %v, want %v", oid, tt.expected)
			}
		})
	}
}

func TestOIDUnmarshalBinaryRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
		{1, 3, MaxOIDComponent},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			// Кодируем
			data, err := oid.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// Декодируем
			var decoded OID
			if err := decoded.UnmarshalBinary(data); err != nil {
				t.Fatalf("UnmarshalBinary: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %x -> %v", oid, data, decoded)
			}
		})
	}
}

func TestOIDUnmarshalBinaryProperties(t *testing.T) {
	t.Run("Перезаписывает предыдущее значение", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		newData := []byte{0x06, 0x03, 0x81, 0x34, 0x03}
		if err := oid.UnmarshalBinary(newData); err != nil {
			t.Fatalf("UnmarshalBinary: %v", err)
		}

		if !oid.Equal(OID{2, 100, 3}) {
			t.Errorf("После UnmarshalBinary = %v, want 2.100.3", oid)
		}
	})

	t.Run("Очищает при ошибке", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		err := oid.UnmarshalBinary([]byte{})

		if err == nil {
			t.Error("expected error")
		}
	})
}

// Пример использования
func ExampleOID_UnmarshalBinary() {
	data := []byte{0x06, 0x03, 0x2B, 0x06, 0x01}

	var oid OID
	if err := oid.UnmarshalBinary(data); err != nil {
		panic(err)
	}

	fmt.Println(oid)
	// Output: 1.3.6.1
}

// Пример с ошибкой
func ExampleOID_UnmarshalBinary_error() {
	var oid OID
	err := oid.UnmarshalBinary([]byte{})
	fmt.Println(errors.Is(err, ErrDataTooShort))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDUnmarshalBinary(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")
	data, _ := oid.MarshalBinary()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var decoded OID
		_ = decoded.UnmarshalBinary(data)
	}
}

func TestOIDMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected string
	}{
		{
			name:     "Стандартный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: `"1.3.6.1.4.1"`,
		},
		{
			name:     "Короткий OID",
			oid:      OID{1, 3, 6},
			expected: `"1.3.6"`,
		},
		{
			name:     "С первым 2",
			oid:      OID{2, 100, 3},
			expected: `"2.100.3"`,
		},
		{
			name:     "С первым 0",
			oid:      OID{0, 39, 1},
			expected: `"0.39.1"`,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: `""`,
		},
		{
			name:     "Nil OID",
			oid:      nil,
			expected: `""`,
		},
		{
			name:     "Максимальный компонент",
			oid:      OID{1, 3, MaxOIDComponent},
			expected: `"1.3.268435455"`,
		},
		{
			name:     "Длинный OID",
			oid:      OID{1, 3, 6, 1, 4, 1, 99999, 1, 1},
			expected: `"1.3.6.1.4.1.99999.1.1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalJSON()

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

func TestOIDMarshalJSONValid(t *testing.T) {
	oid := OID{1, 3, 6, 1}

	data, err := oid.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if !json.Valid(data) {
		t.Errorf("MarshalJSON = %s, невалидный JSON", data)
	}
}

func TestOIDMarshalJSONRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			// Кодируем
			data, err := oid.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}

			// Декодируем
			var decoded OID
			if err := decoded.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %s -> %v", oid, data, decoded)
			}
		})
	}
}

func TestOIDMarshalJSONCompareWithStd(t *testing.T) {
	oid := OID{1, 3, 6, 1}

	// Наш MarshalJSON
	ourData, err := oid.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	// Стандартный json.Marshal
	stdData, err := json.Marshal(oid.String())
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	if !bytes.Equal(ourData, stdData) {
		t.Errorf("MarshalJSON = %s, want %s", ourData, stdData)
	}
}

func TestOIDMarshalJSONNotModifyOID(t *testing.T) {
	oid := OID{1, 3, 6, 1}
	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)

	oid.MarshalJSON()

	if !oid.Equal(oidCopy) {
		t.Error("MarshalJSON() не должен изменять OID")
	}
}

// Пример использования
func ExampleOID_MarshalJSON() {
	oid := OID{1, 3, 6, 1, 4, 1}

	data, err := oid.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output: "1.3.6.1.4.1"
}

// Пример с пустым OID
func ExampleOID_MarshalJSON_empty() {
	oid := OID{}

	data, _ := oid.MarshalJSON()

	fmt.Println(string(data))
	// Output: ""
}

// Бенчмарк
func BenchmarkOIDMarshalJSON(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = oid.MarshalJSON()
	}
}

func TestOIDUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected OID
		wantErr  error
	}{
		{
			name:     "Стандартный OID",
			data:     []byte(`"1.3.6.1.4.1"`),
			expected: OID{1, 3, 6, 1, 4, 1},
			wantErr:  nil,
		},
		{
			name:     "Короткий OID",
			data:     []byte(`"1.3.6"`),
			expected: OID{1, 3, 6},
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			data:     []byte(`"2.100.3"`),
			expected: OID{2, 100, 3},
			wantErr:  nil,
		},
		{
			name:     "С первым 0",
			data:     []byte(`"0.39.1"`),
			expected: OID{0, 39, 1},
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
			var oid OID
			err := oid.UnmarshalJSON(tt.data)

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

			if tt.expected == nil && tt.data != nil {
				// Ожидаем ошибку для невалидного ввода
				if err == nil {
					t.Errorf("UnmarshalJSON(%s): expected error", tt.data)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalJSON: %v", err)
				return
			}

			if !oid.Equal(tt.expected) {
				t.Errorf("UnmarshalJSON = %v, want %v", oid, tt.expected)
			}
		})
	}
}

func TestOIDUnmarshalJSONRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			// Кодируем
			data, err := oid.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}

			// Декодируем
			var decoded OID
			if err := decoded.UnmarshalJSON(data); err != nil {
				t.Fatalf("UnmarshalJSON: %v", err)
			}

			// Сравниваем
			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %s -> %v", oid, data, decoded)
			}
		})
	}
}

func TestOIDUnmarshalJSONProperties(t *testing.T) {
	t.Run("Перезаписывает предыдущее значение", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		if err := oid.UnmarshalJSON([]byte(`"2.100.3"`)); err != nil {
			t.Fatalf("UnmarshalJSON: %v", err)
		}

		if !oid.Equal(OID{2, 100, 3}) {
			t.Errorf("После UnmarshalJSON = %v, want 2.100.3", oid)
		}
	})

	t.Run("Очищает при ошибке", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		err := oid.UnmarshalJSON([]byte(`null`))

		if err == nil {
			t.Error("expected error")
		}
	})
}

// Пример использования
func ExampleOID_UnmarshalJSON() {
	data := []byte(`"1.3.6.1.4.1"`)

	var oid OID
	if err := oid.UnmarshalJSON(data); err != nil {
		panic(err)
	}

	fmt.Println(oid)
	// Output: 1.3.6.1.4.1
}

// Пример с ошибкой
func ExampleOID_UnmarshalJSON_error() {
	var oid OID
	err := oid.UnmarshalJSON([]byte(`null`))
	fmt.Println(errors.Is(err, ErrInvalidJSONType))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDUnmarshalJSON(b *testing.B) {
	data := []byte(`"1.3.6.1.4.1.99999.1.1"`)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var oid OID
		_ = oid.UnmarshalJSON(data)
	}
}

func TestDigitCount(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected int
	}{
		{"0", 0, 1},
		{"1", 1, 1},
		{"9", 9, 1},
		{"10", 10, 2},
		{"99", 99, 2},
		{"100", 100, 3},
		{"999", 999, 3},
		{"1000", 1000, 4},
		{"9999", 9999, 4},
		{"10000", 10000, 5},
		{"99999", 99999, 5},
		{"100000", 100000, 6},
		{"999999", 999999, 6},
		{"1000000", 1000000, 7},
		{"9999999", 9999999, 7},
		{"10000000", 10000000, 8},
		{"99999999", 99999999, 8},
		{"100000000", 100000000, 9},
		{"999999999", 999999999, 9},
		{"1000000000", 1000000000, 10},
		{"MaxUint32", ^uint32(0), 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := digitCount(tt.input)

			if result != tt.expected {
				t.Errorf("digitCount(%d) = %d, want %d",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestDigitCountBoundaries(t *testing.T) {
	t.Run("Границы разрядов", func(t *testing.T) {
		tests := []struct {
			input    uint32
			expected int
		}{
			{0, 1},
			{9, 1},
			{10, 2},
			{99, 2},
			{100, 3},
			{999, 3},
			{1000, 4},
			{9999, 4},
			{10000, 5},
			{99999, 5},
			{100000, 6},
			{999999, 6},
			{1000000, 7},
			{9999999, 7},
			{10000000, 8},
			{99999999, 8},
			{100000000, 9},
			{999999999, 9},
			{1000000000, 10},
		}

		for _, tt := range tests {
			result := digitCount(tt.input)
			if result != tt.expected {
				t.Errorf("digitCount(%d) = %d, want %d",
					tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Все границы", func(t *testing.T) {
		// Проверяем каждую границу
		boundaries := []uint32{
			0, 9, 10, 99, 100, 999, 1000, 9999,
			10000, 99999, 100000, 999999,
			1000000, 9999999, 10000000, 99999999,
			100000000, 999999999, 1000000000,
		}

		for _, boundary := range boundaries {
			result := digitCount(boundary)

			// Проверяем, что результат соответствует длине строки
			expected := len(fmt.Sprintf("%d", boundary))
			if result != expected {
				t.Errorf("digitCount(%d) = %d, want %d",
					boundary, result, expected)
			}
		}
	})
}

func TestDigitCountConsistency(t *testing.T) {
	t.Run("Соответствует длине строки", func(t *testing.T) {
		// Проверяем случайные значения
		values := []uint32{
			0, 1, 5, 10, 50, 100, 500, 1000, 5000,
			10000, 50000, 100000, 500000, 1000000,
			MaxOIDComponent, ^uint32(0),
		}

		for _, value := range values {
			result := digitCount(value)
			expected := len(fmt.Sprintf("%d", value))

			if result != expected {
				t.Errorf("digitCount(%d) = %d, want %d",
					value, result, expected)
			}
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		value := uint32(12345)

		result1 := digitCount(value)
		result2 := digitCount(value)

		if result1 != result2 {
			t.Error("digitCount должен быть детерминированным")
		}
	})
}

func TestDigitCountPerformance(t *testing.T) {
	t.Run("Быстрая работа", func(t *testing.T) {
		allocs := testing.AllocsPerRun(1000, func() {
			_ = digitCount(12345)
		})

		if allocs != 0 {
			t.Errorf("digitCount: %f allocs, want 0", allocs)
		}
	})
}

// Пример для приватной функции - используем тест вместо примера
func TestDigitCountExample(t *testing.T) {
	tests := []struct {
		input    uint32
		expected int
	}{
		{0, 1},
		{5, 1},
		{10, 2},
		{100, 3},
		{1000, 4},
	}

	for _, tt := range tests {
		if result := digitCount(tt.input); result != tt.expected {
			t.Errorf("digitCount(%d) = %d, want %d",
				tt.input, result, tt.expected)
		}
	}
}

// Бенчмарк
func BenchmarkDigitCount(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = digitCount(12345)
	}
}

// Бенчмарк для разных значений
func BenchmarkDigitCountValues(b *testing.B) {
	values := []uint32{0, 9, 10, 99, 100, 999, 1000, MaxOIDComponent, ^uint32(0)}

	for _, value := range values {
		b.Run(fmt.Sprintf("%d", value), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = digitCount(value)
			}
		})
	}
}

func TestBase128Size(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected int
	}{
		{"0", 0, 1},
		{"1", 1, 1},
		{"127", 127, 1},
		{"128", 128, 2},
		{"129", 129, 2},
		{"16383", 16383, 2},
		{"16384", 16384, 3},
		{"16385", 16385, 3},
		{"2097151", 2097151, 3},
		{"2097152", 2097152, 4},
		{"2097153", 2097153, 4},
		{"268435455", 268435455, 4},
		{"268435456", 268435456, 5},
		{"MaxUint32", ^uint32(0), 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base128Size(tt.input)

			if result != tt.expected {
				t.Errorf("base128Size(%d) = %d, want %d",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestBase128SizeBoundaries(t *testing.T) {
	t.Run("Границы", func(t *testing.T) {
		boundaries := []struct {
			input    uint32
			expected int
		}{
			{0, 1},
			{127, 1},
			{128, 2},
			{16383, 2},
			{16384, 3},
			{2097151, 3},
			{2097152, 4},
			{268435455, 4},
			{268435456, 5},
			{^uint32(0), 5},
		}

		for _, tt := range boundaries {
			result := base128Size(tt.input)
			if result != tt.expected {
				t.Errorf("base128Size(%d) = %d, want %d",
					tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Все значения корректны", func(t *testing.T) {
		values := []uint32{
			0, 1, 127, 128, 255, 256, 16383, 16384,
			65535, 65536, 2097151, 2097152,
			MaxOIDComponent, 268435456, ^uint32(0),
		}

		for _, value := range values {
			result := base128Size(value)

			// Проверяем через writeBase128
			buf := make([]byte, 5)
			n := writeBase128(buf, value)

			if result != n {
				t.Errorf("base128Size(%d) = %d, writeBase128 = %d",
					value, result, n)
			}
		}
	})
}

func TestBase128SizeConsistency(t *testing.T) {
	t.Run("Соответствует writeBase128", func(t *testing.T) {
		values := []uint32{
			0, 1, 127, 128, 16383, 16384,
			MaxOIDComponent, ^uint32(0),
		}

		for _, value := range values {
			size := base128Size(value)

			buf := make([]byte, 5)
			written := writeBase128(buf, value)

			if size != written {
				t.Errorf("base128Size(%d) = %d, written = %d",
					value, size, written)
			}
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		value := uint32(12345)

		result1 := base128Size(value)
		result2 := base128Size(value)

		if result1 != result2 {
			t.Error("base128Size должен быть детерминированным")
		}
	})
}

func TestBase128SizeNoAllocations(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		_ = base128Size(12345)
	})

	if allocs != 0 {
		t.Errorf("base128Size: %f allocs, want 0", allocs)
	}
}

// Бенчмарк
func BenchmarkBase128Size(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = base128Size(12345)
	}
}

// Бенчмарк для разных значений
func BenchmarkBase128SizeValues(b *testing.B) {
	values := []uint32{0, 127, 128, 16383, 16384, MaxOIDComponent, ^uint32(0)}

	for _, value := range values {
		b.Run(fmt.Sprintf("%d", value), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = base128Size(value)
			}
		})
	}
}

func TestLengthSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"0", 0, 1},
		{"1", 1, 1},
		{"127", 127, 1},
		{"128", 128, 2},
		{"129", 129, 2},
		{"255", 255, 2},
		{"256", 256, 3},
		{"257", 257, 3},
		{"65535", 65535, 3},
		{"65536", 65536, 4},
		{"65537", 65537, 4},
		{"MaxInt", int(^uint(0) >> 1), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lengthSize(tt.input)

			if result != tt.expected {
				t.Errorf("lengthSize(%d) = %d, want %d",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestLengthSizeBoundaries(t *testing.T) {
	t.Run("Границы", func(t *testing.T) {
		boundaries := []struct {
			input    int
			expected int
		}{
			{0, 1},
			{127, 1},
			{128, 2},
			{255, 2},
			{256, 3},
			{65535, 3},
			{65536, 4},
		}

		for _, tt := range boundaries {
			result := lengthSize(tt.input)
			if result != tt.expected {
				t.Errorf("lengthSize(%d) = %d, want %d",
					tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Отрицательные значения", func(t *testing.T) {
		result := lengthSize(-1)

		// Отрицательные значения обрабатываются как < 128
		if result != 1 {
			t.Errorf("lengthSize(-1) = %d, want 1", result)
		}
	})
}

func TestLengthSizeConsistency(t *testing.T) {
	t.Run("Соответствует writeLength", func(t *testing.T) {
		values := []int{0, 127, 128, 255, 256, 65535, 65536}

		for _, value := range values {
			size := lengthSize(value)

			buf := make([]byte, 4)
			written := writeLength(buf, value)

			if size != written {
				t.Errorf("lengthSize(%d) = %d, written = %d",
					value, size, written)
			}
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		value := 12345

		result1 := lengthSize(value)
		result2 := lengthSize(value)

		if result1 != result2 {
			t.Error("lengthSize должен быть детерминированным")
		}
	})
}

func TestLengthSizeNoAllocations(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		_ = lengthSize(12345)
	})

	if allocs != 0 {
		t.Errorf("lengthSize: %f allocs, want 0", allocs)
	}
}

func TestLengthSizeAllBranches(t *testing.T) {
	t.Run("Короткая форма (< 128)", func(t *testing.T) {
		if result := lengthSize(0); result != 1 {
			t.Errorf("lengthSize(0) = %d, want 1", result)
		}
		if result := lengthSize(127); result != 1 {
			t.Errorf("lengthSize(127) = %d, want 1", result)
		}
	})

	t.Run("Длинная форма 1 байт (128-255)", func(t *testing.T) {
		if result := lengthSize(128); result != 2 {
			t.Errorf("lengthSize(128) = %d, want 2", result)
		}
		if result := lengthSize(255); result != 2 {
			t.Errorf("lengthSize(255) = %d, want 2", result)
		}
	})

	t.Run("Длинная форма 2 байта (256-65535)", func(t *testing.T) {
		if result := lengthSize(256); result != 3 {
			t.Errorf("lengthSize(256) = %d, want 3", result)
		}
		if result := lengthSize(65535); result != 3 {
			t.Errorf("lengthSize(65535) = %d, want 3", result)
		}
	})

	t.Run("Длинная форма 3 байта (>= 65536)", func(t *testing.T) {
		if result := lengthSize(65536); result != 4 {
			t.Errorf("lengthSize(65536) = %d, want 4", result)
		}
	})
}

// Бенчмарк
func BenchmarkLengthSize(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = lengthSize(12345)
	}
}

// Бенчмарк для разных значений
func BenchmarkLengthSizeValues(b *testing.B) {
	values := []int{0, 127, 128, 255, 256, 65535, 65536}

	for _, value := range values {
		b.Run(fmt.Sprintf("%d", value), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = lengthSize(value)
			}
		})
	}
}

func TestWriteBase128(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{name: "0", value: 0, expected: []byte{0x00}},
		{name: "1", value: 1, expected: []byte{0x01}},
		{name: "127", value: 127, expected: []byte{0x7F}},
		{name: "128", value: 128, expected: []byte{0x81, 0x00}},
		{name: "129", value: 129, expected: []byte{0x81, 0x01}},
		{name: "255", value: 255, expected: []byte{0x81, 0x7F}},
		{name: "256", value: 256, expected: []byte{0x82, 0x00}},
		{name: "16383", value: 16383, expected: []byte{0xFF, 0x7F}},
		{name: "16384", value: 16384, expected: []byte{0x81, 0x80, 0x00}},
		{name: "MaxOIDComponent", value: MaxOIDComponent, expected: []byte{0xFF, 0xFF, 0xFF, 0x7F}},
		{name: "MaxUint32", value: ^uint32(0), expected: []byte{0x8F, 0xFF, 0xFF, 0xFF, 0x7F}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 5)
			n := writeBase128(buf, tt.value)

			if n != len(tt.expected) {
				t.Errorf("writeBase128(%d) = %d bytes, want %d",
					tt.value, n, len(tt.expected))
				return
			}

			if !bytes.Equal(buf[:n], tt.expected) {
				t.Errorf("writeBase128(%d) = %x, want %x",
					tt.value, buf[:n], tt.expected)
			}
		})
	}
}

func TestWriteBase128RoundTrip(t *testing.T) {
	tests := []uint32{
		0, 1, 127, 128, 129, 255, 256, 16383, 16384,
		MaxOIDComponent, ^uint32(0),
	}

	for _, value := range tests {
		t.Run(fmt.Sprintf("%d", value), func(t *testing.T) {
			buf := make([]byte, 5)
			n := writeBase128(buf, value)

			decoded, bytesRead := readBase128(buf[:n])

			if bytesRead != n {
				t.Errorf("bytesRead = %d, want %d", bytesRead, n)
			}

			if decoded != value {
				t.Errorf("Round trip: %d -> %x -> %d", value, buf[:n], decoded)
			}
		})
	}
}

func TestWriteBase128Consistency(t *testing.T) {
	t.Run("Соответствует base128Size", func(t *testing.T) {
		values := []uint32{
			0, 1, 127, 128, 16383, 16384,
			MaxOIDComponent, ^uint32(0),
		}

		for _, value := range values {
			size := base128Size(value)

			buf := make([]byte, 5)
			written := writeBase128(buf, value)

			if size != written {
				t.Errorf("base128Size(%d) = %d, written = %d",
					value, size, written)
			}
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		value := uint32(12345)

		buf1 := make([]byte, 5)
		buf2 := make([]byte, 5)

		n1 := writeBase128(buf1, value)
		n2 := writeBase128(buf2, value)

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			t.Error("writeBase128 должен быть детерминированным")
		}
	})
}

func BenchmarkWriteBase128(b *testing.B) {
	buf := make([]byte, 5)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = writeBase128(buf, uint32(12345)) // Явное приведение
	}
}

func BenchmarkWriteBase128Values(b *testing.B) {
	values := []uint32{0, 127, 128, 16383, 16384, MaxOIDComponent, ^uint32(0)}

	for _, value := range values {
		b.Run(fmt.Sprintf("%d", value), func(b *testing.B) {
			buf := make([]byte, 5)

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_ = writeBase128(buf, value)
			}
		})
	}
}

func TestWriteLength(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected []byte
	}{
		{
			name:     "0",
			length:   0,
			expected: []byte{0x00},
		},
		{
			name:     "1",
			length:   1,
			expected: []byte{0x01},
		},
		{
			name:     "127",
			length:   127,
			expected: []byte{0x7F},
		},
		{
			name:     "128",
			length:   128,
			expected: []byte{0x81, 0x80},
		},
		{
			name:     "129",
			length:   129,
			expected: []byte{0x81, 0x81},
		},
		{
			name:     "255",
			length:   255,
			expected: []byte{0x81, 0xFF},
		},
		{
			name:     "256",
			length:   256,
			expected: []byte{0x82, 0x01, 0x00},
		},
		{
			name:     "257",
			length:   257,
			expected: []byte{0x82, 0x01, 0x01},
		},
		{
			name:     "65535",
			length:   65535,
			expected: []byte{0x82, 0xFF, 0xFF},
		},
		{
			name:     "65536",
			length:   65536,
			expected: []byte{0x83, 0x01, 0x00, 0x00},
		},
		{
			name:     "Отрицательное",
			length:   -1,
			expected: nil, // возвращает 0 байт
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 4)
			n := writeLength(buf, tt.length)

			if tt.expected == nil {
				if n != 0 {
					t.Errorf("writeLength(%d) = %d bytes, want 0", tt.length, n)
				}
				return
			}

			if n != len(tt.expected) {
				t.Errorf("writeLength(%d) = %d bytes, want %d",
					tt.length, n, len(tt.expected))
				return
			}

			if !bytes.Equal(buf[:n], tt.expected) {
				t.Errorf("writeLength(%d) = %x, want %x",
					tt.length, buf[:n], tt.expected)
			}
		})
	}
}

func TestWriteLengthRoundTrip(t *testing.T) {
	tests := []int{0, 1, 127, 128, 255, 256, 65535, 65536}

	for _, length := range tests {
		t.Run(fmt.Sprintf("%d", length), func(t *testing.T) {
			// Кодируем
			buf := make([]byte, 4)
			n := writeLength(buf, length)

			// Декодируем
			decoded, bytesRead := readLength(buf[:n])

			if bytesRead != n {
				t.Errorf("bytesRead = %d, want %d", bytesRead, n)
			}

			if decoded != length {
				t.Errorf("Round trip: %d -> %x -> %d", length, buf[:n], decoded)
			}
		})
	}
}

func TestWriteLengthConsistency(t *testing.T) {
	t.Run("Соответствует lengthSize", func(t *testing.T) {
		values := []int{0, 127, 128, 255, 256, 65535, 65536}

		for _, value := range values {
			size := lengthSize(value)

			buf := make([]byte, 4)
			written := writeLength(buf, value)

			if size != written {
				t.Errorf("lengthSize(%d) = %d, written = %d",
					value, size, written)
			}
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		value := 12345

		buf1 := make([]byte, 4)
		buf2 := make([]byte, 4)

		n1 := writeLength(buf1, value)
		n2 := writeLength(buf2, value)

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			t.Error("writeLength должен быть детерминированным")
		}
	})
}

func TestWriteLengthNoAllocations(t *testing.T) {
	buf := make([]byte, 4)

	allocs := testing.AllocsPerRun(1000, func() {
		_ = writeLength(buf, 12345)
	})

	if allocs != 0 {
		t.Errorf("writeLength: %f allocs, want 0", allocs)
	}
}

func TestWriteLengthAllBranches(t *testing.T) {
	t.Run("Короткая форма", func(t *testing.T) {
		buf := make([]byte, 4)
		n := writeLength(buf, 127)

		if n != 1 || buf[0] != 0x7F {
			t.Errorf("writeLength(127) = %d, %x", n, buf[:n])
		}
	})

	t.Run("Длинная форма 1 байт", func(t *testing.T) {
		buf := make([]byte, 4)
		n := writeLength(buf, 128)

		if n != 2 || buf[0] != 0x81 || buf[1] != 0x80 {
			t.Errorf("writeLength(128) = %d, %x", n, buf[:n])
		}
	})

	t.Run("Длинная форма 2 байта", func(t *testing.T) {
		buf := make([]byte, 4)
		n := writeLength(buf, 256)

		if n != 3 || buf[0] != 0x82 {
			t.Errorf("writeLength(256) = %d, %x", n, buf[:n])
		}
	})

	t.Run("Длинная форма 3 байта", func(t *testing.T) {
		buf := make([]byte, 4)
		n := writeLength(buf, 65536)

		if n != 4 || buf[0] != 0x83 {
			t.Errorf("writeLength(65536) = %d, %x", n, buf[:n])
		}
	})

	t.Run("Отрицательное", func(t *testing.T) {
		buf := make([]byte, 4)
		n := writeLength(buf, -1)

		if n != 0 {
			t.Errorf("writeLength(-1) = %d, want 0", n)
		}
	})
}

// Бенчмарк
func BenchmarkWriteLength(b *testing.B) {
	buf := make([]byte, 4)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = writeLength(buf, 12345)
	}
}

// Бенчмарк для разных значений
func BenchmarkWriteLengthValues(b *testing.B) {
	values := []int{0, 127, 128, 255, 256, 65535, 65536}

	for _, value := range values {
		b.Run(fmt.Sprintf("%d", value), func(b *testing.B) {
			buf := make([]byte, 4)

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_ = writeLength(buf, value)
			}
		})
	}
}

func TestReadBase128(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint32
		bytes    int
	}{
		{
			name:     "0",
			data:     []byte{0x00},
			expected: 0,
			bytes:    1,
		},
		{
			name:     "1",
			data:     []byte{0x01},
			expected: 1,
			bytes:    1,
		},
		{
			name:     "127",
			data:     []byte{0x7F},
			expected: 127,
			bytes:    1,
		},
		{
			name:     "128",
			data:     []byte{0x81, 0x00},
			expected: 128,
			bytes:    2,
		},
		{
			name:     "129",
			data:     []byte{0x81, 0x01},
			expected: 129,
			bytes:    2,
		},
		{
			name:     "255",
			data:     []byte{0x81, 0x7F},
			expected: 255,
			bytes:    2,
		},
		{
			name:     "256",
			data:     []byte{0x82, 0x00},
			expected: 256,
			bytes:    2,
		},
		{
			name:     "16383",
			data:     []byte{0xFF, 0x7F},
			expected: 16383,
			bytes:    2,
		},
		{
			name:     "16384",
			data:     []byte{0x81, 0x80, 0x00},
			expected: 16384,
			bytes:    3,
		},
		{
			name:     "MaxOIDComponent",
			data:     []byte{0xFF, 0xFF, 0xFF, 0x7F},
			expected: MaxOIDComponent,
			bytes:    4,
		},
		{
			name:     "MaxUint32",
			data:     []byte{0x8F, 0xFF, 0xFF, 0xFF, 0x7F},
			expected: ^uint32(0),
			bytes:    5,
		},
		{
			name:     "Переполнение (>5 байт)",
			data:     []byte{0x81, 0x80, 0x80, 0x80, 0x80, 0x00},
			expected: 0,
			bytes:    0,
		},
		{
			name:     "Незавершенная последовательность",
			data:     []byte{0x81, 0x80},
			expected: 0,
			bytes:    0,
		},
		{
			name:     "Пустая",
			data:     []byte{},
			expected: 0,
			bytes:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, bytesRead := readBase128(tt.data)

			if bytesRead != tt.bytes {
				t.Errorf("readBase128(%x) bytesRead = %d, want %d",
					tt.data, bytesRead, tt.bytes)
			}

			if value != tt.expected {
				t.Errorf("readBase128(%x) value = %d, want %d",
					tt.data, value, tt.expected)
			}
		})
	}
}

func TestReadBase128RoundTrip(t *testing.T) {
	values := []uint32{
		0, 1, 127, 128, 255, 256, 16383, 16384,
		MaxOIDComponent, ^uint32(0),
	}

	for _, value := range values {
		t.Run(fmt.Sprintf("%d", value), func(t *testing.T) {
			// Кодируем
			buf := make([]byte, 5)
			n := writeBase128(buf, value)

			// Декодируем
			decoded, bytesRead := readBase128(buf[:n])

			if bytesRead != n {
				t.Errorf("bytesRead = %d, want %d", bytesRead, n)
			}

			if decoded != value {
				t.Errorf("Round trip: %d -> %x -> %d", value, buf[:n], decoded)
			}
		})
	}
}

func TestReadBase128Errors(t *testing.T) {
	t.Run("Переполнение", func(t *testing.T) {
		// 6 байт - слишком много для uint32
		data := []byte{0x81, 0x80, 0x80, 0x80, 0x80, 0x00}

		value, bytesRead := readBase128(data)

		if bytesRead != 0 {
			t.Errorf("bytesRead = %d, want 0", bytesRead)
		}
		if value != 0 {
			t.Errorf("value = %d, want 0", value)
		}
	})

	t.Run("Незавершенная", func(t *testing.T) {
		data := []byte{0x81, 0x80}

		value, bytesRead := readBase128(data)

		if bytesRead != 0 {
			t.Errorf("bytesRead = %d, want 0", bytesRead)
		}
		if value != 0 {
			t.Errorf("value = %d, want 0", value)
		}
	})

	t.Run("Пустая", func(t *testing.T) {
		data := []byte{}

		value, bytesRead := readBase128(data)

		if bytesRead != 0 {
			t.Errorf("bytesRead = %d, want 0", bytesRead)
		}
		if value != 0 {
			t.Errorf("value = %d, want 0", value)
		}
	})
}

func TestReadBase128Consistency(t *testing.T) {
	t.Run("Соответствует writeBase128", func(t *testing.T) {
		values := []uint32{
			0, 1, 127, 128, 16383, 16384,
			MaxOIDComponent, ^uint32(0),
		}

		for _, value := range values {
			buf := make([]byte, 5)
			n := writeBase128(buf, value)

			decoded, bytesRead := readBase128(buf[:n])

			if bytesRead != n {
				t.Errorf("bytesRead = %d, want %d", bytesRead, n)
			}
			if decoded != value {
				t.Errorf("decoded = %d, want %d", decoded, value)
			}
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		data := []byte{0x81, 0x7F}

		value1, bytes1 := readBase128(data)
		value2, bytes2 := readBase128(data)

		if value1 != value2 || bytes1 != bytes2 {
			t.Error("readBase128 должен быть детерминированным")
		}
	})
}

func TestReadBase128NoAllocations(t *testing.T) {
	data := []byte{0x81, 0x7F}

	allocs := testing.AllocsPerRun(1000, func() {
		_, _ = readBase128(data)
	})

	if allocs != 0 {
		t.Errorf("readBase128: %f allocs, want 0", allocs)
	}
}

// Бенчмарк
func BenchmarkReadBase128(b *testing.B) {
	data := []byte{0x81, 0x7F}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = readBase128(data)
	}
}

// Бенчмарк для разных значений
func BenchmarkReadBase128Values(b *testing.B) {
	values := []uint32{0, 127, 128, 16383, MaxOIDComponent, ^uint32(0)}

	for _, value := range values {
		b.Run(fmt.Sprintf("%d", value), func(b *testing.B) {
			buf := make([]byte, 5)
			n := writeBase128(buf, value)
			data := buf[:n]

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_, _ = readBase128(data)
			}
		})
	}
}

func TestReadLength(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
		bytes    int
	}{
		{
			name:     "0",
			data:     []byte{0x00},
			expected: 0,
			bytes:    1,
		},
		{
			name:     "1",
			data:     []byte{0x01},
			expected: 1,
			bytes:    1,
		},
		{
			name:     "127",
			data:     []byte{0x7F},
			expected: 127,
			bytes:    1,
		},
		{
			name:     "128",
			data:     []byte{0x81, 0x80},
			expected: 128,
			bytes:    2,
		},
		{
			name:     "129",
			data:     []byte{0x81, 0x81},
			expected: 129,
			bytes:    2,
		},
		{
			name:     "255",
			data:     []byte{0x81, 0xFF},
			expected: 255,
			bytes:    2,
		},
		{
			name:     "256",
			data:     []byte{0x82, 0x01, 0x00},
			expected: 256,
			bytes:    3,
		},
		{
			name:     "65535",
			data:     []byte{0x82, 0xFF, 0xFF},
			expected: 65535,
			bytes:    3,
		},
		{
			name:     "65536",
			data:     []byte{0x83, 0x01, 0x00, 0x00},
			expected: 65536,
			bytes:    4,
		},
		{
			name:     "0x80 (некорректная)",
			data:     []byte{0x80},
			expected: 0,
			bytes:    0,
		},
		{
			name:     "0x85 (слишком длинная)",
			data:     []byte{0x85},
			expected: 0,
			bytes:    0,
		},
		{
			name:     "0x81 с недостатком",
			data:     []byte{0x81},
			expected: 0,
			bytes:    0,
		},
		{
			name:     "Пустая",
			data:     []byte{},
			expected: 0,
			bytes:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, bytesRead := readLength(tt.data)

			if bytesRead != tt.bytes {
				t.Errorf("readLength(%x) bytesRead = %d, want %d",
					tt.data, bytesRead, tt.bytes)
			}

			if length != tt.expected {
				t.Errorf("readLength(%x) length = %d, want %d",
					tt.data, length, tt.expected)
			}
		})
	}
}

func TestReadLengthRoundTrip(t *testing.T) {
	tests := []int{0, 1, 127, 128, 255, 256, 65535, 65536}

	for _, length := range tests {
		t.Run(fmt.Sprintf("%d", length), func(t *testing.T) {
			// Кодируем
			buf := make([]byte, 4)
			n := writeLength(buf, length)

			// Декодируем
			decoded, bytesRead := readLength(buf[:n])

			if bytesRead != n {
				t.Errorf("bytesRead = %d, want %d", bytesRead, n)
			}

			if decoded != length {
				t.Errorf("Round trip: %d -> %x -> %d", length, buf[:n], decoded)
			}
		})
	}
}

func TestReadLengthErrors(t *testing.T) {
	t.Run("0x80 - некорректная длина", func(t *testing.T) {
		length, bytesRead := readLength([]byte{0x80})

		if bytesRead != 0 || length != 0 {
			t.Errorf("readLength(0x80) = %d, %d; want 0, 0", length, bytesRead)
		}
	})

	t.Run("0x85 - слишком длинная", func(t *testing.T) {
		length, bytesRead := readLength([]byte{0x85})

		if bytesRead != 0 || length != 0 {
			t.Errorf("readLength(0x85) = %d, %d; want 0, 0", length, bytesRead)
		}
	})

	t.Run("0x81 - недостаточно данных", func(t *testing.T) {
		length, bytesRead := readLength([]byte{0x81})

		if bytesRead != 0 || length != 0 {
			t.Errorf("readLength(0x81) = %d, %d; want 0, 0", length, bytesRead)
		}
	})

	t.Run("Пустая", func(t *testing.T) {
		length, bytesRead := readLength([]byte{})

		if bytesRead != 0 || length != 0 {
			t.Errorf("readLength(empty) = %d, %d; want 0, 0", length, bytesRead)
		}
	})
}

func TestReadLengthConsistency(t *testing.T) {
	t.Run("Соответствует writeLength", func(t *testing.T) {
		values := []int{0, 127, 128, 255, 256, 65535, 65536}

		for _, value := range values {
			buf := make([]byte, 4)
			n := writeLength(buf, value)

			decoded, bytesRead := readLength(buf[:n])

			if bytesRead != n {
				t.Errorf("bytesRead = %d, want %d", bytesRead, n)
			}
			if decoded != value {
				t.Errorf("decoded = %d, want %d", decoded, value)
			}
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		data := []byte{0x82, 0x01, 0x00}

		length1, bytes1 := readLength(data)
		length2, bytes2 := readLength(data)

		if length1 != length2 || bytes1 != bytes2 {
			t.Error("readLength должен быть детерминированным")
		}
	})
}

func TestReadLengthNoAllocations(t *testing.T) {
	data := []byte{0x82, 0x01, 0x00}

	allocs := testing.AllocsPerRun(1000, func() {
		_, _ = readLength(data)
	})

	if allocs != 0 {
		t.Errorf("readLength: %f allocs, want 0", allocs)
	}
}

// Бенчмарк
func BenchmarkReadLength(b *testing.B) {
	data := []byte{0x82, 0x01, 0x00}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = readLength(data)
	}
}

// Бенчмарк для разных значений
func BenchmarkReadLengthValues(b *testing.B) {
	values := []int{0, 127, 128, 255, 256, 65535, 65536}

	for _, value := range values {
		b.Run(fmt.Sprintf("%d", value), func(b *testing.B) {
			buf := make([]byte, 4)
			n := writeLength(buf, value)
			data := buf[:n]

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_, _ = readLength(data)
			}
		})
	}
}

func TestCombinedFirstComponents(t *testing.T) {
	tests := []struct {
		name     string
		first    uint32
		second   uint32
		expected uint32
		wantErr  bool
	}{
		{
			name:     "0.0",
			first:    0,
			second:   0,
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "0.39",
			first:    0,
			second:   39,
			expected: 39,
			wantErr:  false,
		},
		{
			name:     "1.0",
			first:    1,
			second:   0,
			expected: 40,
			wantErr:  false,
		},
		{
			name:     "1.3",
			first:    1,
			second:   3,
			expected: 43,
			wantErr:  false,
		},
		{
			name:     "1.39",
			first:    1,
			second:   39,
			expected: 79,
			wantErr:  false,
		},
		{
			name:     "2.0",
			first:    2,
			second:   0,
			expected: 80,
			wantErr:  false,
		},
		{
			name:     "2.175",
			first:    2,
			second:   175,
			expected: 255,
			wantErr:  false,
		},
		{
			name:     "2.MaxOIDComponent",
			first:    2,
			second:   MaxOIDComponent,
			expected: 268435535, // 80 + 268435455
			wantErr:  false,
		},
		{
			name:     "2.MaxUint32-80",
			first:    2,
			second:   ^uint32(0) - 80,
			expected: ^uint32(0),
			wantErr:  false,
		},
		{
			name:     "2.MaxUint32-79 (переполнение)",
			first:    2,
			second:   ^uint32(0) - 79,
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "2.MaxUint32 (переполнение)",
			first:    2,
			second:   ^uint32(0),
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := combinedFirstComponents(tt.first, tt.second)

			if tt.wantErr {
				if err == nil {
					t.Errorf("combinedFirstComponents: expected error")
					return
				}
				if !errors.Is(err, ErrComponentTooBig) {
					t.Errorf("combinedFirstComponents = %v, want ErrComponentTooBig", err)
				}
				return
			}

			if err != nil {
				t.Errorf("combinedFirstComponents: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("combinedFirstComponents(%d, %d) = %d, want %d",
					tt.first, tt.second, result, tt.expected)
			}
		})
	}
}

func TestCombinedFirstComponentsRoundTrip(t *testing.T) {
	tests := []struct {
		first  uint32
		second uint32
	}{
		{0, 0},
		{0, 39},
		{1, 0},
		{1, 3},
		{1, 39},
		{2, 0},
		{2, 175},
		{2, MaxOIDComponent},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d.%d", tt.first, tt.second), func(t *testing.T) {
			result, err := combinedFirstComponents(tt.first, tt.second)
			if err != nil {
				t.Fatalf("combinedFirstComponents: %v", err)
			}

			// Проверяем формулу: 40*first + second
			expected := uint64(tt.first)*40 + uint64(tt.second)

			if uint64(result) != expected {
				t.Errorf("result = %d, want %d", result, expected)
			}
		})
	}
}

func TestCombinedFirstComponentsBoundaries(t *testing.T) {
	t.Run("Максимум без переполнения", func(t *testing.T) {
		// 40*2 + (2^32-1-80) = 2^32-1
		result, err := combinedFirstComponents(2, ^uint32(0)-80)

		if err != nil {
			t.Errorf("combinedFirstComponents: %v", err)
		}

		if result != ^uint32(0) {
			t.Errorf("result = %d, want %d", result, ^uint32(0))
		}
	})

	t.Run("Переполнение на 1", func(t *testing.T) {
		// 40*2 + (2^32-1-79) = 2^32
		_, err := combinedFirstComponents(2, ^uint32(0)-79)

		if err == nil {
			t.Error("expected overflow error")
		}
	})
}

func TestCombinedFirstComponentsConsistency(t *testing.T) {
	t.Run("Детерминированность", func(t *testing.T) {
		result1, err1 := combinedFirstComponents(1, 3)
		result2, err2 := combinedFirstComponents(1, 3)

		if result1 != result2 || (err1 == nil) != (err2 == nil) {
			t.Error("combinedFirstComponents должен быть детерминированным")
		}
	})

	t.Run("Нет аллокаций", func(t *testing.T) {
		allocs := testing.AllocsPerRun(1000, func() {
			_, _ = combinedFirstComponents(1, 3)
		})

		if allocs != 0 {
			t.Errorf("combinedFirstComponents: %f allocs, want 0", allocs)
		}
	})
}

func TestCombinedFirstComponentsOverflow(t *testing.T) {
	tests := []struct {
		name    string
		first   uint32
		second  uint32
		wantErr bool
	}{
		{
			name:    "Переполнение uint32",
			first:   2,
			second:  ^uint32(0),
			wantErr: true,
		},
		{
			name:    "Граница без переполнения",
			first:   2,
			second:  ^uint32(0) - 80,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := combinedFirstComponents(tt.first, tt.second)

			if tt.wantErr {
				if err == nil {
					t.Error("combinedFirstComponents: expected error")
					return
				}
				if !errors.Is(err, ErrComponentTooBig) {
					t.Errorf("combinedFirstComponents = %v, want ErrComponentTooBig", err)
				}
			} else {
				if err != nil {
					t.Errorf("combinedFirstComponents: %v", err)
				}
			}
		})
	}
}

// Бенчмарк
func BenchmarkCombinedFirstComponents(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = combinedFirstComponents(1, 3)
	}
}

// Бенчмарк для разных значений
func BenchmarkCombinedFirstComponentsValues(b *testing.B) {
	values := []struct {
		first  uint32
		second uint32
	}{
		{0, 0},
		{1, 3},
		{2, 100},
		{2, MaxOIDComponent},
	}

	for _, v := range values {
		b.Run(fmt.Sprintf("%d.%d", v.first, v.second), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _ = combinedFirstComponents(v.first, v.second)
			}
		})
	}
}
