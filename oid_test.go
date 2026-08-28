// oid/oid_test.go
package oid

import (
	"encoding/asn1"
	"encoding/json"
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

// Helper функция для тестов
func parentOrSelf(o OID) OID {
	if len(o) <= 1 {
		return o
	}
	return o[:len(o)-1]
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

func TestGlobalNoCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1.4.1")
	Register("test", oid)

	// Глобальный NoCopy
	noCopy, exists := LookupByNameNoCopy("test")
	if !exists {
		t.Fatal("OID не найден")
	}
	if !noCopy.Equal(oid) {
		t.Error("NoCopy OID не совпадает")
	}

	// Глобальный ListNoCopy
	list := ListNoCopy()
	if len(list) != 1 {
		t.Errorf("len = %d, ожидалось 1", len(list))
	}

	// Глобальный OIDsNoCopy
	oids := OIDsNoCopy()
	if len(oids) != 1 {
		t.Errorf("len = %d, ожидалось 1", len(oids))
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
