// oid/ber_test.go - исправленная версия
package oid

import (
	"bytes"
	"encoding/asn1"
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
