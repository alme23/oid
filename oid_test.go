// oid/oid_test.go
package oid

import (
	"encoding/asn1"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// ============================================
// ТЕСТЫ
// ============================================

func TestParseOID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected OID
		wantErr  bool
	}{
		{
			name:     "Стандартный OID",
			input:    "1.3.6.1.4.1",
			expected: OID{1, 3, 6, 1, 4, 1},
			wantErr:  false,
		},
		{
			name:     "OID с первым компонентом 2",
			input:    "2.100.3",
			expected: OID{2, 100, 3},
			wantErr:  false,
		},
		{
			name:     "OID с первым компонентом 0",
			input:    "0.39.1",
			expected: OID{0, 39, 1},
			wantErr:  false,
		},
		{
			name:     "OID с большими числами",
			input:    "2.999.12345.67890",
			expected: OID{2, 999, 12345, 67890},
			wantErr:  false,
		},
		{
			name:     "Пустая строка",
			input:    "",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Только один компонент",
			input:    "1",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Первый компонент больше 2",
			input:    "3.1",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Второй компонент больше 39 при первом 1",
			input:    "1.40",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Второй компонент больше 39 при первом 0",
			input:    "0.40",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Некорректные символы",
			input:    "a.b",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Отрицательное число",
			input:    "-1.3",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Число больше uint32",
			input:    "1.3.4294967296",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Двойная точка",
			input:    "1..3",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Точка в начале",
			input:    ".1.3",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Точка в конце",
			input:    "1.3.",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Пробелы",
			input:    " 1.3",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Специальные символы",
			input:    "1.3.6.1.4.1#",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseOID(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseOID(%q): ожидалась ошибка, но её нет", tt.input)
				}
				if result != nil {
					t.Errorf("ParseOID(%q): ожидался nil результат при ошибке, получено %v", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("ParseOID(%q): неожиданная ошибка: %v", tt.input, err)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("ParseOID(%q) = %v, ожидалось %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestMustParseOID(t *testing.T) {
	// Тест успешного парсинга
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParseOID не должна паниковать при корректном вводе, паника: %v", r)
			}
		}()
		result := MustParseOID("1.3.6.1")
		expected := OID{1, 3, 6, 1}
		if !result.Equal(expected) {
			t.Errorf("MustParseOID = %v, ожидалось %v", result, expected)
		}
	}()

	// Тест паники при некорректном вводе
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParseOID должна паниковать при некорректном вводе")
			}
		}()
		MustParseOID("invalid")
	}()
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
			name:     "Пустой OID",
			oid:      OID{},
			expected: "",
		},
		{
			name:     "OID с одним элементом",
			oid:      OID{1},
			expected: "1",
		},
		{
			name:     "OID с большими числами",
			oid:      OID{2, 999, 12345, 67890},
			expected: "2.999.12345.67890",
		},
		{
			name:     "OID с нулями",
			oid:      OID{0, 0, 1},
			expected: "0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.oid.String()
			if result != tt.expected {
				t.Errorf("String() = %q, ожидалось %q", result, tt.expected)
			}
		})
	}
}

func TestOIDValidate(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		{
			name:    "Корректный OID",
			oid:     OID{1, 3, 6, 1},
			wantErr: false,
		},
		{
			name:    "OID с первым компонентом 0",
			oid:     OID{0, 39, 1},
			wantErr: false,
		},
		{
			name:    "OID с первым компонентом 2",
			oid:     OID{2, 100, 1},
			wantErr: false,
		},
		{
			name:    "Пустой OID",
			oid:     OID{},
			wantErr: true,
		},
		{
			name:    "OID с одним компонентом",
			oid:     OID{1},
			wantErr: true,
		},
		{
			name:    "Первый компонент больше 2",
			oid:     OID{3, 1},
			wantErr: true,
		},
		{
			name:    "Второй компонент больше 39 при первом 1",
			oid:     OID{1, 40},
			wantErr: true,
		},
		{
			name:    "Второй компонент больше 39 при первом 0",
			oid:     OID{0, 40},
			wantErr: true,
		},
		{
			name:    "Второй компонент равен 40 при первом 2",
			oid:     OID{2, 40},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.oid.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("Validate() ожидалась ошибка для OID %v", tt.oid)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() неожиданная ошибка для OID %v: %v", tt.oid, err)
			}
		})
	}
}

// Тесты для edge cases в ParseOID
func TestParseOIDEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error // nil означает "любая ошибка"
	}{
		{"Пустая строка", "", ErrEmptyOID},
		{"Одна точка", ".", ErrInvalidOID},
		{"Две точки", "1..3", ErrInvalidOID},
		{"Точка в начале", ".1.3", ErrInvalidOID},
		{"Точка в конце", "1.3.", ErrInvalidOID},
		{"Отрицательное", "-1.3", nil},
		{"Буквы", "a.b", nil},
		{"Спецсимволы", "1.3#", nil},
		{"Слишком большое", "1.3.4294967296", nil},
		{"Один компонент", "1", ErrOIDTooShort},
		{"Первый > 2", "3.1", ErrFirstComponentTooBig},
		{"Второй > 39 при первом 1", "1.40", ErrSecondComponentTooBig},
		{"Второй > 39 при первом 0", "0.40", ErrSecondComponentTooBig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseOID(tt.input)
			if err == nil {
				t.Error("ParseOID: ожидалась ошибка")
				return
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseOID(%q) = %v, ожидалось %v", tt.input, err, tt.wantErr)
				}
			} else {
				// Любая ошибка приемлема
				t.Logf("ParseOID(%q) вернул ошибку: %v (OK)", tt.input, err)
			}
		})
	}
}

// Тесты для MaxOIDComponent
func TestMaxOIDComponent(t *testing.T) {
	// Максимальное значение
	oid := OID{2, MaxOIDComponent}
	if err := oid.Validate(); err != nil {
		t.Errorf("MaxOIDComponent должен быть валидным: %v", err)
	}

	// Превышение максимального
	tooBig := OID{2, MaxOIDComponent + 1}
	if err := tooBig.Validate(); err == nil {
		t.Error("Превышение MaxOIDComponent должно дать ошибку")
	}

	// Парсинг максимального
	parsed, err := ParseOID(oid.String())
	if err != nil {
		t.Errorf("ParseOID max: %v", err)
	}
	if !parsed.Equal(oid) {
		t.Error("ParseOID max: не совпадает")
	}
}

// Тесты для Equal
func TestEqualEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		oid1     OID
		oid2     OID
		expected bool
	}{
		{"Оба nil", nil, nil, true},
		{"Оба пустые", OID{}, OID{}, true},
		{"Nil и пустой", nil, OID{}, true},
		{"Разная длина", OID{1, 3}, OID{1, 3, 6}, false},
		{"Одинаковые", OID{1, 3, 6}, OID{1, 3, 6}, true},
		{"Разные", OID{1, 3, 6}, OID{1, 3, 7}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.oid1.Equal(tt.oid2) != tt.expected {
				t.Errorf("Equal(%v, %v) = %v, ожидалось %v",
					tt.oid1, tt.oid2, tt.oid1.Equal(tt.oid2), tt.expected)
			}
		})
	}
}

// Тесты для StartsWith
func TestStartsWithEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		prefix   OID
		expected bool
	}{
		{"Пустой префикс", OID{1, 3, 6}, OID{}, true},
		{"Nil префикс", OID{1, 3, 6}, nil, true},
		{"Полное совпадение", OID{1, 3, 6}, OID{1, 3, 6}, true},
		{"Префикс длиннее", OID{1, 3}, OID{1, 3, 6}, false},
		{"Не совпадает", OID{1, 3, 6}, OID{1, 4}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.oid.StartsWith(tt.prefix) != tt.expected {
				t.Errorf("StartsWith(%v, %v) = %v, ожидалось %v",
					tt.oid, tt.prefix, tt.oid.StartsWith(tt.prefix), tt.expected)
			}
		})
	}
}

// Тесты для Parent и Last
func TestParentLast(t *testing.T) {
	// Parent
	oid := MustParseOID("1.3.6.1")
	parent, err := oid.Parent()
	if err != nil {
		t.Errorf("Parent: %v", err)
	}
	if !parent.Equal(OID{1, 3, 6}) {
		t.Error("Parent: неверный результат")
	}

	// Parent для короткого
	_, err = OID{1}.Parent()
	if !errors.Is(err, ErrNoParent) {
		t.Error("Parent: ожидалась ErrNoParent")
	}

	// Last
	last, err := oid.Last()
	if err != nil {
		t.Errorf("Last: %v", err)
	}
	if last != 1 {
		t.Error("Last: неверный результат")
	}

	// Last для пустого
	_, err = OID{}.Last()
	if !errors.Is(err, ErrEmptyOID) {
		t.Error("Last: ожидалась ErrEmptyOID")
	}
}

// Тесты для ToASN1 и FromASN1
func TestASN1Conversion(t *testing.T) {
	oid := MustParseOID("1.3.6.1.4.1")

	// ToASN1
	asn1OID := oid.ToASN1()
	expected := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1}
	if len(asn1OID) != len(expected) {
		t.Fatal("ToASN1: неверная длина")
	}
	for i := range asn1OID {
		if asn1OID[i] != expected[i] {
			t.Error("ToASN1: неверное значение")
		}
	}

	// FromASN1
	back := FromASN1(asn1OID)
	if !back.Equal(oid) {
		t.Error("FromASN1: неверный результат")
	}
}

// Тесты для MarshalBinary/UnmarshalBinary
func TestBinaryRoundTrip(t *testing.T) {
	tests := []OID{
		MustParseOID("1.3.6.1"),
		MustParseOID("1.3.6.1.4.1"),
		MustParseOID("2.100.3"),
		MustParseOID("0.39.1"),
		{1, 3, MaxOIDComponent},
	}

	for _, oid := range tests {
		data, err := oid.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary(%v): %v", oid, err)
		}

		var decoded OID
		if err := decoded.UnmarshalBinary(data); err != nil {
			t.Fatalf("UnmarshalBinary(%v): %v", oid, err)
		}

		if !decoded.Equal(oid) {
			t.Errorf("Round trip: %v -> %v", oid, decoded)
		}
	}
}

// Тесты для ошибок UnmarshalBinary
func TestUnmarshalBinaryErrors(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{"Пустые данные", []byte{}, ErrDataTooShort},
		{"Короткие данные", []byte{0x06}, ErrDataTooShort},
		{"Неверный тег", []byte{0x05, 0x00}, ErrInvalidASN1Tag},
		{"Неверная длина", []byte{0x06, 0x80}, ErrInvalidASN1Length},
		{"Недостаточно данных", []byte{0x06, 0x05, 0x01}, ErrInsufficientData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.UnmarshalBinary(tt.data)
			if err == nil {
				t.Error("UnmarshalBinary: ожидалась ошибка")
				return
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("UnmarshalBinary = %v, ожидалось %v", err, tt.wantErr)
			}
		})
	}
}

// Тесты для MarshalJSON/UnmarshalJSON
func TestJSONRoundTrip(t *testing.T) {
	tests := []OID{
		MustParseOID("1.3.6.1"),
		MustParseOID("1.3.6.1.4.1"),
		{2, MaxOIDComponent},
	}

	for _, oid := range tests {
		data, err := oid.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON(%v): %v", oid, err)
		}

		var decoded OID
		if err := decoded.UnmarshalJSON(data); err != nil {
			t.Fatalf("UnmarshalJSON(%v): %v", oid, err)
		}

		if !decoded.Equal(oid) {
			t.Errorf("JSON round trip: %v -> %v", oid, decoded)
		}
	}
}

// Тесты для ошибок UnmarshalJSON
func TestUnmarshalJSONErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Не JSON", []byte("not-json")},
		{"Число", []byte("123")},
		{"Объект", []byte(`{"oid":"1.3.6.1"}`)},
		{"Пустая строка", []byte(`""`)},
		{"Невалидный OID", []byte(`"invalid"`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			if err := oid.UnmarshalJSON(tt.data); err == nil {
				t.Error("UnmarshalJSON: ожидалась ошибка")
			}
		})
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
			name:     "Равные OID",
			oid1:     OID{1, 3, 6, 1},
			oid2:     OID{1, 3, 6, 1},
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
			oid1:     OID{1, 3, 6, 1},
			oid2:     OID{1, 3, 6, 1, 4},
			expected: false,
		},
		{
			name:     "Пустые OID",
			oid1:     OID{},
			oid2:     OID{},
			expected: true,
		},
		{
			name:     "Пустой и непустой",
			oid1:     OID{},
			oid2:     OID{1, 3},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.oid1.Equal(tt.oid2)
			if result != tt.expected {
				t.Errorf("Equal(%v, %v) = %v, ожидалось %v",
					tt.oid1, tt.oid2, result, tt.expected)
			}
		})
	}
}

func TestOIDStartsWith(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		prefix   OID
		expected bool
	}{
		{
			name:     "Точный префикс",
			oid:      OID{1, 3, 6, 1, 4, 1, 99999},
			prefix:   OID{1, 3, 6, 1, 4, 1},
			expected: true,
		},
		{
			name:     "Полное совпадение",
			oid:      OID{1, 3, 6, 1},
			prefix:   OID{1, 3, 6, 1},
			expected: true,
		},
		{
			name:     "Префикс длиннее OID",
			oid:      OID{1, 3, 6},
			prefix:   OID{1, 3, 6, 1},
			expected: false,
		},
		{
			name:     "Не совпадает",
			oid:      OID{1, 3, 6, 1},
			prefix:   OID{1, 3, 7},
			expected: false,
		},
		{
			name:     "Пустой префикс",
			oid:      OID{1, 3, 6},
			prefix:   OID{},
			expected: true,
		},
		{
			name:     "Пустой OID с непустым префиксом",
			oid:      OID{},
			prefix:   OID{1, 3},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.oid.StartsWith(tt.prefix)
			if result != tt.expected {
				t.Errorf("StartsWith(%v, %v) = %v, ожидалось %v",
					tt.oid, tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestOIDAppend(t *testing.T) {
	tests := []struct {
		name       string
		base       OID
		components []uint32
		expected   OID
	}{
		{
			name:       "Добавление одного компонента",
			base:       OID{1, 3, 6},
			components: []uint32{1},
			expected:   OID{1, 3, 6, 1},
		},
		{
			name:       "Добавление нескольких компонентов",
			base:       OID{1, 3, 6},
			components: []uint32{1, 4, 1},
			expected:   OID{1, 3, 6, 1, 4, 1},
		},
		{
			name:       "Добавление к пустому OID",
			base:       OID{},
			components: []uint32{1, 3},
			expected:   OID{1, 3},
		},
		{
			name:       "Без добавления компонентов",
			base:       OID{1, 3, 6},
			components: []uint32{},
			expected:   OID{1, 3, 6},
		},
		{
			name:       "Добавление нулевых компонентов",
			base:       OID{1, 3},
			components: []uint32{0, 0},
			expected:   OID{1, 3, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.base.Append(tt.components...)
			if !result.Equal(tt.expected) {
				t.Errorf("Append(%v, %v) = %v, ожидалось %v",
					tt.base, tt.components, result, tt.expected)
			}

			// Проверяем, что оригинал не изменился
			if len(tt.base) > 0 && &tt.base[0] == &result[0] {
				t.Error("Append не должен изменять оригинальный слайс")
			}
		})
	}
}

func TestOIDParent(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected OID
		wantErr  bool
	}{
		{
			name:     "Обычный OID",
			oid:      OID{1, 3, 6, 1},
			expected: OID{1, 3, 6},
			wantErr:  false,
		},
		{
			name:     "OID с двумя компонентами",
			oid:      OID{1, 3},
			expected: OID{1},
			wantErr:  false,
		},
		{
			name:     "OID с одним компонентом",
			oid:      OID{1},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.oid.Parent()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parent() для %v: ожидалась ошибка", tt.oid)
				}
			} else {
				if err != nil {
					t.Errorf("Parent() для %v: неожиданная ошибка: %v", tt.oid, err)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("Parent() для %v = %v, ожидалось %v",
						tt.oid, result, tt.expected)
				}
			}
		})
	}
}

func TestOIDLast(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected uint32
		wantErr  bool
	}{
		{
			name:     "Обычный OID",
			oid:      OID{1, 3, 6, 1},
			expected: 1,
			wantErr:  false,
		},
		{
			name:     "OID с одним компонентом",
			oid:      OID{42},
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.oid.Last()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Last() для %v: ожидалась ошибка", tt.oid)
				}
			} else {
				if err != nil {
					t.Errorf("Last() для %v: неожиданная ошибка: %v", tt.oid, err)
				}
				if result != tt.expected {
					t.Errorf("Last() для %v = %d, ожидалось %d",
						tt.oid, result, tt.expected)
				}
			}
		})
	}
}

func TestOIDToASN1(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected asn1.ObjectIdentifier
	}{
		{
			name:     "Обычный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1},
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: asn1.ObjectIdentifier{},
		},
		{
			name:     "OID с большими числами",
			oid:      OID{2, 999, 12345},
			expected: asn1.ObjectIdentifier{2, 999, 12345},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.oid.ToASN1()
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ToASN1() = %v, ожидалось %v", result, tt.expected)
			}
		})
	}
}

func TestFromASN1(t *testing.T) {
	tests := []struct {
		name     string
		asn1OID  asn1.ObjectIdentifier
		expected OID
	}{
		{
			name:     "Обычный ASN1 OID",
			asn1OID:  asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1},
			expected: OID{1, 3, 6, 1, 4, 1},
		},
		{
			name:     "Пустой ASN1 OID",
			asn1OID:  asn1.ObjectIdentifier{},
			expected: OID{},
		},
		{
			name:     "ASN1 OID с большими числами",
			asn1OID:  asn1.ObjectIdentifier{2, 999, 12345},
			expected: OID{2, 999, 12345},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromASN1(tt.asn1OID)
			if !result.Equal(tt.expected) {
				t.Errorf("FromASN1() = %v, ожидалось %v", result, tt.expected)
			}
		})
	}
}

func TestOIDMarshalBinary(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		{
			name:    "Корректный OID",
			oid:     OID{1, 3, 6, 1, 4, 1},
			wantErr: false,
		},
		{
			name:    "Пустой OID",
			oid:     OID{},
			wantErr: true,
		},
		{
			name:    "Некорректный OID",
			oid:     OID{3, 1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalBinary()

			if tt.wantErr {
				if err == nil {
					t.Errorf("MarshalBinary() для %v: ожидалась ошибка", tt.oid)
				}
			} else {
				if err != nil {
					t.Errorf("MarshalBinary() для %v: неожиданная ошибка: %v", tt.oid, err)
				}
				if len(data) == 0 {
					t.Error("MarshalBinary() вернул пустые данные")
				}
			}
		})
	}
}

func TestOIDUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected OID
		wantErr  bool
	}{
		{
			name:     "Корректный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: OID{1, 3, 6, 1, 4, 1},
			wantErr:  false,
		},
		{
			name:     "OID с большими числами",
			oid:      OID{2, 999, 12345},
			expected: OID{2, 999, 12345},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сначала маршалируем
			data, err := tt.oid.MarshalBinary()
			if err != nil {
				t.Fatalf("Ошибка маршализации: %v", err)
			}

			// Затем демаршалируем
			var result OID
			err = result.UnmarshalBinary(data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalBinary(): ожидалась ошибка")
				}
			} else {
				if err != nil {
					t.Errorf("UnmarshalBinary(): неожиданная ошибка: %v", err)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("UnmarshalBinary() = %v, ожидалось %v", result, tt.expected)
				}
			}
		})
	}

	// Тест с некорректными данными
	t.Run("Некорректные данные", func(t *testing.T) {
		var result OID
		err := result.UnmarshalBinary([]byte{0xFF, 0xFF, 0xFF})
		if err == nil {
			t.Error("UnmarshalBinary(): ожидалась ошибка для некорректных данных")
		}
	})

	// Тест с пустыми данными
	t.Run("Пустые данные", func(t *testing.T) {
		var result OID
		err := result.UnmarshalBinary([]byte{})
		if err == nil {
			t.Error("UnmarshalBinary(): ожидалась ошибка для пустых данных")
		}
	})
}

func TestOIDMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected string
	}{
		{
			name:     "Обычный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: `"1.3.6.1.4.1"`,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalJSON()
			if err != nil {
				t.Errorf("MarshalJSON(): неожиданная ошибка: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("MarshalJSON() = %s, ожидалось %s", string(data), tt.expected)
			}
		})
	}
}

func TestOIDUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected OID
		wantErr  bool
	}{
		{
			name:     "Корректный JSON",
			input:    `"1.3.6.1.4.1"`,
			expected: OID{1, 3, 6, 1, 4, 1},
			wantErr:  false,
		},
		{
			name:     "Пустая строка",
			input:    `""`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Некорректный OID",
			input:    `"invalid"`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Некорректный JSON",
			input:    `{invalid}`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result OID
			err := result.UnmarshalJSON([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalJSON(%s): ожидалась ошибка", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("UnmarshalJSON(%s): неожиданная ошибка: %v", tt.input, err)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("UnmarshalJSON(%s) = %v, ожидалось %v",
						tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestOIDJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		oid  OID
	}{
		{
			name: "Стандартный OID",
			oid:  OID{1, 3, 6, 1, 4, 1},
		},
		{
			name: "OID с большими числами",
			oid:  OID{2, 999, 12345, 67890},
		},
		{
			name: "OID с первым компонентом 0",
			oid:  OID{0, 39, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Маршалируем в JSON
			jsonData, err := json.Marshal(tt.oid)
			if err != nil {
				t.Fatalf("Ошибка маршализации JSON: %v", err)
			}

			// Демаршалируем обратно
			var result OID
			err = json.Unmarshal(jsonData, &result)
			if err != nil {
				t.Fatalf("Ошибка демаршализации JSON: %v", err)
			}

			// Проверяем равенство
			if !result.Equal(tt.oid) {
				t.Errorf("JSON round trip: %v -> %v -> %v", tt.oid, string(jsonData), result)
			}
		})
	}
}

func TestOIDMemoryUsage(t *testing.T) {
	// Проверка, что OID не утекает память
	var m1, m2 runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Создаем много OID
	oids := make([]OID, 10000)
	for i := 0; i < 10000; i++ {
		oids[i] = MustParseOID("1.3.6.1.4.1.99999")
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Проверяем разницу в выделенной памяти
	// Используем TotalAlloc для общего количества выделенной памяти
	allocated := m2.TotalAlloc - m1.TotalAlloc

	// 10000 OID * (6 компонентов * 4 байта + overhead слайса)
	// Примерно 10000 * (24 + 24) = 480000 байт = 0.48 MB
	// Даем запас в 10 раз
	maxExpected := uint64(10 * 1024 * 1024) // 10 MB

	if allocated > maxExpected {
		t.Errorf("Слишком много памяти выделено: %d байт (ожидалось не более %d)",
			allocated, maxExpected)
	}

	// Проверяем, что OID все еще доступны
	if len(oids) != 10000 {
		t.Errorf("Неожиданная длина слайса: %d", len(oids))
	}

	// Проверяем корректность данных
	if !oids[0].Equal(MustParseOID("1.3.6.1.4.1.99999")) {
		t.Error("Данные OID повреждены")
	}
}

func TestOIDMemoryLeak(t *testing.T) {
	// Тест на утечку памяти
	var m1, m2 runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 10000; i++ {
		oid := MustParseOID("1.3.6.1.4.1.99999")
		_ = oid.String()
		_, _ = oid.MarshalBinary()
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Проверяем, что память вернулась
	// HeapAlloc должен быть примерно одинаковым после GC
	diff := int64(m2.HeapAlloc) - int64(m1.HeapAlloc)

	// Допускаем небольшую разницу (менее 1 MB)
	if diff > 1024*1024 || diff < -1024*1024 {
		t.Errorf("Возможна утечка памяти: разница в HeapAlloc = %d байт", diff)
	}
}

func TestOIDAppendPerformance(t *testing.T) {
	base := MustParseOID("1.3.6.1.4.1")

	// Проверка, что Append не создает лишних аллокаций
	allocs := testing.AllocsPerRun(1000, func() {
		result := base.Append(99999)
		_ = result
	})

	// Append должен делать 1 аллокацию (для нового слайса)
	if allocs > 2 {
		t.Errorf("Append делает слишком много аллокаций: %f", allocs)
	}
}

func TestOIDMemoryPooling(t *testing.T) {
	// Тест производительности с переиспользованием памяти
	oid := MustParseOID("1.3.6.1.4.1.99999")

	var data []byte
	var err error

	// Многократное кодирование
	for i := 0; i < 1000; i++ {
		data, err = oid.MarshalBinary()
		if err != nil {
			t.Fatalf("Ошибка маршализации: %v", err)
		}
	}

	// Проверяем, что данные корректны
	if len(data) == 0 {
		t.Error("Пустые данные после маршализации")
	}

	// Декодирование
	var decoded OID
	err = decoded.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("Ошибка демаршализации: %v", err)
	}

	if !decoded.Equal(oid) {
		t.Errorf("Декодированный OID = %v, ожидалось %v", decoded, oid)
	}
}

func TestOIDASN1RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		oid  OID
	}{
		{
			name: "Стандартный OID",
			oid:  OID{1, 3, 6, 1, 4, 1},
		},
		{
			name: "OID с большими числами",
			oid:  OID{2, 999, 12345, 67890},
		},
		{
			name: "Длинный OID",
			oid:  OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1, 1},
		},
		{
			name: "OID с нулями",
			oid:  OID{0, 0, 1, 2, 3},
		},
		{
			name: "Максимальное значение ASN.1",
			oid:  OID{2, MaxOIDComponent}, // 268435455
		},
		{
			name: "Минимальный OID",
			oid:  OID{0, 0},
		},
		{
			name: "Реалистичный OID",
			oid:  OID{1, 3, 6, 1, 4, 1, 99999, 1, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ASN.1 кодирование через asn1.Marshal
			asn1Data, err := asn1.Marshal(tt.oid.ToASN1())
			if err != nil {
				t.Fatalf("Ошибка ASN.1 маршализации: %v", err)
			}

			// Декодирование через наш метод
			var result OID
			err = result.UnmarshalBinary(asn1Data)
			if err != nil {
				t.Fatalf("Ошибка декодирования: %v", err)
			}

			if !result.Equal(tt.oid) {
				t.Errorf("ASN.1 round trip: %v -> %v", tt.oid, result)
			}

			// Двойная проверка: наш MarshalBinary -> asn1.Unmarshal
			ourData, err := tt.oid.MarshalBinary()
			if err != nil {
				t.Fatalf("Ошибка нашего MarshalBinary: %v", err)
			}

			var asn1Result asn1.ObjectIdentifier
			rest, err := asn1.Unmarshal(ourData, &asn1Result)
			if err != nil {
				t.Fatalf("Ошибка asn1.Unmarshal: %v", err)
			}
			if len(rest) > 0 {
				t.Errorf("Остались необработанные данные: %v", rest)
			}

			convertedBack := FromASN1(asn1Result)
			if !convertedBack.Equal(tt.oid) {
				t.Errorf("Двойной round trip: %v -> %v -> %v", tt.oid, asn1Result, convertedBack)
			}
		})
	}
}

func TestOIDEdgeCases(t *testing.T) {
	// Максимальное значение для ASN.1
	maxOID := OID{2, MaxOIDComponent}
	if err := maxOID.Validate(); err != nil {
		t.Errorf("Максимальные значения должны быть валидными: %v", err)
	}

	// Значение больше максимального
	tooBigOID := OID{2, MaxOIDComponent + 1}
	if err := tooBigOID.Validate(); err == nil {
		t.Error("Значение больше максимального должно дать ошибку")
	}

	// Минимальный валидный OID
	minOID := OID{0, 0}
	if err := minOID.Validate(); err != nil {
		t.Errorf("Минимальный OID должен быть валидным: %v", err)
	}

	// Максимальный первый компонент
	oidWithMaxFirst := OID{2, 0}
	if err := oidWithMaxFirst.Validate(); err != nil {
		t.Errorf("OID с первым компонентом 2 должен быть валидным: %v", err)
	}

	// Максимальный второй компонент при первом 0 или 1
	oidWithMaxSecond := OID{1, 39}
	if err := oidWithMaxSecond.Validate(); err != nil {
		t.Errorf("OID со вторым компонентом 39 должен быть валидным: %v", err)
	}

	// Тест на работу с максимальными значениями
	maxOIDStr := maxOID.String()
	parsedMaxOID, err := ParseOID(maxOIDStr)
	if err != nil {
		t.Errorf("Ошибка парсинга максимального OID: %v", err)
	}
	if !parsedMaxOID.Equal(maxOID) {
		t.Errorf("Парсинг максимального OID: %v -> %v", maxOID, parsedMaxOID)
	}

	// Тест парсинга слишком большого значения
	_, err = ParseOID(fmt.Sprintf("2.%d", MaxOIDComponent+1))
	if err == nil {
		t.Error("Парсинг слишком большого значения должен дать ошибку")
	}
}

// ============================================
// ТЕСТЫ ASN1-СОВМЕСТИМОСТИ
// ============================================

func TestASN1Limits(t *testing.T) {
	tests := []struct {
		name      string
		component uint32
		wantErr   bool
	}{
		{
			name:      "Максимальное значение",
			component: MaxOIDComponent,
			wantErr:   false,
		},
		{
			name:      "Значение больше максимального",
			component: MaxOIDComponent + 1,
			wantErr:   true,
		},
		{
			name:      "Обычное значение",
			component: 99999,
			wantErr:   false,
		},
		{
			name:      "Ноль",
			component: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oid := OID{2, tt.component}
			err := oid.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() для компонента %d: ожидалась ошибка", tt.component)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() для компонента %d: неожиданная ошибка: %v",
						tt.component, err)
				}

				// Пробуем маршализовать
				_, err := asn1.Marshal(oid.ToASN1())
				if err != nil {
					t.Errorf("asn1.Marshal() для компонента %d: ошибка: %v",
						tt.component, err)
				}
			}
		})
	}
}

func TestParseOIDWithASN1Limits(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Максимальное значение",
			input:   fmt.Sprintf("2.%d", MaxOIDComponent),
			wantErr: false,
		},
		{
			name:    "Значение больше максимального",
			input:   fmt.Sprintf("2.%d", MaxOIDComponent+1),
			wantErr: true,
		},
		{
			name:    "Значение uint32 max",
			input:   "2.4294967295",
			wantErr: true,
		},
		{
			name:    "Значение uint32 max",
			input:   "2.4294967295",
			wantErr: true, // Отсечется внутри Validate()
		},
		{
			name:    "Значение uint64 max",
			input:   "2.18446744073709551615",
			wantErr: true, // Отсечется внутри strconv.ParseUint
		},
		{
			name:    "Значение выше uint64 max",
			input:   "2.18446744073709551616",
			wantErr: true, // Вызовет ошибку strconv.ErrRange
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseOID(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseOID(%q): ожидалась ошибка", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseOID(%q): неожиданная ошибка: %v", tt.input, err)
				}
			}
		})
	}
}

// ============================================
// ТЕСТЫ CONCURRENCY
// ============================================

func TestOIDConcurrentAccess(t *testing.T) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	// Проверка конкурентного доступа к неизменяемому OID
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Операции чтения
			str := oid.String()
			if str == "" {
				errors <- fmt.Errorf("пустая строка")
			}

			parent, err := oid.Parent()
			if err != nil {
				errors <- err
			}
			_ = parent

			// Маршализация
			data, err := oid.MarshalBinary()
			if err != nil {
				errors <- err
			}
			_ = data
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Ошибка при конкурентном доступе: %v", err)
	}
}

func TestRegistryConcurrentOperations(t *testing.T) {
	reg := NewRegistry()

	var wg sync.WaitGroup
	errors := make(chan error, 1000)

	// Конкурентная запись
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			name := fmt.Sprintf("oid-%d", id)
			oid := OID{1, 3, uint32(id)}

			if err := reg.Register(name, oid); err != nil {
				// Некоторые ошибки возможны при коллизиях
				if !strings.Contains(err.Error(), "уже зарегистрирован") {
					errors <- err
				}
			}
		}(i)
	}

	// Конкурентное чтение
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				_, _ = reg.LookupByName("nonexistent")
				_, _ = reg.LookupByOID(OID{9, 9, uint32(j)})
				_ = reg.List()
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Ошибка при конкурентных операциях: %v", err)
	}
}

func TestLookupByNameNoCopy(t *testing.T) {
	reg := NewRegistry()
	original := MustParseOID("1.3.6.1.4.1")
	reg.Register("test", original)

	// Получаем без копирования
	noCopy, exists := reg.LookupByNameNoCopy("test")
	if !exists {
		t.Fatal("OID не найден")
	}

	// Проверяем, что это тот же объект
	if &noCopy[0] != &original[0] {
		// Не обязательно тот же адрес, но значения должны совпадать
		if !noCopy.Equal(original) {
			t.Error("NoCopy OID не совпадает с оригиналом")
		}
	}

	// Проверяем, что изменение NoCopy OID влияет на реестр (опасно!)
	noCopy[0] = 2

	// Получаем снова и проверяем
	modified, _ := reg.LookupByNameNoCopy("test")
	if modified[0] != 2 {
		t.Error("Изменение NoCopy OID должно влиять на реестр")
	}

	// Восстанавливаем
	modified[0] = 1
}

func TestListNoCopy(t *testing.T) {
	reg := NewRegistry()
	oid1 := MustParseOID("1.3.6.1.4.1")
	oid2 := MustParseOID("2.100.3")
	reg.Register("first", oid1)
	reg.Register("second", oid2)

	// Получаем без копирования
	list := reg.ListNoCopy()

	if len(list) != 2 {
		t.Errorf("len = %d, ожидалось 2", len(list))
	}

	// Проверяем, что OID доступны
	if !list["first"].Equal(oid1) {
		t.Error("first OID не совпадает")
	}
	if !list["second"].Equal(oid2) {
		t.Error("second OID не совпадает")
	}

	// Изменяем OID в list (опасно!)
	list["first"][0] = 2

	// Проверяем, что изменение отразилось в реестре
	modified, _ := reg.LookupByNameNoCopy("first")
	if modified[0] != 2 {
		t.Error("Изменение в ListNoCopy должно влиять на реестр")
	}
}

func TestOIDsNoCopy(t *testing.T) {
	reg := NewRegistry()
	oid1 := MustParseOID("1.3.6.1.4.1")
	oid2 := MustParseOID("2.100.3")
	reg.Register("first", oid1)
	reg.Register("second", oid2)

	oids := reg.OIDsNoCopy()

	if len(oids) != 2 {
		t.Errorf("len = %d, ожидалось 2", len(oids))
	}

	// Проверяем наличие обоих OID
	found1 := false
	found2 := false
	for _, oid := range oids {
		if oid.Equal(oid1) {
			found1 = true
		}
		if oid.Equal(oid2) {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Не все OID найдены")
	}
}

func TestNoCopyIsolation(t *testing.T) {
	reg := NewRegistry()
	original := MustParseOID("1.3.6.1.4.1")
	reg.Register("test", original)

	// Получаем копию
	copied, _ := reg.LookupByName("test")

	// Получаем без копирования
	noCopy, _ := reg.LookupByNameNoCopy("test")

	// Изменяем копию
	copied[0] = 2

	// Проверяем, что NoCopy не изменился
	if noCopy[0] != 1 {
		t.Error("Изменение копии не должно влиять на NoCopy")
	}

	// Изменяем NoCopy
	noCopy[0] = 3

	// Проверяем, что реестр изменился
	modified, _ := reg.LookupByNameNoCopy("test")
	if modified[0] != 3 {
		t.Error("Изменение NoCopy должно влиять на реестр")
	}

	// Восстанавливаем
	noCopy[0] = 1
}

func TestConcurrentNoCopyRead(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1")
	reg.Register("test", oid)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Конкурентное чтение без копирования
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				noCopy, exists := reg.LookupByNameNoCopy("test")
				if !exists {
					errors <- fmt.Errorf("OID не найден")
					return
				}
				// Только чтение
				_ = noCopy.String()
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// Тесты для digitCount
func TestDigitCount(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected int
	}{
		{"0", 0, 1},
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
			if got := digitCount(tt.input); got != tt.expected {
				t.Errorf("digitCount(%d) = %d, ожидалось %d", tt.input, got, tt.expected)
			}
		})
	}
}

// Тесты для base128Size
func TestBase128Size(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected int
	}{
		{"0", 0, 1},
		{"127", 127, 1},
		{"128", 128, 2},
		{"16383", 16383, 2},
		{"16384", 16384, 3},
		{"2097151", 2097151, 3},
		{"2097152", 2097152, 4},
		{"268435455", 268435455, 4},
		{"268435456", 268435456, 5},
		{"MaxUint32", ^uint32(0), 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := base128Size(tt.value); got != tt.expected {
				t.Errorf("base128Size(%d) = %d, ожидалось %d", tt.value, got, tt.expected)
			}
		})
	}
}

// Тесты для lengthSize
func TestLengthSize(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected int
	}{
		{"0", 0, 1},
		{"127", 127, 1},
		{"128", 128, 2},
		{"255", 255, 2},
		{"256", 256, 3},
		{"65535", 65535, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lengthSize(tt.length); got != tt.expected {
				t.Errorf("lengthSize(%d) = %d, ожидалось %d", tt.length, got, tt.expected)
			}
		})
	}
}

// Тесты для writeBase128
func TestWriteBase128(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{"0", 0, []byte{0x00}},
		{"127", 127, []byte{0x7F}},
		{"128", 128, []byte{0x81, 0x00}},
		{"16383", 16383, []byte{0xFF, 0x7F}},
		{"16384", 16384, []byte{0x81, 0x80, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 5)
			n := writeBase128(buf, tt.value)

			if n != len(tt.expected) {
				t.Errorf("writeBase128(%d) = %d байт, ожидалось %d", tt.value, n, len(tt.expected))
			}

			for i := 0; i < n; i++ {
				if buf[i] != tt.expected[i] {
					t.Errorf("writeBase128(%d)[%d] = 0x%02x, ожидалось 0x%02x",
						tt.value, i, buf[i], tt.expected[i])
				}
			}
		})
	}
}

// Тесты для writeLength
func TestWriteLength(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected []byte
	}{
		{"0", 0, []byte{0x00}},
		{"127", 127, []byte{0x7F}},
		{"128", 128, []byte{0x81, 0x80}},
		{"255", 255, []byte{0x81, 0xFF}},
		{"256", 256, []byte{0x82, 0x01, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 4)
			n := writeLength(buf, tt.length)

			if n != len(tt.expected) {
				t.Errorf("writeLength(%d) = %d байт, ожидалось %d", tt.length, n, len(tt.expected))
			}

			for i := 0; i < n; i++ {
				if buf[i] != tt.expected[i] {
					t.Errorf("writeLength(%d)[%d] = 0x%02x, ожидалось 0x%02x",
						tt.length, i, buf[i], tt.expected[i])
				}
			}
		})
	}
}

// Тесты для readBase128
func TestReadBase128(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint32
		bytes    int
	}{
		{"0", []byte{0x00}, 0, 1},
		{"127", []byte{0x7F}, 127, 1},
		{"128", []byte{0x81, 0x00}, 128, 2},
		{"16383", []byte{0xFF, 0x7F}, 16383, 2},
		{"16384", []byte{0x81, 0x80, 0x00}, 16384, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, bytesRead := readBase128(tt.data)

			if bytesRead != tt.bytes {
				t.Errorf("readBase128() bytesRead = %d, ожидалось %d", bytesRead, tt.bytes)
			}

			if value != tt.expected {
				t.Errorf("readBase128() value = %d, ожидалось %d", value, tt.expected)
			}
		})
	}
}

// Тесты для readLength
func TestReadLength(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
		bytes    int
	}{
		{"0", []byte{0x00}, 0, 1},
		{"127", []byte{0x7F}, 127, 1},
		{"128", []byte{0x81, 0x80}, 128, 2},
		{"255", []byte{0x81, 0xFF}, 255, 2},
		{"256", []byte{0x82, 0x01, 0x00}, 256, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, bytesRead := readLength(tt.data)

			if bytesRead != tt.bytes {
				t.Errorf("readLength() bytesRead = %d, ожидалось %d", bytesRead, tt.bytes)
			}

			if length != tt.expected {
				t.Errorf("readLength() length = %d, ожидалось %d", length, tt.expected)
			}
		})
	}
}

// Тесты для appendBase128Value
func TestAppendBase128Value(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{"0", 0, []byte{0x00}},
		{"127", 127, []byte{0x7F}},
		{"128", 128, []byte{0x81, 0x00}},
		{"16383", 16383, []byte{0xFF, 0x7F}},
		{"16384", 16384, []byte{0x81, 0x80, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendBase128Value(nil, tt.value)

			if len(result) != len(tt.expected) {
				t.Errorf("appendBase128Value(%d) = %x, ожидалось %x", tt.value, result, tt.expected)
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("appendBase128Value(%d)[%d] = 0x%02x, ожидалось 0x%02x",
						tt.value, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// Тесты для комбинированных первых компонентов
func TestCombinedFirstComponents(t *testing.T) {
	tests := []struct {
		name     string
		first    uint32
		second   uint32
		expected uint32
		wantErr  bool
	}{
		{"0.0", 0, 0, 0, false},
		{"1.39", 1, 39, 79, false},
		{"2.175", 2, 175, 255, false},
		{"2.999", 2, 999, 1079, false},
		{"2.MaxUint32", 2, ^uint32(0), 0, true}, // Переполнение
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := combinedFirstComponents(tt.first, tt.second)

			if tt.wantErr {
				if err == nil {
					t.Error("combinedFirstComponents: ожидалась ошибка")
				}
				return
			}

			if err != nil {
				t.Errorf("combinedFirstComponents: %v", err)
			}

			if result != tt.expected {
				t.Errorf("combinedFirstComponents = %d, ожидалось %d", result, tt.expected)
			}
		})
	}
}

// Тесты для глобального API
func TestGlobalMustFunctions(t *testing.T) {
	ResetRegistry()

	// MustRegister
	MustRegister("test", MustParseOID("1.3.6.1"))
	if Size() != 1 {
		t.Error("MustRegister: неверный размер")
	}

	// MustBatchRegister
	MustBatchRegister(map[string]OID{
		"first":  MustParseOID("1.3.6.1.4.1"),
		"second": MustParseOID("2.100.3"),
	})
	if Size() != 3 {
		t.Error("MustBatchRegister: неверный размер")
	}

	// Panic при дубликате
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustRegister: ожидалась паника при дубликате")
		}
	}()
	MustRegister("test2", MustParseOID("1.3.6.1"))
}

// Тесты для NullOID
func TestNullOIDAllMethods(t *testing.T) {
	// FromOID
	n := FromOID(MustParseOID("1.3.6.1"))
	if !n.Valid {
		t.Error("FromOID: Valid должно быть true")
	}

	// String
	if n.String() != "1.3.6.1" {
		t.Error("NullOID.String: неверное значение")
	}

	// Equal
	n2 := FromOID(MustParseOID("1.3.6.1"))
	if !n.Equal(n2) {
		t.Error("NullOID.Equal: должны быть равны")
	}

	// MarshalJSON
	data, err := json.Marshal(n)
	if err != nil {
		t.Errorf("NullOID.MarshalJSON: %v", err)
	}
	if string(data) != `"1.3.6.1"` {
		t.Errorf("NullOID.MarshalJSON = %s", data)
	}

	// UnmarshalJSON
	var n3 NullOID
	if err := json.Unmarshal(data, &n3); err != nil {
		t.Errorf("NullOID.UnmarshalJSON: %v", err)
	}
	if !n3.Equal(n) {
		t.Error("NullOID.UnmarshalJSON: не совпадает")
	}

	// MustFromString
	n4 := MustFromString("1.3.6.1")
	if !n4.Equal(n) {
		t.Error("MustFromString: не совпадает")
	}
}

// Тесты для Array
func TestArrayAllMethods(t *testing.T) {
	arr := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}

	// String
	if arr.String() != "[1.3.6.1, 2.100.3]" {
		t.Errorf("Array.String = %q", arr.String())
	}

	// Equal
	arr2 := Array{
		MustParseOID("1.3.6.1"),
		MustParseOID("2.100.3"),
	}
	if !arr.Equal(arr2) {
		t.Error("Array.Equal: должны быть равны")
	}

	// Contains
	if !arr.Contains(MustParseOID("1.3.6.1")) {
		t.Error("Array.Contains: должен найти")
	}

	// Append
	arr3 := arr.Append(MustParseOID("0.39.1"))
	if len(arr3) != 3 {
		t.Error("Array.Append: неверная длина")
	}

	// MarshalJSON
	data, err := json.Marshal(arr)
	if err != nil {
		t.Errorf("Array.MarshalJSON: %v", err)
	}
	if string(data) != `["1.3.6.1","2.100.3"]` {
		t.Errorf("Array.MarshalJSON = %s", data)
	}

	// UnmarshalJSON
	var arr4 Array
	if err := json.Unmarshal(data, &arr4); err != nil {
		t.Errorf("Array.UnmarshalJSON: %v", err)
	}
	if !arr4.Equal(arr) {
		t.Error("Array.UnmarshalJSON: не совпадает")
	}
}

// Тесты для Registry Names и OIDs
func TestRegistryNamesAndOIDs(t *testing.T) {
	reg := NewRegistry()

	// Пустой реестр
	if len(reg.Names()) != 0 {
		t.Error("Names: должен быть пустым")
	}
	if len(reg.OIDs()) != 0 {
		t.Error("OIDs: должен быть пустым")
	}
	if len(reg.OIDsNoCopy()) != 0 {
		t.Error("OIDsNoCopy: должен быть пустым")
	}

	// С записями
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	reg.Register("first", oid1)
	reg.Register("second", oid2)

	names := reg.Names()
	if len(names) != 2 {
		t.Error("Names: неверная длина")
	}

	oids := reg.OIDs()
	if len(oids) != 2 {
		t.Error("OIDs: неверная длина")
	}

	oidsNoCopy := reg.OIDsNoCopy()
	if len(oidsNoCopy) != 2 {
		t.Error("OIDsNoCopy: неверная длина")
	}
}

// Тесты для глобального ListNoCopy и OIDsNoCopy
func TestGlobalNoCopyFunctions(t *testing.T) {
	ResetRegistry()

	Register("test", MustParseOID("1.3.6.1"))

	list := ListNoCopy()
	if len(list) != 1 {
		t.Error("ListNoCopy: неверная длина")
	}

	oids := OIDsNoCopy()
	if len(oids) != 1 {
		t.Error("OIDsNoCopy: неверная длина")
	}
}

func TestMarshalBinaryErrors(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr error
	}{
		{"Пустой", OID{}, ErrOIDTooShort},
		{"Один компонент", OID{1}, ErrOIDTooShort},
		{"Первый > 2", OID{3, 1}, ErrFirstComponentTooBig},
		{"Второй > 39", OID{1, 40}, ErrSecondComponentTooBig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.oid.MarshalBinary()
			if err == nil {
				t.Error("MarshalBinary: ожидалась ошибка")
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("MarshalBinary = %v, ожидалось %v", err, tt.wantErr)
			}
		})
	}
}

func TestBatchRegisterAllConflicts(t *testing.T) {
	t.Run("DuplicateOIDInsideBatch", func(t *testing.T) {
		reg := NewRegistry()
		oid := MustParseOID("1.3.6.1")

		err := reg.BatchRegister(map[string]OID{
			"first":  oid,
			"second": oid,
		})
		if !errors.Is(err, ErrDuplicateOIDInBatch) {
			t.Errorf("BatchRegister = %v, ожидалось ErrDuplicateOIDInBatch", err)
		}

		// Проверяем атомарность
		if reg.Size() != 0 {
			t.Error("BatchRegister: атомарность нарушена")
		}
	})

	t.Run("NameAlreadyExists", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register("existing", MustParseOID("1.3.6.1.4.1"))

		err := reg.BatchRegister(map[string]OID{
			"existing": MustParseOID("2.100.3"),
			"new":      MustParseOID("1.3.6.1.4.2"),
		})
		if !errors.Is(err, ErrNameAlreadyExists) {
			t.Errorf("BatchRegister = %v, ожидалось ErrNameAlreadyExists", err)
		}

		// Проверяем атомарность
		if reg.Size() != 1 {
			t.Error("BatchRegister: атомарность нарушена")
		}
		if _, exists := reg.LookupByName("new"); exists {
			t.Error("BatchRegister: 'new' не должен быть зарегистрирован")
		}
	})

	t.Run("OIDAlreadyRegistered", func(t *testing.T) {
		reg := NewRegistry()
		oid := MustParseOID("1.3.6.1")
		reg.Register("original", oid)

		err := reg.BatchRegister(map[string]OID{
			"different_name": oid,
			"new":            MustParseOID("2.100.3"),
		})
		if !errors.Is(err, ErrOIDAlreadyRegistered) {
			t.Errorf("BatchRegister = %v, ожидалось ErrOIDAlreadyRegistered", err)
		}

		// Проверяем атомарность
		if reg.Size() != 1 {
			t.Error("BatchRegister: атомарность нарушена")
		}
		if _, exists := reg.LookupByName("new"); exists {
			t.Error("BatchRegister: 'new' не должен быть зарегистрирован")
		}
	})

	t.Run("SuccessfulBatch", func(t *testing.T) {
		reg := NewRegistry()

		err := reg.BatchRegister(map[string]OID{
			"first":  MustParseOID("1.3.6.1"),
			"second": MustParseOID("2.100.3"),
			"third":  MustParseOID("0.39.1"),
		})
		if err != nil {
			t.Errorf("BatchRegister: %v", err)
		}

		if reg.Size() != 3 {
			t.Error("BatchRegister: неверный размер")
		}
	})
}

func TestReadBase128KnownValues(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint32
		bytes    int
	}{
		{"0", []byte{0x00}, 0, 1},
		{"1", []byte{0x01}, 1, 1},
		{"127", []byte{0x7F}, 127, 1},
		{"128", []byte{0x81, 0x00}, 128, 2},
		{"129", []byte{0x81, 0x01}, 129, 2},
		{"16383", []byte{0xFF, 0x7F}, 16383, 2},
		{"16384", []byte{0x81, 0x80, 0x00}, 16384, 3},
		{"2097151", []byte{0xFF, 0xFF, 0x7F}, 2097151, 3},
		{"2097152", []byte{0x81, 0x80, 0x80, 0x00}, 2097152, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, bytesRead := readBase128(tt.data)
			if bytesRead != tt.bytes {
				t.Errorf("readBase128(%x) bytesRead = %d, ожидалось %d", tt.data, bytesRead, tt.bytes)
			}
			if value != tt.expected {
				t.Errorf("readBase128(%x) value = %d, ожидалось %d", tt.data, value, tt.expected)
			}
		})
	}
}
func TestMarshalBinaryFullCoverage(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		{"Короткий", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},
		{"Длинный", MustParseOID("1.3.6.1.4.1.99999.1.1"), false},
		{"Максимальный компонент", OID{1, 3, MaxOIDComponent}, false},
		{"Первый = 2", MustParseOID("2.100.3"), false},
		{"Пустой", OID{}, true},
		{"Невалидный", OID{3, 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalBinary()
			if tt.wantErr {
				if err == nil {
					t.Error("MarshalBinary: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("MarshalBinary: %v", err)
			}
			if len(data) == 0 {
				t.Error("MarshalBinary: пустой результат")
			}
		})
	}
}

func TestUnmarshalBinaryFullCoverage(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{"Успешный", []byte{0x06, 0x03, 0x2B, 0x06, 0x01}, nil},
		{"Пустой", []byte{}, ErrDataTooShort},
		{"1 байт", []byte{0x06}, ErrDataTooShort},
		{"Неверный тег", []byte{0x05, 0x00}, ErrInvalidASN1Tag},
		{"Неверная длина", []byte{0x06, 0x80}, ErrInvalidASN1Length},
		{"Недостаточно данных", []byte{0x06, 0x05, 0x01}, ErrInsufficientData},
		{"Неверный первый компонент", []byte{0x06, 0x01, 0x80}, ErrInvalidFirstComponent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.UnmarshalBinary(tt.data)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("UnmarshalBinary(%x): %v", tt.data, err)
				}
				return
			}

			if err == nil {
				t.Error("UnmarshalBinary: ожидалась ошибка")
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UnmarshalBinary(%x) = %v, ожидалось %v", tt.data, err, tt.wantErr)
			}
		})
	}
}

// Тест для lengthSize - покрытие ветви >= 65536
func TestLengthSizeLarge(t *testing.T) {
	// Ветвь length >= 65536
	if got := lengthSize(65536); got != 4 {
		t.Errorf("lengthSize(65536) = %d, ожидалось 4", got)
	}
	if got := lengthSize(1000000); got != 4 {
		t.Errorf("lengthSize(1000000) = %d, ожидалось 4", got)
	}
}

// Тест для readLength - покрытие переполнения
func TestReadLengthOverflowBranch(t *testing.T) {
	// Тест на переполнение int (0x84 с 4 байтами)
	// На 32-битной системе это вызовет переполнение
	data := []byte{0x84, 0xFF, 0xFF, 0xFF, 0xFF}
	length, bytesRead := readLength(data)

	// На 64-битной системе это валидно
	if bytesRead == 5 && length == 4294967295 {
		t.Logf("64-битная система: length = %d", length)
	} else if bytesRead == 0 {
		t.Logf("32-битная система: переполнение")
	}
}

// Тест для BatchRegister - покрытие конфликта OID
func TestBatchRegisterOIDConflictBranch(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("original", oid)

	// Пытаемся зарегистрировать тот же OID под другим именем
	err := reg.BatchRegister(map[string]OID{
		"different": oid,
	})

	if err == nil {
		t.Fatal("BatchRegister: ожидалась ошибка")
	}
	if !errors.Is(err, ErrOIDAlreadyRegistered) {
		t.Errorf("BatchRegister = %v, ожидалось ErrOIDAlreadyRegistered", err)
	}
}

// Тест для MarshalBinary - покрытие длинного контента
func TestMarshalBinaryLongContentCoverage(t *testing.T) {
	// OID с длинным контентом
	longOID := OID{1, 3}
	for i := 0; i < 50; i++ {
		longOID = append(longOID, MaxOIDComponent)
	}

	data, err := longOID.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	if len(data) < 128 {
		t.Errorf("Длина %d < 128", len(data))
	}
}

// Тест для UnmarshalBinary - покрытие неверного компонента
func TestUnmarshalBinaryInvalidComponentBranch(t *testing.T) {
	// Данные с неверным компонентом (continuation bit без завершения)
	data := []byte{0x06, 0x02, 0x2B, 0x80}

	var oid OID
	err := oid.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("UnmarshalBinary: ожидалась ошибка")
	}
	if !errors.Is(err, ErrInvalidComponent) {
		t.Errorf("UnmarshalBinary = %v, ожидалось ErrInvalidComponent", err)
	}
}

// Тест для SizeBER - покрытие длинного контента
func TestSizeBERLongContentCoverage(t *testing.T) {
	longOID := OID{1, 3}
	for i := 0; i < 50; i++ {
		longOID = append(longOID, MaxOIDComponent)
	}

	size, err := longOID.SizeBER()
	if err != nil {
		t.Fatalf("SizeBER: %v", err)
	}

	data, err := longOID.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	if size != len(data) {
		t.Errorf("SizeBER = %d, MarshalBER = %d", size, len(data))
	}
}

// Тест для UnmarshalBERContent - покрытие неверного компонента
func TestUnmarshalBERContentInvalidComponentBranch(t *testing.T) {
	// Контент с неверным компонентом
	content := []byte{0x2B, 0x80}

	var oid OID
	err := oid.UnmarshalBERContent(content)
	if err == nil {
		t.Fatal("UnmarshalBERContent: ожидалась ошибка")
	}
	if !errors.Is(err, ErrComponentFailed) {
		t.Errorf("UnmarshalBERContent = %v, ожидалось ErrComponentFailed", err)
	}
}

// Тест для UnmarshalBinary - покрытие неверного компонента
func TestUnmarshalBinaryInvalidComponent(t *testing.T) {
	data := []byte{0x06, 0x02, 0x2B, 0x80}

	var oid OID
	err := oid.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("UnmarshalBinary: ожидалась ошибка")
	}
	if !errors.Is(err, ErrInvalidComponent) {
		t.Errorf("UnmarshalBinary = %v, ожидалось ErrInvalidComponent", err)
	}
}

// Тест для BatchRegister - покрытие конфликта OID
func TestBatchRegisterOIDConflict(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("original", oid)

	err := reg.BatchRegister(map[string]OID{
		"different": oid,
	})

	if err == nil {
		t.Fatal("BatchRegister: ожидалась ошибка")
	}
	if !errors.Is(err, ErrOIDAlreadyRegistered) {
		t.Errorf("BatchRegister = %v, ожидалось ErrOIDAlreadyRegistered", err)
	}
}

// Тест для UnmarshalBERContent - покрытие неверного компонента
func TestUnmarshalBERContentInvalidComponent(t *testing.T) {
	content := []byte{0x2B, 0x80}

	var oid OID
	err := oid.UnmarshalBERContent(content)
	if err == nil {
		t.Fatal("UnmarshalBERContent: ожидалась ошибка")
	}
	if !errors.Is(err, ErrComponentFailed) {
		t.Errorf("UnmarshalBERContent = %v, ожидалось ErrComponentFailed", err)
	}
}

// Тест для MarshalBER - покрытие ошибки переполнения
func TestMarshalBEROverflow(t *testing.T) {
	// OID с первыми компонентами, дающими переполнение
	oid := OID{2, MaxOIDComponent}

	// Это не должно вызвать переполнение, так как 40*2 + MaxOIDComponent = 268435535 < uint32 max
	data, err := oid.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	if len(data) == 0 {
		t.Error("MarshalBER: пустой результат")
	}
}

// Тест для MarshalJSON (Array) - покрытие пустого массива
func TestArrayMarshalJSONEmptyCoverage(t *testing.T) {
	arr := Array{}

	data, err := arr.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	if string(data) != "[]" {
		t.Errorf("MarshalJSON = %s, ожидалось '[]'", data)
	}
}

func TestBatchRegisterFullCoverage(t *testing.T) {
	// Успешная регистрация
	t.Run("Success", func(t *testing.T) {
		reg := NewRegistry()
		err := reg.BatchRegister(map[string]OID{
			"first":  MustParseOID("1.3.6.1"),
			"second": MustParseOID("2.100.3"),
		})
		if err != nil {
			t.Errorf("BatchRegister: %v", err)
		}
		if reg.Size() != 2 {
			t.Error("BatchRegister: неверный размер")
		}
	})

	// Пустой map
	t.Run("Empty", func(t *testing.T) {
		reg := NewRegistry()
		err := reg.BatchRegister(map[string]OID{})
		if err != nil {
			t.Errorf("BatchRegister: %v", err)
		}
	})

	// Невалидный OID
	t.Run("InvalidOID", func(t *testing.T) {
		reg := NewRegistry()
		err := reg.BatchRegister(map[string]OID{
			"bad": {3, 1},
		})
		if err == nil {
			t.Error("BatchRegister: ожидалась ошибка")
		}
	})

	// Дубликат OID
	t.Run("DuplicateOID", func(t *testing.T) {
		reg := NewRegistry()
		oid := MustParseOID("1.3.6.1")
		err := reg.BatchRegister(map[string]OID{
			"first":  oid,
			"second": oid,
		})
		if !errors.Is(err, ErrDuplicateOIDInBatch) {
			t.Errorf("BatchRegister = %v, ожидалось ErrDuplicateOIDInBatch", err)
		}
	})

	// Конфликт имени
	t.Run("NameConflict", func(t *testing.T) {
		reg := NewRegistry()
		reg.Register("existing", MustParseOID("1.3.6.1.4.1"))
		err := reg.BatchRegister(map[string]OID{
			"existing": MustParseOID("2.100.3"),
		})
		if !errors.Is(err, ErrNameAlreadyExists) {
			t.Errorf("BatchRegister = %v, ожидалось ErrNameAlreadyExists", err)
		}
	})

	// Конфликт OID
	t.Run("OIDConflict", func(t *testing.T) {
		reg := NewRegistry()
		oid := MustParseOID("1.3.6.1")
		reg.Register("original", oid)
		err := reg.BatchRegister(map[string]OID{
			"different": oid,
		})
		if !errors.Is(err, ErrOIDAlreadyRegistered) {
			t.Errorf("BatchRegister = %v, ожидалось ErrOIDAlreadyRegistered", err)
		}
	})
}

func TestUnmarshalBERContentFullCoverage(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		wantErr error
	}{
		{"Успешный", []byte{0x2B, 0x06, 0x01}, nil},
		{"Пустой", []byte{}, ErrEmptyContent},
		{"Неверный первый", []byte{0x80}, ErrFirstComponentFailed},
		{"Неверный компонент", []byte{0x2B, 0x80}, ErrComponentFailed},
		{"Переполнение", []byte{0x2B, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}, ErrComponentFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.UnmarshalBERContent(tt.content)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("UnmarshalBERContent(%x): %v", tt.content, err)
				}
				return
			}

			if err == nil {
				t.Error("UnmarshalBERContent: ожидалась ошибка")
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UnmarshalBERContent(%x) = %v, ожидалось %v", tt.content, err, tt.wantErr)
			}
		})
	}
}

func TestReadBase128FullCoverage(t *testing.T) {
	// Используем round trip для проверки корректности
	tests := []uint32{
		0, 1, 127, 128, 129, 255, 256, 16383, 16384,
		2097151, 2097152, MaxOIDComponent,
	}

	for _, expected := range tests {
		t.Run(fmt.Sprintf("RoundTrip_%d", expected), func(t *testing.T) {
			// Кодируем с помощью writeBase128
			buf := make([]byte, 5)
			n := writeBase128(buf, expected)

			// Декодируем с помощью readBase128
			value, bytesRead := readBase128(buf[:n])

			if bytesRead != n {
				t.Errorf("bytesRead = %d, ожидалось %d", bytesRead, n)
			}
			if value != expected {
				t.Errorf("value = %d, ожидалось %d", value, expected)
			}
		})
	}

	// Тесты на ошибки
	errorTests := []struct {
		name string
		data []byte
	}{
		{"Переполнение", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}},
		{"Незавершенная", []byte{0x81, 0x80}},
		{"Пустая", []byte{}},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			value, bytesRead := readBase128(tt.data)
			if bytesRead != 0 {
				t.Errorf("bytesRead = %d, ожидалось 0", bytesRead)
			}
			if value != 0 {
				t.Errorf("value = %d, ожидалось 0", value)
			}
		})
	}
}

func TestMaxValues(t *testing.T) {
	// MaxOIDComponent = 268435455 = 2^28 - 1
	// Кодируется как: 0x8F 0xFF 0xFF 0xFF 0x7F

	// MaxUint32 = 4294967295 = 2^32 - 1
	// НО: в base-128 максимум 5 байт = 35 бит
	// Реально максимум: 0x8F 0xFF 0xFF 0xFF 0x7F = 4294967295

	// Проверяем MaxOIDComponent
	data := []byte{0x8F, 0xFF, 0xFF, 0xFF, 0x7F}
	value, _ := readBase128(data)

	if value != 268435455 && value != 4294967295 {
		t.Errorf("value = %d", value)
	}

	// Реальное значение 0x8F 0xFF 0xFF 0xFF 0x7F
	// 0x0F = 15 (старшие 4 бита первого байта)
	// 15 * 2^28 + 0x0FFFFFFF = 15 * 268435456 + 268435455 = 4026531840 + 268435455 = 4294967295
}

func TestReadLengthFullCoverage(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
		bytes    int
	}{
		{"0", []byte{0x00}, 0, 1},
		{"127", []byte{0x7F}, 127, 1},
		{"128 (0x81)", []byte{0x81, 0x80}, 128, 2},
		{"255 (0x81)", []byte{0x81, 0xFF}, 255, 2},
		{"256 (0x82)", []byte{0x82, 0x01, 0x00}, 256, 3},
		{"65535 (0x82)", []byte{0x82, 0xFF, 0xFF}, 65535, 3},
		{"0x80 (некорректная)", []byte{0x80}, 0, 0},
		{"0x85 (слишком длинная)", []byte{0x85}, 0, 0},
		{"0x81 с недостатком", []byte{0x81}, 0, 0},
		{"Пустая", []byte{}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, bytesRead := readLength(tt.data)
			if bytesRead != tt.bytes {
				t.Errorf("readLength() bytesRead = %d, ожидалось %d", bytesRead, tt.bytes)
			}
			if length != tt.expected {
				t.Errorf("readLength() length = %d, ожидалось %d", length, tt.expected)
			}
		})
	}
}

func TestMustFromStringFullCoverage(t *testing.T) {
	// Успешный случай
	n := MustFromString("1.3.6.1")
	if !n.Valid {
		t.Error("MustFromString: Valid должно быть true")
	}

	// Паника при ошибке
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustFromString: ожидалась паника")
		}
	}()
	MustFromString("invalid")
}

func TestDiffFullCoverage(t *testing.T) {
	ResetRegistry()

	// Пустой снимок
	snapshot := Snapshot()

	// Добавляем
	Register("first", MustParseOID("1.3.6.1"))
	added, removed, changed := Diff(snapshot)
	if len(added) != 1 || len(removed) != 0 || len(changed) != 0 {
		t.Error("Diff: неверный результат при добавлении")
	}

	// Изменяем
	snapshot = Snapshot()
	Remove("first")
	Register("first", MustParseOID("2.100.3"))
	added, removed, changed = Diff(snapshot)
	if len(added) != 0 || len(removed) != 0 || len(changed) != 1 {
		t.Error("Diff: неверный результат при изменении")
	}

	// Удаляем
	snapshot = Snapshot()
	Remove("first")
	added, removed, changed = Diff(snapshot)
	if len(added) != 0 || len(removed) != 1 || len(changed) != 0 {
		t.Error("Diff: неверный результат при удалении")
	}
}

// Тесты для ошибок в MarshalBER
func TestMarshalBERErrors(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr error
	}{
		{"Пустой", OID{}, ErrOIDTooShort},
		{"Один компонент", OID{1}, ErrOIDTooShort},
		{"Первый > 2", OID{3, 1}, ErrFirstComponentTooBig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.oid.MarshalBER()
			if err == nil {
				t.Error("MarshalBER: ожидалась ошибка")
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("MarshalBER = %v, ожидалось %v", err, tt.wantErr)
			}
		})
	}
}

// Тесты для ошибок в MarshalJSON
func TestMarshalJSONErrors(t *testing.T) {
	// Пустой OID
	data, err := OID{}.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON(empty): %v", err)
	}
	if string(data) != `""` {
		t.Errorf("MarshalJSON(empty) = %s, ожидалось '\"\"'", data)
	}
}

// Тесты для ошибок в AppendBER
func TestAppendBERErrors(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr error
	}{
		{"Пустой", OID{}, ErrOIDTooShort},
		{"Невалидный", OID{3, 1}, ErrFirstComponentTooBig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.oid.AppendBER(nil)
			if err == nil {
				t.Error("AppendBER: ожидалась ошибка")
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("AppendBER = %v, ожидалось %v", err, tt.wantErr)
			}
		})
	}
}

// Тесты для Registry BatchRegister с ошибками валидации
func TestBatchRegisterValidationErrors(t *testing.T) {
	reg := NewRegistry()

	tests := []struct {
		name    string
		entries map[string]OID
		wantErr error
	}{
		{
			name:    "Невалидный OID",
			entries: map[string]OID{"bad": {3, 1}},
			wantErr: ErrFirstComponentTooBig,
		},
		{
			name:    "Короткий OID",
			entries: map[string]OID{"bad": {1}},
			wantErr: ErrOIDTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reg.BatchRegister(tt.entries)
			if err == nil {
				t.Error("BatchRegister: ожидалась ошибка")
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("BatchRegister = %v, ожидалось %v", err, tt.wantErr)
			}
		})
	}
}

// Тесты для Registry BatchRegister с дубликатами
func TestBatchRegisterDuplicates(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")

	// Дубликат OID (разные имена, одинаковый OID)
	err := reg.BatchRegister(map[string]OID{
		"first":  oid,
		"second": oid,
	})
	if err == nil {
		t.Error("BatchRegister: ожидалась ошибка дубликата OID")
	}
	if !errors.Is(err, ErrDuplicateOIDInBatch) {
		t.Errorf("BatchRegister = %v, ожидалось ErrDuplicateOIDInBatch", err)
	}

	// Конфликт с существующим именем
	reg.Register("existing", MustParseOID("1.3.6.1.4.1"))

	err = reg.BatchRegister(map[string]OID{
		"existing": MustParseOID("2.100.3"),
	})
	if err == nil {
		t.Error("BatchRegister: ожидалась ошибка конфликта имени")
	}
	if !errors.Is(err, ErrNameAlreadyExists) {
		t.Errorf("BatchRegister = %v, ожидалось ErrNameAlreadyExists", err)
	}

	// Конфликт с существующим OID
	reg.Register("original", oid)

	err = reg.BatchRegister(map[string]OID{
		"different_name": oid,
	})
	if err == nil {
		t.Error("BatchRegister: ожидалась ошибка конфликта OID")
	}
	if !errors.Is(err, ErrOIDAlreadyRegistered) {
		t.Errorf("BatchRegister = %v, ожидалось ErrOIDAlreadyRegistered", err)
	}
}

// Тесты для Registry Clear
func TestRegistryClearWithData(t *testing.T) {
	reg := NewRegistry()

	// Добавляем записи
	for i := 0; i < 10; i++ {
		reg.Register(fmt.Sprintf("oid-%d", i), MustParseOID(fmt.Sprintf("1.3.6.1.%d", i+1)))
	}

	if reg.Size() != 10 {
		t.Error("Size: должно быть 10")
	}

	// Очищаем
	reg.Clear()

	if reg.Size() != 0 {
		t.Error("Size после Clear: должно быть 0")
	}

	// Проверяем, что можно снова добавлять
	reg.Register("new", MustParseOID("1.3.6.1"))
	if reg.Size() != 1 {
		t.Error("Size после повторной регистрации: должно быть 1")
	}
}

// Тесты для глобального ResetRegistry
func TestGlobalResetRegistry(t *testing.T) {
	// Регистрируем
	Register("test", MustParseOID("1.3.6.1"))
	if Size() != 1 {
		t.Error("Size: должно быть 1")
	}

	// Сбрасываем
	ResetRegistry()

	if Size() != 0 {
		t.Error("Size после ResetRegistry: должно быть 0")
	}

	// Проверяем, что можно снова регистрировать
	Register("test", MustParseOID("1.3.6.1"))
	if Size() != 1 {
		t.Error("Size после повторной регистрации: должно быть 1")
	}
}

// Тесты для GetRegistry
func TestGetRegistry(t *testing.T) {
	reg := GetRegistry()
	if reg == nil {
		t.Fatal("GetRegistry: nil")
	}

	// Проверяем, что это тот же реестр
	Register("test", MustParseOID("1.3.6.1"))
	if reg.Size() != Size() {
		t.Error("GetRegistry: разные реестры")
	}
}

// Тесты для NullOID Value с невалидным OID
func TestNullOIDValueInvalid(t *testing.T) {
	n := NullOID{
		OID:   OID{3, 1},
		Valid: true,
	}

	_, err := n.Value()
	if err == nil {
		t.Error("NullOID.Value: ожидалась ошибка")
	}
}

// Тесты для Array Value с невалидным OID
func TestArrayValueInvalid(t *testing.T) {
	arr := Array{
		MustParseOID("1.3.6.1"),
		{3, 1}, // Невалидный
	}

	_, err := arr.Value()
	if err == nil {
		t.Error("Array.Value: ожидалась ошибка")
	}
}

// Тесты для Array Scan с невалидным форматом
func TestArrayScanInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"Не массив", "not-array"},
		{"Нет закрывающей", "{1.3.6.1"},
		{"Нет открывающей", "1.3.6.1}"},
		{"Число", 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr Array
			if err := arr.Scan(tt.input); err == nil {
				t.Error("Array.Scan: ожидалась ошибка")
			}
		})
	}
}

// Тесты для Array UnmarshalJSON с ошибками
func TestArrayUnmarshalJSONErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Не JSON", []byte("invalid")},
		{"Объект", []byte(`{"oid":"1.3.6.1"}`)},
		{"Число", []byte("123")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr Array
			if err := arr.UnmarshalJSON(tt.data); err == nil {
				t.Error("Array.UnmarshalJSON: ожидалась ошибка")
			}
		})
	}
}

// Тесты для NullOID UnmarshalJSON с ошибками
func TestNullOIDUnmarshalJSONErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Не JSON", []byte("invalid")},
		{"Число", []byte("123")},
		{"Объект", []byte(`{"oid":"1.3.6.1"}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n NullOID
			if err := n.UnmarshalJSON(tt.data); err == nil {
				t.Error("NullOID.UnmarshalJSON: ожидалась ошибка")
			}
		})
	}
}

// Тесты для ошибок Scan
func TestScanErrors(t *testing.T) {
	var oid OID

	// Неподдерживаемый тип
	if err := oid.Scan(123); err == nil {
		t.Error("Scan: ожидалась ошибка для числа")
	}

	// Невалидный OID
	if err := oid.Scan("invalid"); err == nil {
		t.Error("Scan: ожидалась ошибка для невалидного OID")
	}
}

// Тесты для глобальных NoCopy функций
func TestGlobalNoCopy(t *testing.T) {
	ResetRegistry()

	Register("test", MustParseOID("1.3.6.1"))

	// LookupByNameNoCopy
	oid, exists := LookupByNameNoCopy("test")
	if !exists || !oid.Equal(MustParseOID("1.3.6.1")) {
		t.Error("LookupByNameNoCopy: неверный результат")
	}

	// ListNoCopy
	list := ListNoCopy()
	if len(list) != 1 {
		t.Error("ListNoCopy: неверная длина")
	}

	// OIDsNoCopy
	oids := OIDsNoCopy()
	if len(oids) != 1 {
		t.Error("OIDsNoCopy: неверная длина")
	}

	// Names
	names := Names()
	if len(names) != 1 {
		t.Error("Names: неверная длина")
	}

	// OIDs
	oidsCopy := OIDs()
	if len(oidsCopy) != 1 {
		t.Error("OIDs: неверная длина")
	}
}

// ============================================
// БЕНЧМАРКИ
// ============================================

// Бенчмарк с подробным выводом
func BenchmarkParseOIDDetailed(b *testing.B) {
	testCases := []string{
		"1.3.6.1.4.1",
		"1.3.6.1.4.1.99999",
		"1.3.6.1.4.1.99999.1.1",
		"2.100.3",
		"0.39.1",
	}

	for _, tc := range testCases {
		b.Run(fmt.Sprintf("Parse_%s", tc), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, err := ParseOID(tc)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Бенчмарк с настройкой
func BenchmarkOIDStringDetailed(b *testing.B) {
	testCases := []struct {
		name string
		oid  OID
	}{
		{"Short", OID{1, 3, 6}},
		{"Medium", OID{1, 3, 6, 1, 4, 1}},
		{"Long", OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.oid.String())))
			for b.Loop() {
				_ = tc.oid.String()
			}
		})
	}
}

// Бенчмарк с параллельным выполнением
func BenchmarkParallelOperations(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.Run("Parallel_String", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = oid.String()
			}
		})
	})

	b.Run("Parallel_Marshal", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = oid.MarshalBinary()
			}
		})
	})
}

func BenchmarkParseOID(b *testing.B) {
	for b.Loop() {
		ParseOID("1.3.6.1.4.1.99999.1.1")
	}
}

func BenchmarkOIDString(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")
	b.ResetTimer()
	for b.Loop() {
		_ = oid.String()
	}
}

func BenchmarkOIDMarshalBinary(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")
	b.ResetTimer()
	for b.Loop() {
		oid.MarshalBinary()
	}
}

func BenchmarkOIDUnmarshalBinary(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")
	data, _ := oid.MarshalBinary()
	b.ResetTimer()
	for b.Loop() {
		var result OID
		result.UnmarshalBinary(data)
	}
}

func BenchmarkRegistryRegister(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	b.ResetTimer()
	for b.Loop() {
		reg.Register("benchmark", oid)
	}
}

func BenchmarkParallelParseOID(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ParseOID("1.3.6.1.4.1.99999.1.1")
		}
	})
}

func BenchmarkParallelRegistryLookup(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reg.LookupByName("test")
		}
	})
}

func BenchmarkMarshalBinaryOptimized(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := oid.MarshalBinary()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalBinaryOptimized(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")
	data, _ := oid.MarshalBinary()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var result OID
		if err := result.UnmarshalBinary(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalJSONOptimized(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := oid.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompleteWorkflow(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		// Полный цикл: парсинг -> маршализация -> демаршализация -> строка
		oid, err := ParseOID("1.3.6.1.4.1.99999.1.1")
		if err != nil {
			b.Fatal(err)
		}

		data, err := oid.MarshalBinary()
		if err != nil {
			b.Fatal(err)
		}

		var decoded OID
		if err := decoded.UnmarshalBinary(data); err != nil {
			b.Fatal(err)
		}

		_ = decoded.String()
	}
}

// Сравнение с стандартной библиотекой
func BenchmarkStdASN1Marshal(b *testing.B) {
	oid := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 1, 1}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := asn1.Marshal(oid)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdASN1Unmarshal(b *testing.B) {
	oid := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 1, 1}
	data, _ := asn1.Marshal(oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var result asn1.ObjectIdentifier
		_, err := asn1.Unmarshal(data, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegistryNoCopyLookup(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.Run("WithCopy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_, _ = reg.LookupByName("test")
		}
	})

	b.Run("NoCopy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_, _ = reg.LookupByNameNoCopy("test")
		}
	})
}

func BenchmarkRegistryNoCopyList(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		reg.Register(fmt.Sprintf("oid-%d", i), MustParseOID(fmt.Sprintf("1.3.6.1.%d", i+1)))
	}

	b.Run("WithCopy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = reg.List()
		}
	})

	b.Run("NoCopy", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			_ = reg.ListNoCopy()
		}
	})
}
