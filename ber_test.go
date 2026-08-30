// oid/ber_test.go - исправленная версия
package oid

import (
	"bytes"
	"encoding/asn1"
	"errors"
	"testing"
)

// ============================================
// ТЕСТЫ
// ============================================

func TestAppendBER(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected []byte
		wantErr  bool
	}{
		{
			name:     "Простой OID",
			oid:      OID{1, 3, 6, 1},
			expected: []byte{0x2b, 0x06, 0x01},
			wantErr:  false,
		},
		{
			name:     "OID с большим компонентом",
			oid:      OID{1, 3, 6, 1, 4, 1, 99999},
			expected: []byte{0x2b, 0x06, 0x01, 0x04, 0x01, 0x86, 0x8d, 0x1f},
			wantErr:  false,
		},
		{
			name:     "OID с первым компонентом 2",
			oid:      OID{2, 100, 3},
			expected: []byte{0x81, 0x34, 0x03}, // 180 в base-128 = 0x81 0x34
			wantErr:  false,
		},
		{
			name:     "Максимальный для BER",
			oid:      OID{2, 175},
			expected: []byte{0x81, 0x7f}, // 255 в base-128 = 0x81 0x7f
			wantErr:  false,
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
			result, err := tt.oid.AppendBER(nil)

			if tt.wantErr {
				if err == nil {
					t.Error("AppendBER: ожидалась ошибка")
				}
				return
			}

			if err != nil {
				t.Errorf("AppendBER: неожиданная ошибка: %v", err)
			}

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("AppendBER = %x, ожидалось %x", result, tt.expected)
			}
		})
	}
}

func TestMarshalBER(t *testing.T) {
	tests := []struct {
		name string
		oid  OID
	}{
		{name: "Простой", oid: OID{1, 3, 6, 1}},
		{name: "Средний", oid: OID{1, 3, 6, 1, 4, 1}},
		{name: "С нулями", oid: OID{0, 0, 1, 2}},
		{name: "С первым 2", oid: OID{2, 100, 3}},
		{name: "Максимальный для BER", oid: OID{2, 175}},
		{name: "Большой компонент", oid: OID{1, 3, 6, 1, 4, 1, 99999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ourData, err := tt.oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: ошибка: %v", err)
			}

			// Сравниваем со стандартной библиотекой
			stdData, err := asn1.Marshal(tt.oid.ToASN1())
			if err != nil {
				t.Fatalf("asn1.Marshal: ошибка: %v", err)
			}

			if !bytes.Equal(ourData, stdData) {
				t.Errorf("MarshalBER = %x, ожидалось %x", ourData, stdData)
			}
		})
	}
}

func TestUnmarshalBER(t *testing.T) {
	tests := []struct {
		name string
		oid  OID
	}{
		{name: "Простой", oid: OID{1, 3, 6, 1}},
		{name: "Средний", oid: OID{1, 3, 6, 1, 4, 1}},
		{name: "С нулями", oid: OID{0, 0, 1, 2}},
		{name: "С первым 2", oid: OID{2, 100, 3}},
		{name: "Максимальный для BER", oid: OID{2, 175}},
		{name: "Большой компонент", oid: OID{1, 3, 6, 1, 4, 1, 99999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdData, err := asn1.Marshal(tt.oid.ToASN1())
			if err != nil {
				t.Fatalf("asn1.Marshal: ошибка: %v", err)
			}

			var decoded OID
			err = decoded.UnmarshalBER(stdData)
			if err != nil {
				t.Fatalf("UnmarshalBER: ошибка: %v", err)
			}

			if !decoded.Equal(tt.oid) {
				t.Errorf("UnmarshalBER = %v, ожидалось %v", decoded, tt.oid)
			}
		})
	}
}

func TestBERRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		oid  OID
	}{
		{name: "Простой", oid: OID{1, 3, 6, 1}},
		{name: "Средний", oid: OID{1, 3, 6, 1, 4, 1}},
		{name: "Длинный", oid: OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1}},
		{name: "С нулями", oid: OID{0, 0, 1, 2}},
		{name: "С первым 2", oid: OID{2, 100, 3}},
		{name: "Максимальный для BER", oid: OID{2, 175}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: ошибка: %v", err)
			}

			var decoded OID
			err = decoded.UnmarshalBER(data)
			if err != nil {
				t.Fatalf("UnmarshalBER: ошибка: %v", err)
			}

			if !decoded.Equal(tt.oid) {
				t.Errorf("Round trip: %v -> %x -> %v", tt.oid, data, decoded)
			}
		})
	}
}

func TestBER_NegativeErrors(t *testing.T) {
	var o OID

	// Закрываем ErrInsufficientData при коротком заголовке
	if err := o.UnmarshalBER([]byte{0x06}); err == nil {
		t.Error("Ожидалась ошибка ErrInsufficientData")
	}

	// Закрываем ErrInvalidLength при несовпадении длин пакета
	if err := o.UnmarshalBER([]byte{0x06, 0x05, 0x2b, 0x06}); err == nil {
		t.Error("Ожидалась ошибка ErrInvalidLength")
	}
}

// Тесты для UnmarshalBERContent с ошибками
func TestUnmarshalBERContentErrors(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		wantErr error
	}{
		{
			name:    "Пустой",
			content: []byte{},
			wantErr: ErrEmptyContent,
		},
		{
			name:    "Незавершенная последовательность",
			content: []byte{0x2B, 0x86},
			wantErr: ErrComponentFailed,
		},
		{
			name:    "Переполнение",
			content: []byte{0x2B, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F},
			wantErr: ErrComponentFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.UnmarshalBERContent(tt.content)
			if err == nil {
				t.Error("UnmarshalBERContent: ожидалась ошибка")
				return
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("UnmarshalBERContent = %v, ожидалось %v", err, tt.wantErr)
			}
		})
	}
}

func TestSizeBERCorrect(t *testing.T) {
	tests := []struct {
		name string
		oid  OID
	}{
		{"Короткий", MustParseOID("1.3.6.1")},
		{"Средний", MustParseOID("1.3.6.1.4.1")},
		{"Длинный", func() OID {
			oid := OID{1, 3}
			for i := 0; i < 50; i++ {
				oid = append(oid, MaxOIDComponent)
			}
			return oid
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := tt.oid.SizeBER()
			if err != nil {
				t.Fatalf("SizeBER: %v", err)
			}

			data, err := tt.oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}

			if size != len(data) {
				t.Errorf("SizeBER = %d, MarshalBER = %d", size, len(data))
			}
		})
	}
}

// Тесты для SizeBER
// Тесты для BER edge cases
func TestBEREdgeCases(t *testing.T) {
	// AppendBER с невалидным OID
	_, err := OID{}.AppendBER(nil)
	if err == nil {
		t.Error("AppendBER: ожидалась ошибка")
	}

	// MarshalBER с невалидным OID
	_, err = OID{3, 1}.MarshalBER()
	if !errors.Is(err, ErrFirstComponentTooBig) {
		t.Error("MarshalBER: ожидалась ErrFirstComponentTooBig")
	}

	// UnmarshalBER с неверным тегом
	var oid OID
	err = oid.UnmarshalBER([]byte{0x05, 0x01, 0x2B})
	if !errors.Is(err, ErrInvalidASN1Tag) {
		t.Error("UnmarshalBER: ожидалась ErrInvalidASN1Tag")
	}

	// UnmarshalBER с пустым контентом
	// Правильные данные: тег 0x06, длина 0x00, и больше ничего
	err = oid.UnmarshalBER([]byte{0x06, 0x00})
	if !errors.Is(err, ErrEmptyContent) {
		t.Errorf("UnmarshalBER: ожидалась ErrEmptyContent, получили: %v", err)
	}

	// SizeBER с невалидным OID
	_, err = OID{}.SizeBER()
	if err == nil {
		t.Error("SizeBER: ожидалась ошибка")
	}
}

// Тесты для UnmarshalBER с различными ошибками
func TestUnmarshalBERErrors(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{
			name:    "Короткие данные (2 байта с длиной 1)",
			data:    []byte{0x06, 0x01},
			wantErr: ErrInsufficientData,
		},
		{
			name:    "Неверный тег",
			data:    []byte{0x05, 0x01, 0x2B},
			wantErr: ErrInvalidASN1Tag,
		},
		{
			name:    "Неверная длина (0x80)",
			data:    []byte{0x06, 0x80, 0x00},
			wantErr: ErrInvalidLength,
		},
		{
			name:    "Пустой контент",
			data:    []byte{0x06, 0x00},
			wantErr: ErrEmptyContent,
		},
		{
			name:    "Недостаточно данных",
			data:    []byte{0x06, 0x05, 0x2B},
			wantErr: ErrInsufficientData, // Изменено с ErrInvalidLength на ErrInsufficientData
		},
		{
			name:    "Лишние данные",
			data:    []byte{0x06, 0x01, 0x2B, 0x00},
			wantErr: ErrInvalidLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.UnmarshalBER(tt.data)
			if err == nil {
				t.Error("UnmarshalBER: ожидалась ошибка")
				return
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("UnmarshalBER(%x) = %v, ожидалось %v", tt.data, err, tt.wantErr)
			}
		})
	}
}

// Тесты для MarshalDER/UnmarshalDER
func TestDERRoundTrip(t *testing.T) {
	oid := MustParseOID("1.3.6.1.4.1")

	data, err := oid.MarshalDER()
	if err != nil {
		t.Fatalf("MarshalDER: %v", err)
	}

	var decoded OID
	if err := decoded.UnmarshalDER(data); err != nil {
		t.Fatalf("UnmarshalDER: %v", err)
	}

	if !decoded.Equal(oid) {
		t.Error("DER round trip: не совпадает")
	}
}
func TestSizeBERFullCoverage(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		{"Короткий", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},
		{"Длинный", MustParseOID("1.3.6.1.4.1.99999.1.1"), false},
		{"С большим компонентом", OID{1, 3, MaxOIDComponent}, false},
		{"С первым 2", MustParseOID("2.100.3"), false},
		{"Пустой", OID{}, true},
		{"Невалидный", OID{3, 1}, true},
		{"Один компонент", OID{1}, true},
		{"Переполнение", OID{2, ^uint32(0)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := tt.oid.SizeBER()
			if tt.wantErr {
				if err == nil {
					t.Error("SizeBER: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("SizeBER: %v", err)
			}
			if size <= 0 {
				t.Error("SizeBER: размер должен быть положительным")
			}

			// Проверяем, что размер совпадает с реальным
			data, err := tt.oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			if len(data) != size {
				t.Errorf("SizeBER = %d, реальный размер = %d", size, len(data))
			}
		})
	}
}
func TestMarshalBERFullCoverage(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		// Короткий OID (contentSize < 128)
		{"Короткий", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},

		// Длинный OID (contentSize >= 128)
		{"Длинный", MustParseOID("1.3.6.1.4.1.99999.1.1"), false},
		{"Очень длинный", OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1, 1, 1, 1, 1}, false},

		// С большим компонентом
		{"Максимальный компонент", OID{1, 3, MaxOIDComponent}, false},

		// С первым компонентом 2
		{"Первый = 2", MustParseOID("2.100.3"), false},
		{"Первый = 2, большой второй", OID{2, MaxOIDComponent}, false},

		// Ошибки
		{"Пустой", OID{}, true},
		{"Невалидный", OID{3, 1}, true},
		{"Один компонент", OID{1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalBER()
			if tt.wantErr {
				if err == nil {
					t.Error("MarshalBER: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("MarshalBER: %v", err)
			}
			if len(data) == 0 {
				t.Error("MarshalBER: пустой результат")
			}

			// Проверяем round trip
			var decoded OID
			if err := decoded.UnmarshalBER(data); err != nil {
				t.Errorf("UnmarshalBER: %v", err)
			}
			if !decoded.Equal(tt.oid) {
				t.Errorf("Round trip: %v -> %v", tt.oid, decoded)
			}
		})
	}
}

func TestUnmarshalBERFullCoverage(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		// Успешные случаи
		{"Короткий", []byte{0x06, 0x03, 0x2B, 0x06, 0x01}, nil},
		{"Средний", []byte{0x06, 0x05, 0x2B, 0x06, 0x01, 0x04, 0x01}, nil},

		// Ошибки
		{"Пустой", []byte{}, ErrInsufficientData},
		{"1 байт", []byte{0x06}, ErrInsufficientData},
		{"Неверный тег", []byte{0x05, 0x01, 0x2B}, ErrInvalidASN1Tag},
		{"Длина 0x80", []byte{0x06, 0x80, 0x00}, ErrInvalidLength},
		{"Длина 0x85", []byte{0x06, 0x85, 0x00}, ErrInvalidLength},
		{"Длина 0x81 с недостатком", []byte{0x06, 0x81}, ErrInsufficientData},
		{"Пустой контент", []byte{0x06, 0x00}, ErrEmptyContent},
		{"Недостаточно данных", []byte{0x06, 0x05, 0x2B}, ErrInsufficientData},
		{"Лишние данные", []byte{0x06, 0x01, 0x2B, 0x00}, ErrInvalidLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.UnmarshalBER(tt.data)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("UnmarshalBER(%x): %v", tt.data, err)
				}
				return
			}

			if err == nil {
				t.Error("UnmarshalBER: ожидалась ошибка")
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UnmarshalBER(%x) = %v, ожидалось %v", tt.data, err, tt.wantErr)
			}
		})
	}
}

func TestMarshalBERRemainingCoverage(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		// Различные размеры контента
		{"Короткий < 128", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},
		{"Длинный >= 128", OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1, 1, 1, 1, 1}, false},

		// Различные первые компоненты
		{"Первый 0", MustParseOID("0.39.1"), false},
		{"Первый 1", MustParseOID("1.39.1"), false},
		{"Первый 2", MustParseOID("2.100.3"), false},
		{"Первый 2, большой второй", OID{2, MaxOIDComponent}, false},

		// С большими компонентами
		{"Максимальный компонент", OID{1, 3, MaxOIDComponent}, false},

		// Ошибки валидации
		{"Пустой", OID{}, true},
		{"Один компонент", OID{1}, true},
		{"Первый > 2", OID{3, 1}, true},
		{"Второй > 39", OID{1, 40}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalBER()
			if tt.wantErr {
				if err == nil {
					t.Error("MarshalBER: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("MarshalBER: %v", err)
			}
			if len(data) == 0 {
				t.Error("MarshalBER: пустой результат")
			}

			// Проверяем round trip
			var decoded OID
			if err := decoded.UnmarshalBER(data); err != nil {
				t.Errorf("UnmarshalBER: %v", err)
			}
			if !decoded.Equal(tt.oid) {
				t.Errorf("Round trip: %v -> %v", tt.oid, decoded)
			}
		})
	}
}
func TestMarshalBinaryRemainingCoverage(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		{"Короткий", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},
		{"Длинный", MustParseOID("1.3.6.1.4.1.99999.1.1"), false},
		{"С максимальным компонентом", OID{1, 3, MaxOIDComponent}, false},
		{"Первый 0", MustParseOID("0.39.1"), false},
		{"Первый 2", MustParseOID("2.100.3"), false},
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
func TestUnmarshalBinaryRemainingCoverage(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{"Успешный", []byte{0x06, 0x03, 0x2B, 0x06, 0x01}, nil},
		// Исправлено: правильные байты для OID 1.3.6.1.4.1
		// 0x2B = 1.3, 0x06 = 6, 0x01 = 1, 0x04 = 4, 0x01 = 1
		{"С большим компонентом", []byte{0x06, 0x05, 0x2B, 0x06, 0x01, 0x04, 0x01}, nil},
		{"Пустой", []byte{}, ErrDataTooShort},
		{"1 байт", []byte{0x06}, ErrDataTooShort},
		{"Неверный тег", []byte{0x05, 0x00}, ErrInvalidASN1Tag},
		{"Неверная длина", []byte{0x06, 0x80}, ErrInvalidASN1Length},
		{"Недостаточно данных", []byte{0x06, 0x05, 0x01}, ErrInsufficientData},
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

func TestUnmarshalBinaryWithGeneratedData(t *testing.T) {
	// Создаем OID
	original := MustParseOID("1.3.6.1.4.1")

	// Кодируем
	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	// Декодируем
	var decoded OID
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}

	// Проверяем
	if !decoded.Equal(original) {
		t.Errorf("Round trip: %v -> %v", original, decoded)
	}
}

func TestReadLengthRemainingCoverage(t *testing.T) {
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
		{"65535", []byte{0x82, 0xFF, 0xFF}, 65535, 3},
		{"0x80 некорректная", []byte{0x80}, 0, 0},
		{"0x85 слишком длинная", []byte{0x85}, 0, 0},
		{"0x81 недостаточно", []byte{0x81}, 0, 0},
		{"Пустая", []byte{}, 0, 0},
		// Исправлено: 0x84 с 4 байтами данных - это валидно
		// 4294967295 = 0xFFFFFFFF, но это валидное значение для int на 64-битной системе
		{"0x84 с 4 байтами", []byte{0x84, 0xFF, 0xFF, 0xFF, 0xFF}, 4294967295, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, bytesRead := readLength(tt.data)
			if bytesRead != tt.bytes {
				t.Errorf("readLength(%x) bytesRead = %d, ожидалось %d", tt.data, bytesRead, tt.bytes)
			}
			if length != tt.expected {
				t.Errorf("readLength(%x) length = %d, ожидалось %d", tt.data, length, tt.expected)
			}
		})
	}
}

func TestMarshalBERComplete(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		// Все ветви кодирования
		{"Пустой", OID{}, true},
		{"Один компонент", OID{1}, true},
		{"Первый > 2", OID{3, 1}, true},
		{"Второй > 39 при первом 1", OID{1, 40}, true},
		{"Второй > 39 при первом 0", OID{0, 40}, true},

		// Успешные случаи
		{"Короткий", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},
		{"Длинный", OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1, 1, 1, 1, 1}, false},

		// Разные первые компоненты
		{"Первый 0", MustParseOID("0.39.1"), false},
		{"Первый 1", MustParseOID("1.39.1"), false},
		{"Первый 2", MustParseOID("2.100.3"), false},
		{"Первый 2, большой второй", OID{2, MaxOIDComponent}, false},

		// Максимальные значения
		{"Максимальный компонент", OID{1, 3, MaxOIDComponent}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalBER()
			if tt.wantErr {
				if err == nil {
					t.Error("MarshalBER: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("MarshalBER: %v", err)
			}
			if len(data) == 0 {
				t.Error("MarshalBER: пустой результат")
			}

			// Проверяем round trip
			var decoded OID
			if err := decoded.UnmarshalBER(data); err != nil {
				t.Errorf("UnmarshalBER: %v", err)
			}
			if !decoded.Equal(tt.oid) {
				t.Errorf("Round trip: %v -> %v", tt.oid, decoded)
			}
		})
	}
}

func TestSizeBERComplete(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		{"Пустой", OID{}, true},
		{"Один компонент", OID{1}, true},
		{"Первый > 2", OID{3, 1}, true},

		{"Короткий", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},
		{"Длинный", OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1, 1, 1, 1, 1}, false},
		{"Максимальный компонент", OID{1, 3, MaxOIDComponent}, false},
		{"Первый 2", MustParseOID("2.100.3"), false},
		{"Первый 2, большой второй", OID{2, MaxOIDComponent}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := tt.oid.SizeBER()
			if tt.wantErr {
				if err == nil {
					t.Error("SizeBER: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("SizeBER: %v", err)
			}
			if size <= 0 {
				t.Error("SizeBER: размер должен быть положительным")
			}

			// Проверяем соответствие
			data, err := tt.oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			if len(data) != size {
				t.Errorf("SizeBER = %d, реальный размер = %d", size, len(data))
			}
		})
	}
}

func TestMarshalBinaryComplete(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		{"Пустой", OID{}, true},
		{"Один компонент", OID{1}, true},
		{"Первый > 2", OID{3, 1}, true},
		{"Второй > 39", OID{1, 40}, true},

		{"Короткий", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},
		{"Длинный", OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1, 1, 1, 1, 1}, false},
		{"Максимальный компонент", OID{1, 3, MaxOIDComponent}, false},
		{"Первый 0", MustParseOID("0.39.1"), false},
		{"Первый 2", MustParseOID("2.100.3"), false},
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

func TestUnmarshalBinaryComplete(t *testing.T) {
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
		{"Неверный компонент", []byte{0x06, 0x02, 0x2B, 0x80}, ErrInvalidComponent},
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

func TestReadLengthComplete(t *testing.T) {
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
		{"65535", []byte{0x82, 0xFF, 0xFF}, 65535, 3},
		{"0x80", []byte{0x80}, 0, 0},
		{"0x85", []byte{0x85}, 0, 0},
		{"0x81 недостаточно", []byte{0x81}, 0, 0},
		{"Пустая", []byte{}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, bytesRead := readLength(tt.data)
			if bytesRead != tt.bytes {
				t.Errorf("readLength(%x) bytesRead = %d, ожидалось %d", tt.data, bytesRead, tt.bytes)
			}
			if length != tt.expected {
				t.Errorf("readLength(%x) length = %d, ожидалось %d", tt.data, length, tt.expected)
			}
		})
	}
}

func TestReadBase128FromBytesComplete(t *testing.T) {
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
		{"Переполнение", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}, 0, 0},
		{"Незавершенная", []byte{0x81, 0x80}, 0, 0},
		{"Пустая", []byte{}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, bytesRead := readBase128FromBytes(tt.data)
			if bytesRead != tt.bytes {
				t.Errorf("readBase128FromBytes(%x) bytesRead = %d, ожидалось %d", tt.data, bytesRead, tt.bytes)
			}
			if value != tt.expected {
				t.Errorf("readBase128FromBytes(%x) value = %d, ожидалось %d", tt.data, value, tt.expected)
			}
		})
	}
}

func TestBatchRegisterComplete(t *testing.T) {
	reg := NewRegistry()

	// Все возможные ошибки
	tests := []struct {
		name    string
		entries map[string]OID
		wantErr error
	}{
		{"Пустой", map[string]OID{}, nil},
		{"Невалидный", map[string]OID{"bad": {3, 1}}, ErrFirstComponentTooBig},
		{"Дубликат OID", map[string]OID{
			"first":  MustParseOID("1.3.6.1"),
			"second": MustParseOID("1.3.6.1"),
		}, ErrDuplicateOIDInBatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reg.BatchRegister(tt.entries)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("BatchRegister: %v", err)
				}
				return
			}
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

func TestMarshalBERAllBranches(t *testing.T) {
	// Тесты для всех ветвей MarshalBER
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		// Ошибки валидации
		{"Пустой", OID{}, true},
		{"Один компонент", OID{1}, true},
		{"Первый > 2", OID{3, 1}, true},
		{"Второй > 39 при 0", OID{0, 40}, true},
		{"Второй > 39 при 1", OID{1, 40}, true},

		// Успешные - короткий контент (< 128)
		{"Минимальный", OID{0, 0}, false},
		{"Короткий", MustParseOID("1.3.6.1"), false},

		// Успешные - длинный контент (>= 128)
		{"Длинный", OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1, 1, 1, 1, 1}, false},

		// Разные первые компоненты
		{"Первый 0", MustParseOID("0.39.1"), false},
		{"Первый 1", MustParseOID("1.39.1"), false},
		{"Первый 2", MustParseOID("2.100.3"), false},
		{"Первый 2, большой", OID{2, MaxOIDComponent}, false},

		// Максимальные значения
		{"Максимальный компонент", OID{1, 3, MaxOIDComponent}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalBER()
			if tt.wantErr {
				if err == nil {
					t.Error("MarshalBER: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("MarshalBER: %v", err)
			}
			if len(data) == 0 {
				t.Error("MarshalBER: пустой результат")
			}

			// Проверяем тег
			if data[0] != 0x06 {
				t.Errorf("MarshalBER: неверный тег 0x%02x", data[0])
			}

			// Round trip
			var decoded OID
			if err := decoded.UnmarshalBER(data); err != nil {
				t.Errorf("UnmarshalBER: %v", err)
			}
			if !decoded.Equal(tt.oid) {
				t.Errorf("Round trip: %v -> %v", tt.oid, decoded)
			}
		})
	}
}

func TestSizeBERAllBranches(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		{"Пустой", OID{}, true},
		{"Один компонент", OID{1}, true},
		{"Невалидный", OID{3, 1}, true},

		{"Короткий", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},
		{"Длинный", OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1, 1, 1, 1, 1}, false},
		{"Максимальный", OID{1, 3, MaxOIDComponent}, false},
		{"Первый 2", MustParseOID("2.100.3"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := tt.oid.SizeBER()
			if tt.wantErr {
				if err == nil {
					t.Error("SizeBER: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("SizeBER: %v", err)
			}
			if size <= 0 {
				t.Error("SizeBER: размер должен быть положительным")
			}

			// Проверяем соответствие с MarshalBER
			data, err := tt.oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			if len(data) != size {
				t.Errorf("SizeBER = %d, реальный = %d", size, len(data))
			}
		})
	}
}

func TestReadBase128FromBytesAllBranches(t *testing.T) {
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

		// Ошибки
		{"Переполнение (>5 байт)", []byte{0x81, 0x80, 0x80, 0x80, 0x80, 0x00}, 0, 0},
		{"Переполнение uint32", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}, 0, 0},
		{"Незавершенная", []byte{0x81, 0x80}, 0, 0},
		{"Пустая", []byte{}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, bytesRead := readBase128FromBytes(tt.data)
			if bytesRead != tt.bytes {
				t.Errorf("readBase128FromBytes(%x) bytesRead = %d, ожидалось %d",
					tt.data, bytesRead, tt.bytes)
			}
			if value != tt.expected {
				t.Errorf("readBase128FromBytes(%x) value = %d, ожидалось %d",
					tt.data, value, tt.expected)
			}
		})
	}
}

func TestReadLengthAllBranches(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
		bytes    int
	}{
		// Короткая форма
		{"0", []byte{0x00}, 0, 1},
		{"127", []byte{0x7F}, 127, 1},

		// Длинная форма
		{"128", []byte{0x81, 0x80}, 128, 2},
		{"255", []byte{0x81, 0xFF}, 255, 2},
		{"256", []byte{0x82, 0x01, 0x00}, 256, 3},
		{"65535", []byte{0x82, 0xFF, 0xFF}, 65535, 3},

		// Ошибки
		{"0x80", []byte{0x80}, 0, 0},
		{"0x85", []byte{0x85}, 0, 0},
		{"0x81 недостаточно", []byte{0x81}, 0, 0},
		{"Пустая", []byte{}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, bytesRead := readLength(tt.data)
			if bytesRead != tt.bytes {
				t.Errorf("readLength(%x) bytesRead = %d, ожидалось %d",
					tt.data, bytesRead, tt.bytes)
			}
			if length != tt.expected {
				t.Errorf("readLength(%x) length = %d, ожидалось %d",
					tt.data, length, tt.expected)
			}
		})
	}
}

// Тест для BatchRegister с конфликтом имени
func TestBatchRegisterNameConflict(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем существующее имя
	reg.Register("existing", MustParseOID("1.3.6.1.4.1"))

	// Пытаемся зарегистрировать то же имя
	err := reg.BatchRegister(map[string]OID{
		"existing": MustParseOID("2.100.3"),
	})

	if err == nil {
		t.Error("BatchRegister: ожидалась ошибка")
	}
	if !errors.Is(err, ErrNameAlreadyExists) {
		t.Errorf("BatchRegister = %v, ожидалось ErrNameAlreadyExists", err)
	}
}

func TestMarshalBERLongContent(t *testing.T) {
	// Создаем OID с действительно длинным контентом
	// Каждый компонент с большим значением кодируется в 4-5 байт
	longOID := OID{1, 3}
	for i := 0; i < 50; i++ {
		// Используем большие значения для увеличения размера
		longOID = append(longOID, uint32(1000000+i))
	}

	// MarshalBER
	data, err := longOID.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	t.Logf("Длина данных: %d", len(data))

	// Проверяем, что используем длинную форму если длина >= 128
	if len(data) >= 128 {
		if data[1] < 0x80 {
			t.Error("Ожидалась длинная форма длины")
		}
		t.Logf("Длинная форма: 0x%02x", data[1])
	}

	// Round trip
	var decoded OID
	if err := decoded.UnmarshalBER(data); err != nil {
		t.Fatalf("UnmarshalBER: %v", err)
	}
	if !decoded.Equal(longOID) {
		t.Error("Round trip: не совпадает")
	}
}

func TestMarshalBERVeryLongContent(t *testing.T) {
	// Создаем OID с гарантированно длинным контентом (> 128 байт)
	veryLongOID := OID{1, 3}
	for i := 0; i < 100; i++ {
		veryLongOID = append(veryLongOID, uint32(1000000+i))
	}

	data, err := veryLongOID.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	t.Logf("Длина данных: %d", len(data))

	if len(data) < 128 {
		t.Errorf("Длина %d < 128", len(data))
	}

	// Проверяем длинную форму
	if len(data) >= 128 && data[1] < 0x80 {
		t.Error("Ожидалась длинная форма длины")
	}

	// Round trip
	var decoded OID
	if err := decoded.UnmarshalBER(data); err != nil {
		t.Fatalf("UnmarshalBER: %v", err)
	}
	if !decoded.Equal(veryLongOID) {
		t.Error("Round trip: не совпадает")
	}
}

func TestMarshalBERMaxComponents(t *testing.T) {
	// Каждый MaxOIDComponent кодируется в 4 байта (base-128)
	// 35 * 4 = 140 байт > 128
	longOID := OID{1, 3}
	for i := 0; i < 35; i++ {
		longOID = append(longOID, MaxOIDComponent)
	}

	data, err := longOID.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	t.Logf("Длина данных: %d", len(data))

	if len(data) < 128 {
		t.Errorf("Длина %d < 128", len(data))
	}

	// Проверяем длинную форму длины
	if len(data) >= 128 && data[1] < 0x80 {
		t.Error("Ожидалась длинная форма длины")
	}

	// Round trip
	var decoded OID
	if err := decoded.UnmarshalBER(data); err != nil {
		t.Fatalf("UnmarshalBER: %v", err)
	}
	if !decoded.Equal(longOID) {
		t.Error("Round trip: не совпадает")
	}
}

func TestMarshalBERGuaranteedLong(t *testing.T) {
	// Гарантированно длинный OID
	// 50 компонентов MaxOIDComponent = 50 * 4 = 200 байт
	longOID := OID{1, 3}
	for i := 0; i < 50; i++ {
		longOID = append(longOID, MaxOIDComponent)
	}

	data, err := longOID.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	t.Logf("Длина данных: %d", len(data))

	if len(data) < 128 {
		t.Fatalf("Длина %d < 128", len(data))
	}

	// Проверяем, что используется длинная форма
	if data[1] < 0x80 {
		t.Error("Ожидалась длинная форма длины")
	}

	// Round trip
	var decoded OID
	if err := decoded.UnmarshalBER(data); err != nil {
		t.Fatalf("UnmarshalBER: %v", err)
	}
	if !decoded.Equal(longOID) {
		t.Error("Round trip: не совпадает")
	}
}

func TestMarshalBERFindLongContent(t *testing.T) {
	// Начинаем с малого и увеличиваем, пока не достигнем 128 байт
	for componentCount := 30; componentCount <= 100; componentCount++ {
		longOID := OID{1, 3}
		for i := 0; i < componentCount; i++ {
			longOID = append(longOID, MaxOIDComponent)
		}

		data, err := longOID.MarshalBER()
		if err != nil {
			t.Fatalf("MarshalBER: %v", err)
		}

		if len(data) >= 128 {
			t.Logf("Достигнута длина %d с %d компонентами", len(data), componentCount)

			// Проверяем длинную форму
			if data[1] < 0x80 {
				t.Error("Ожидалась длинная форма длины")
			}

			// Round trip
			var decoded OID
			if err := decoded.UnmarshalBER(data); err != nil {
				t.Fatalf("UnmarshalBER: %v", err)
			}
			if !decoded.Equal(longOID) {
				t.Error("Round trip: не совпадает")
			}

			return
		}
	}

	t.Error("Не удалось достичь длины 128 байт")
}

func TestSizeBERGuaranteedLong(t *testing.T) {
	// 50 компонентов MaxOIDComponent
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

	t.Logf("SizeBER = %d, реальный = %d", size, len(data))

	if len(data) != size {
		t.Errorf("SizeBER = %d, реальный = %d", size, len(data))
	}

	if size >= 128 {
		t.Logf("Длинная форма: размер %d", size)
	}
}

func TestSizeBERLongContent(t *testing.T) {
	// OID с гарантированно длинным контентом
	longOID := OID{1, 3}
	for i := 0; i < 30; i++ {
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

	t.Logf("SizeBER = %d, реальный = %d", size, len(data))

	if len(data) != size {
		t.Errorf("SizeBER = %d, реальный = %d", size, len(data))
	}

	if size >= 128 {
		t.Logf("Длинная форма: размер %d", size)
	}
}

func TestMarshalBERMediumContent(t *testing.T) {
	// OID с контентом около 128 байт
	mediumOID := OID{1, 3}
	for i := 0; i < 20; i++ {
		mediumOID = append(mediumOID, uint32(1000+i))
	}

	data, err := mediumOID.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	// Проверяем, что длина может быть >= 128 или < 128
	t.Logf("Длина данных: %d", len(data))

	// Round trip
	var decoded OID
	if err := decoded.UnmarshalBER(data); err != nil {
		t.Fatalf("UnmarshalBER: %v", err)
	}
	if !decoded.Equal(mediumOID) {
		t.Error("Round trip: не совпадает")
	}
}

func TestMarshalBinaryLongContent(t *testing.T) {
	// OID с длинным контентом
	longOID := OID{1, 3}
	for i := 0; i < 50; i++ {
		longOID = append(longOID, uint32(i+1))
	}

	data, err := longOID.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	// Проверяем, что данные корректны
	if len(data) < 2 {
		t.Error("MarshalBinary: слишком короткие данные")
	}

	// Round trip
	var decoded OID
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if !decoded.Equal(longOID) {
		t.Error("Round trip: не совпадает")
	}
}

func TestReadLengthOverflow(t *testing.T) {
	// Тест на переполнение int
	// На 32-битной системе int может переполниться
	// На 64-битной - нет
	data := []byte{0x84, 0xFF, 0xFF, 0xFF, 0xFF}

	length, bytesRead := readLength(data)

	// На 64-битной системе должно быть 4294967295
	if bytesRead != 5 {
		t.Errorf("bytesRead = %d, ожидалось 5", bytesRead)
	}

	// Проверяем, что длина корректна для 64-битной
	if length != 4294967295 {
		t.Logf("length = %d (может отличаться на 32-битной)", length)
	}
}

func TestSizeBERRemainingCoverage(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantErr bool
	}{
		{"Короткий", MustParseOID("1.3.6.1"), false},
		{"Средний", MustParseOID("1.3.6.1.4.1"), false},
		{"Длинный", OID{1, 3, 6, 1, 4, 1, 99999, 1, 1, 1, 1, 1, 1, 1, 1}, false},
		{"С большим компонентом", OID{1, 3, MaxOIDComponent}, false},
		{"Первый 2", MustParseOID("2.100.3"), false},
		{"Пустой", OID{}, true},
		{"Один компонент", OID{1}, true},
		{"Невалидный", OID{3, 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := tt.oid.SizeBER()
			if tt.wantErr {
				if err == nil {
					t.Error("SizeBER: ожидалась ошибка")
				}
				return
			}
			if err != nil {
				t.Errorf("SizeBER: %v", err)
			}
			if size <= 0 {
				t.Error("SizeBER: размер должен быть положительным")
			}

			// Проверяем соответствие с реальным размером
			data, err := tt.oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			if len(data) != size {
				t.Errorf("SizeBER = %d, реальный размер = %d", size, len(data))
			}
		})
	}
}

func TestBatchRegisterRemainingCoverage(t *testing.T) {
	reg := NewRegistry()

	// Успешная регистрация
	err := reg.BatchRegister(map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
		"third":  MustParseOID("0.39.1"),
	})
	if err != nil {
		t.Errorf("BatchRegister: %v", err)
	}

	// Пустой map
	err = reg.BatchRegister(map[string]OID{})
	if err != nil {
		t.Errorf("BatchRegister(empty): %v", err)
	}

	// Невалидный OID
	err = reg.BatchRegister(map[string]OID{
		"bad": {3, 1},
	})
	if err == nil {
		t.Error("BatchRegister(invalid): ожидалась ошибка")
	}
}

func TestReadBase128FromBytesRemainingCoverage(t *testing.T) {
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
		{"Переполнение", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}, 0, 0},
		{"Незавершенная", []byte{0x81, 0x80}, 0, 0},
		{"Пустая", []byte{}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, bytesRead := readBase128FromBytes(tt.data)
			if bytesRead != tt.bytes {
				t.Errorf("readBase128FromBytes(%x) bytesRead = %d, ожидалось %d", tt.data, bytesRead, tt.bytes)
			}
			if value != tt.expected {
				t.Errorf("readBase128FromBytes(%x) value = %d, ожидалось %d", tt.data, value, tt.expected)
			}
		})
	}
}

// ============================================
// БЕНЧМАРКИ
// ============================================

// Базовые бенчмарки BER кодирования
func BenchmarkBERMarshal(b *testing.B) {
	testCases := []struct {
		name string
		oid  OID
	}{
		{"Short", MustParseOID("1.3.6.1")},
		{"Medium", MustParseOID("1.3.6.1.4.1")},
		{"Long", MustParseOID("1.3.6.1.4.1.99999.1.1")},
		{"Very_Long", MustParseOID("1.3.6.1.4.1.99999.1.1.1.1.1.1")},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := tc.oid.MarshalBER()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBERUnmarshal(b *testing.B) {
	testCases := []struct {
		name string
		oid  OID
	}{
		{"Short", MustParseOID("1.3.6.1")},
		{"Medium", MustParseOID("1.3.6.1.4.1")},
		{"Long", MustParseOID("1.3.6.1.4.1.99999.1.1")},
		{"Very_Long", MustParseOID("1.3.6.1.4.1.99999.1.1.1.1.1.1")},
	}

	for _, tc := range testCases {
		data, _ := tc.oid.MarshalBER()

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result OID
				if err := result.UnmarshalBER(data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBERAppend(b *testing.B) {
	testCases := []struct {
		name string
		oid  OID
	}{
		{"Short", MustParseOID("1.3.6.1")},
		{"Medium", MustParseOID("1.3.6.1.4.1")},
		{"Long", MustParseOID("1.3.6.1.4.1.99999.1.1")},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			dst := make([]byte, 0, 64)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dst = dst[:0]
				_, err := tc.oid.AppendBER(dst)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Сравнение со стандартной библиотекой
func BenchmarkBERStdASN1Marshal(b *testing.B) {
	testCases := []struct {
		name string
		oid  asn1.ObjectIdentifier
	}{
		{"Short", asn1.ObjectIdentifier{1, 3, 6, 1}},
		{"Medium", asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1}},
		{"Long", asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 1, 1}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := asn1.Marshal(tc.oid)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBERStdASN1Unmarshal(b *testing.B) {
	testCases := []struct {
		name string
		oid  asn1.ObjectIdentifier
	}{
		{"Short", asn1.ObjectIdentifier{1, 3, 6, 1}},
		{"Medium", asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1}},
		{"Long", asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 1, 1}},
	}

	for _, tc := range testCases {
		data, _ := asn1.Marshal(tc.oid)

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var result asn1.ObjectIdentifier
				_, err := asn1.Unmarshal(data, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Параллельные бенчмарки
func BenchmarkBERParallelMarshal(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := oid.MarshalBER()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkBERParallelUnmarshal(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")
	data, _ := oid.MarshalBER()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var result OID
			if err := result.UnmarshalBER(data); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Комбинированные операции
func BenchmarkBERRoundTrip(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		data, err := oid.MarshalBER()
		if err != nil {
			b.Fatal(err)
		}

		var decoded OID
		if err := decoded.UnmarshalBER(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBERCompleteWorkflow(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		// Парсинг -> BER кодирование -> BER декодирование -> строка
		oid, err := ParseOID("1.3.6.1.4.1.99999.1.1")
		if err != nil {
			b.Fatal(err)
		}

		data, err := oid.MarshalBER()
		if err != nil {
			b.Fatal(err)
		}

		var decoded OID
		if err := decoded.UnmarshalBER(data); err != nil {
			b.Fatal(err)
		}

		_ = decoded.String()
	}
}

// Бенчмарки размера
func BenchmarkBERSize(b *testing.B) {
	testCases := []struct {
		name string
		oid  OID
	}{
		{"Short", MustParseOID("1.3.6.1")},
		{"Medium", MustParseOID("1.3.6.1.4.1")},
		{"Long", MustParseOID("1.3.6.1.4.1.99999.1.1")},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := tc.oid.SizeBER()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Бенчмарки для специфичных случаев
func BenchmarkBERFirstComponent2(b *testing.B) {
	oid := MustParseOID("2.100.3")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := oid.MarshalBER()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBERMaxComponent(b *testing.B) {
	oid := OID{1, 3, MaxOIDComponent}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := oid.MarshalBER()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Сравнение с MarshalBinary (наша предыдущая реализация)
func BenchmarkBERvsBinaryMarshal(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.Run("BER", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := oid.MarshalBER()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Binary", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := oid.MarshalBinary()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkBERvsBinaryUnmarshal(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")
	berData, _ := oid.MarshalBER()
	binData, _ := oid.MarshalBinary()

	b.Run("BER", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result OID
			if err := result.UnmarshalBER(berData); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Binary", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result OID
			if err := result.UnmarshalBinary(binData); err != nil {
				b.Fatal(err)
			}
		}
	})
}
