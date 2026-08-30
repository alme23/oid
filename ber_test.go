package oid

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestOIDAppendBER(t *testing.T) {
	tests := []struct {
		name     string
		oid      OID
		expected []byte
		wantErr  error
	}{
		{
			name:     "Стандартный OID",
			oid:      OID{1, 3, 6, 1, 4, 1},
			expected: []byte{0x2B, 0x06, 0x01, 0x04, 0x01},
			wantErr:  nil,
		},
		{
			name:     "Короткий OID",
			oid:      OID{1, 3, 6},
			expected: []byte{0x2B, 0x06},
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			oid:      OID{2, 100, 3},
			expected: []byte{0x81, 0x34, 0x03},
			wantErr:  nil,
		},
		{
			name:     "С первым 0",
			oid:      OID{0, 39, 1},
			expected: []byte{0x27, 0x01},
			wantErr:  nil,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: nil,
			wantErr:  ErrOIDTooShort,
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
			result, err := tt.oid.AppendBER(nil)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("AppendBER: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("AppendBER = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("AppendBER: %v", err)
				return
			}

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("AppendBER = %x, want %x", result, tt.expected)
			}
		})
	}
}

func TestOIDAppendBERAppendToExisting(t *testing.T) {
	oid := OID{1, 3, 6}

	dst := []byte{0xAA, 0xBB}
	result, err := oid.AppendBER(dst)

	if err != nil {
		t.Fatalf("AppendBER: %v", err)
	}

	// Проверяем, что данные добавлены к существующему буферу
	expected := []byte{0xAA, 0xBB, 0x2B, 0x06}

	if !bytes.Equal(result, expected) {
		t.Errorf("AppendBER = %x, want %x", result, expected)
	}
}

func TestOIDAppendBERRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			content, err := oid.AppendBER(nil)
			if err != nil {
				t.Fatalf("AppendBER: %v", err)
			}

			// Декодируем через UnmarshalBERContent
			var decoded OID
			if err := decoded.UnmarshalBERContent(content); err != nil {
				t.Fatalf("UnmarshalBERContent: %v", err)
			}

			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %x -> %v", oid, content, decoded)
			}
		})
	}
}

func TestOIDAppendBERCompareWithMarshalBER(t *testing.T) {
	oid := OID{1, 3, 6, 1}

	// AppendBER - только контент
	content, err := oid.AppendBER(nil)
	if err != nil {
		t.Fatalf("AppendBER: %v", err)
	}

	// MarshalBER - полный TLV
	fullData, err := oid.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	// Контент должен быть частью полного TLV
	if !bytes.Contains(fullData, content) {
		t.Error("AppendBER контент должен быть частью MarshalBER")
	}
}

func TestOIDAppendBERNotModifyOID(t *testing.T) {
	oid := OID{1, 3, 6, 1}
	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)

	oid.AppendBER(nil)

	if !oid.Equal(oidCopy) {
		t.Error("AppendBER() не должен изменять OID")
	}
}

// Пример использования
func ExampleOID_AppendBER() {
	oid := OID{1, 3, 6, 1}

	content, err := oid.AppendBER(nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", content)
	// Output: 2b0601
}

// Бенчмарк
func BenchmarkOIDAppendBER(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")
	dst := make([]byte, 0, 32)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		dst = dst[:0]
		_, _ = oid.AppendBER(dst)
	}
}

func TestOIDMarshalBER(t *testing.T) {
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
			name:    "Пустой OID",
			oid:     OID{},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Один компонент",
			oid:     OID{1},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Невалидный",
			oid:     OID{3, 1},
			wantErr: ErrFirstComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalBER()

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

			// Проверяем тег
			if data[0] != 0x06 {
				t.Errorf("Первый байт = 0x%02x, want 0x06", data[0])
			}
		})
	}
}

func TestOIDMarshalBERRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
		{1, 3, MaxOIDComponent},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			data, err := oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}

			var decoded OID
			if err := decoded.UnmarshalBER(data); err != nil {
				t.Fatalf("UnmarshalBER: %v", err)
			}

			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %x -> %v", oid, data, decoded)
			}
		})
	}
}

func TestOIDMarshalBERCompareWithBinary(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			berData, err := oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}

			binData, err := oid.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// Для коротких OID BER == DER
			if !bytes.Equal(berData, binData) {
				t.Errorf("BER = %x, Binary = %x", berData, binData)
			}
		})
	}
}

func TestOIDMarshalBERLongContent(t *testing.T) {
	// Создаем OID с длинным контентом (> 128 байт)
	longOID := OID{1, 3}
	for i := 0; i < 50; i++ {
		longOID = append(longOID, MaxOIDComponent)
	}

	data, err := longOID.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	if len(data) < 128 {
		t.Errorf("len = %d, want >= 128", len(data))
	}

	// Проверяем длинную форму длины
	if data[1] < 0x80 {
		t.Error("Ожидалась длинная форма длины")
	}
}

func TestOIDMarshalBERNotModifyOID(t *testing.T) {
	oid := OID{1, 3, 6, 1}
	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)

	oid.MarshalBER()

	if !oid.Equal(oidCopy) {
		t.Error("MarshalBER() не должен изменять OID")
	}
}

// Пример использования
func ExampleOID_MarshalBER() {
	oid := OID{1, 3, 6, 1}

	data, err := oid.MarshalBER()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", data)
	// Output: 06032b0601
}

// Бенчмарк
func BenchmarkOIDMarshalBER(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = oid.MarshalBER()
	}
}

func TestOIDUnmarshalBER(t *testing.T) {
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
			wantErr:  ErrInsufficientData,
		},
		{
			name:     "Один байт",
			data:     []byte{0x06},
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
		{
			name:     "Недостаточно данных",
			data:     []byte{0x06, 0x05, 0x2B},
			expected: nil,
			wantErr:  ErrInsufficientData,
		},
		{
			name:     "Лишние данные",
			data:     []byte{0x06, 0x01, 0x2B, 0x00},
			expected: nil,
			wantErr:  ErrInvalidLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.UnmarshalBER(tt.data)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UnmarshalBER: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UnmarshalBER = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalBER: %v", err)
				return
			}

			if !oid.Equal(tt.expected) {
				t.Errorf("UnmarshalBER = %v, want %v", oid, tt.expected)
			}
		})
	}
}

func TestOIDUnmarshalBERRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
		{1, 3, MaxOIDComponent},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			data, err := oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}

			var decoded OID
			if err := decoded.UnmarshalBER(data); err != nil {
				t.Fatalf("UnmarshalBER: %v", err)
			}

			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %x -> %v", oid, data, decoded)
			}
		})
	}
}

func TestOIDUnmarshalBERLongContent(t *testing.T) {
	// Создаем OID с длинным контентом
	longOID := OID{1, 3}
	for i := 0; i < 50; i++ {
		longOID = append(longOID, MaxOIDComponent)
	}

	data, err := longOID.MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}

	var decoded OID
	if err := decoded.UnmarshalBER(data); err != nil {
		t.Fatalf("UnmarshalBER: %v", err)
	}

	if !decoded.Equal(longOID) {
		t.Error("Round trip для длинного OID не удался")
	}
}

func TestOIDUnmarshalBERProperties(t *testing.T) {
	t.Run("Перезаписывает предыдущее значение", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		newData := []byte{0x06, 0x03, 0x81, 0x34, 0x03}
		if err := oid.UnmarshalBER(newData); err != nil {
			t.Fatalf("UnmarshalBER: %v", err)
		}

		if !oid.Equal(OID{2, 100, 3}) {
			t.Error("OID должен перезаписаться")
		}
	})
}

// Пример использования
func ExampleOID_UnmarshalBER() {
	data := []byte{0x06, 0x03, 0x2B, 0x06, 0x01}

	var oid OID
	if err := oid.UnmarshalBER(data); err != nil {
		panic(err)
	}

	fmt.Println(oid)
	// Output: 1.3.6.1
}

// Пример с ошибкой
func ExampleOID_UnmarshalBER_error() {
	var oid OID
	err := oid.UnmarshalBER([]byte{})
	fmt.Println(errors.Is(err, ErrInsufficientData))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDUnmarshalBER(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")
	data, _ := oid.MarshalBER()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var decoded OID
		_ = decoded.UnmarshalBER(data)
	}
}

func TestOIDUnmarshalBERContent(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected OID
		wantErr  error
	}{
		{
			name:     "Стандартный контент",
			content:  []byte{0x2B, 0x06, 0x01},
			expected: OID{1, 3, 6, 1},
			wantErr:  nil,
		},
		{
			name:     "Длинный контент",
			content:  []byte{0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00},
			expected: OID{1, 3, 6, 1, 2, 1, 1, 1, 0},
			wantErr:  nil,
		},
		{
			name:     "С первым 2",
			content:  []byte{0x81, 0x34, 0x03},
			expected: OID{2, 100, 3},
			wantErr:  nil,
		},
		{
			name:     "Пустой контент",
			content:  []byte{},
			expected: nil,
			wantErr:  ErrEmptyContent,
		},
		{
			name:     "Неверный первый компонент",
			content:  []byte{0x80},
			expected: nil,
			wantErr:  ErrFirstComponentFailed,
		},
		{
			name:     "Неверный компонент",
			content:  []byte{0x2B, 0x80},
			expected: nil,
			wantErr:  ErrComponentFailed,
		},
		{
			name:     "Переполнение",
			content:  []byte{0x2B, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F},
			expected: nil,
			wantErr:  ErrComponentFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.UnmarshalBERContent(tt.content)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UnmarshalBERContent: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UnmarshalBERContent = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalBERContent: %v", err)
				return
			}

			if !oid.Equal(tt.expected) {
				t.Errorf("UnmarshalBERContent = %v, want %v", oid, tt.expected)
			}
		})
	}
}

func TestOIDUnmarshalBERContentRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
		{1, 3, MaxOIDComponent},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			// Кодируем через AppendBER
			content, err := oid.AppendBER(nil)
			if err != nil {
				t.Fatalf("AppendBER: %v", err)
			}

			// Декодируем
			var decoded OID
			if err := decoded.UnmarshalBERContent(content); err != nil {
				t.Fatalf("UnmarshalBERContent: %v", err)
			}

			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %x -> %v", oid, content, decoded)
			}
		})
	}
}

func TestOIDUnmarshalBERContentProperties(t *testing.T) {
	t.Run("Перезаписывает предыдущее значение", func(t *testing.T) {
		oid := OID{1, 3, 6, 1}

		newContent := []byte{0x81, 0x34, 0x03}
		if err := oid.UnmarshalBERContent(newContent); err != nil {
			t.Fatalf("UnmarshalBERContent: %v", err)
		}

		if !oid.Equal(OID{2, 100, 3}) {
			t.Error("OID должен перезаписаться")
		}
	})
}

// Пример использования
func ExampleOID_UnmarshalBERContent() {
	content := []byte{0x2B, 0x06, 0x01}

	var oid OID
	if err := oid.UnmarshalBERContent(content); err != nil {
		panic(err)
	}

	fmt.Println(oid)
	// Output: 1.3.6.1
}

// Пример с ошибкой
func ExampleOID_UnmarshalBERContent_error() {
	var oid OID
	err := oid.UnmarshalBERContent([]byte{})
	fmt.Println(errors.Is(err, ErrEmptyContent))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDUnmarshalBERContent(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")
	content, _ := oid.AppendBER(nil)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var decoded OID
		_ = decoded.UnmarshalBERContent(content)
	}
}

func TestOIDMarshalDER(t *testing.T) {
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
			name:    "Пустой OID",
			oid:     OID{},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Невалидный",
			oid:     OID{3, 1},
			wantErr: ErrFirstComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.oid.MarshalDER()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("MarshalDER: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("MarshalDER = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("MarshalDER: %v", err)
				return
			}

			if len(data) == 0 {
				t.Error("MarshalDER: пустой результат")
			}
		})
	}
}

func TestOIDMarshalDEREqualsMarshalBER(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			derData, err := oid.MarshalDER()
			if err != nil {
				t.Fatalf("MarshalDER: %v", err)
			}

			berData, err := oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}

			// DER == BER для OID
			if !bytes.Equal(derData, berData) {
				t.Errorf("DER = %x, BER = %x", derData, berData)
			}
		})
	}
}

func TestOIDMarshalDEREqualsMarshalBinary(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			derData, err := oid.MarshalDER()
			if err != nil {
				t.Fatalf("MarshalDER: %v", err)
			}

			binData, err := oid.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary: %v", err)
			}

			// DER == Binary для OID
			if !bytes.Equal(derData, binData) {
				t.Errorf("DER = %x, Binary = %x", derData, binData)
			}
		})
	}
}

func TestOIDMarshalDERRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			data, err := oid.MarshalDER()
			if err != nil {
				t.Fatalf("MarshalDER: %v", err)
			}

			var decoded OID
			if err := decoded.UnmarshalDER(data); err != nil {
				t.Fatalf("UnmarshalDER: %v", err)
			}

			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %x -> %v", oid, data, decoded)
			}
		})
	}
}

// Пример использования
func ExampleOID_MarshalDER() {
	oid := OID{1, 3, 6, 1}

	data, err := oid.MarshalDER()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", data)
	// Output: 06032b0601
}

// Бенчмарк
func BenchmarkOIDMarshalDER(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = oid.MarshalDER()
	}
}

func TestOIDUnmarshalDER(t *testing.T) {
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
			wantErr:  ErrInsufficientData,
		},
		{
			name:     "Неверный тег",
			data:     []byte{0x05, 0x01, 0x2B},
			expected: nil,
			wantErr:  ErrInvalidASN1Tag,
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
			var oid OID
			err := oid.UnmarshalDER(tt.data)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("UnmarshalDER: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UnmarshalDER = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalDER: %v", err)
				return
			}

			if !oid.Equal(tt.expected) {
				t.Errorf("UnmarshalDER = %v, want %v", oid, tt.expected)
			}
		})
	}
}

func TestOIDUnmarshalDEREqualsUnmarshalBER(t *testing.T) {
	tests := [][]byte{
		{0x06, 0x03, 0x2B, 0x06, 0x01},
		{0x06, 0x08, 0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00},
		{0x06, 0x03, 0x81, 0x34, 0x03},
	}

	for _, data := range tests {
		t.Run(fmt.Sprintf("%x", data), func(t *testing.T) {
			var derOID OID
			var berOID OID

			derErr := derOID.UnmarshalDER(data)
			berErr := berOID.UnmarshalBER(data)

			if (derErr == nil) != (berErr == nil) {
				t.Error("Ошибки должны совпадать")
			}

			if derErr == nil {
				if !derOID.Equal(berOID) {
					t.Error("DER и BER должны давать одинаковый результат")
				}
			}
		})
	}
}

func TestOIDUnmarshalDEREqualsUnmarshalBinary(t *testing.T) {
	tests := [][]byte{
		{0x06, 0x03, 0x2B, 0x06, 0x01},
		{0x06, 0x03, 0x81, 0x34, 0x03},
	}

	for _, data := range tests {
		t.Run(fmt.Sprintf("%x", data), func(t *testing.T) {
			var derOID OID
			var binOID OID

			derErr := derOID.UnmarshalDER(data)
			binErr := binOID.UnmarshalBinary(data)

			if (derErr == nil) != (binErr == nil) {
				t.Error("Ошибки должны совпадать")
			}

			if derErr == nil {
				if !derOID.Equal(binOID) {
					t.Error("DER и Binary должны давать одинаковый результат")
				}
			}
		})
	}
}

func TestOIDUnmarshalDERRoundTrip(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			data, err := oid.MarshalDER()
			if err != nil {
				t.Fatalf("MarshalDER: %v", err)
			}

			var decoded OID
			if err := decoded.UnmarshalDER(data); err != nil {
				t.Fatalf("UnmarshalDER: %v", err)
			}

			if !decoded.Equal(oid) {
				t.Errorf("Round trip: %v -> %x -> %v", oid, data, decoded)
			}
		})
	}
}

// Пример использования
func ExampleOID_UnmarshalDER() {
	data := []byte{0x06, 0x03, 0x2B, 0x06, 0x01}

	var oid OID
	if err := oid.UnmarshalDER(data); err != nil {
		panic(err)
	}

	fmt.Println(oid)
	// Output: 1.3.6.1
}

// Пример с ошибкой
func ExampleOID_UnmarshalDER_error() {
	var oid OID
	err := oid.UnmarshalDER([]byte{})
	fmt.Println(errors.Is(err, ErrInsufficientData))
	// Output: true
}

// Бенчмарк
func BenchmarkOIDUnmarshalDER(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")
	data, _ := oid.MarshalDER()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var decoded OID
		_ = decoded.UnmarshalDER(data)
	}
}

func TestOIDSizeBER(t *testing.T) {
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
			name:    "Пустой OID",
			oid:     OID{},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Один компонент",
			oid:     OID{1},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Невалидный",
			oid:     OID{3, 1},
			wantErr: ErrFirstComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := tt.oid.SizeBER()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("SizeBER: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("SizeBER = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("SizeBER: %v", err)
				return
			}

			if size <= 0 {
				t.Errorf("SizeBER = %d, want > 0", size)
			}
		})
	}
}

func TestOIDSizeBERMatchesMarshalBER(t *testing.T) {
	tests := []OID{
		{1, 3, 6, 1},
		{1, 3, 6, 1, 2, 1, 1, 1, 0},
		{2, 100, 3},
		{0, 39, 1},
		{1, 3, MaxOIDComponent},
	}

	for _, oid := range tests {
		t.Run(oid.String(), func(t *testing.T) {
			size, err := oid.SizeBER()
			if err != nil {
				t.Fatalf("SizeBER: %v", err)
			}

			data, err := oid.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}

			if size != len(data) {
				t.Errorf("SizeBER = %d, len(MarshalBER) = %d", size, len(data))
			}
		})
	}
}

func TestOIDSizeBERLongContent(t *testing.T) {
	// Создаем OID с длинным контентом
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
		t.Errorf("SizeBER = %d, len(MarshalBER) = %d", size, len(data))
	}

	if size < 128 {
		t.Errorf("SizeBER = %d, want >= 128", size)
	}
}

func TestOIDSizeBERNotModifyOID(t *testing.T) {
	oid := OID{1, 3, 6, 1}
	oidCopy := make(OID, len(oid))
	copy(oidCopy, oid)

	oid.SizeBER()

	if !oid.Equal(oidCopy) {
		t.Error("SizeBER() не должен изменять OID")
	}
}

func TestOIDSizeBERNoAllocations(t *testing.T) {
	oid := OID{1, 3, 6, 1}

	allocs := testing.AllocsPerRun(1000, func() {
		_, _ = oid.SizeBER()
	})

	if allocs != 0 {
		t.Errorf("SizeBER: %f allocs, want 0", allocs)
	}
}

// Пример использования
func ExampleOID_SizeBER() {
	oid := OID{1, 3, 6, 1}

	size, err := oid.SizeBER()
	if err != nil {
		panic(err)
	}

	fmt.Println(size)
	// Output: 5
}

// Бенчмарк
func BenchmarkOIDSizeBER(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = oid.SizeBER()
	}
}

func TestReadBase128FromBytes(t *testing.T) {
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
			name:     "Переполнение uint32",
			data:     []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F},
			expected: 0,
			bytes:    0,
		},
		{
			name:     "Незавершенная",
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
			value, bytesRead := readBase128FromBytes(tt.data)

			if bytesRead != tt.bytes {
				t.Errorf("readBase128FromBytes(%x) bytesRead = %d, want %d",
					tt.data, bytesRead, tt.bytes)
			}

			if value != tt.expected {
				t.Errorf("readBase128FromBytes(%x) value = %d, want %d",
					tt.data, value, tt.expected)
			}
		})
	}
}

func TestReadBase128FromBytesRoundTrip(t *testing.T) {
	values := []uint32{
		0, 1, 127, 128, 255, 256, 16383, 16384,
		MaxOIDComponent, ^uint32(0),
	}

	for _, value := range values {
		t.Run(fmt.Sprintf("%d", value), func(t *testing.T) {
			// Кодируем через appendBase128Value
			data := appendBase128Value(nil, value)

			// Декодируем
			decoded, bytesRead := readBase128FromBytes(data)

			if bytesRead != len(data) {
				t.Errorf("bytesRead = %d, want %d", bytesRead, len(data))
			}

			if decoded != value {
				t.Errorf("Round trip: %d -> %x -> %d", value, data, decoded)
			}
		})
	}
}

func TestReadBase128FromBytesErrors(t *testing.T) {
	t.Run("Переполнение uint32", func(t *testing.T) {
		data := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F}

		value, bytesRead := readBase128FromBytes(data)

		if bytesRead != 0 {
			t.Errorf("bytesRead = %d, want 0", bytesRead)
		}
		if value != 0 {
			t.Errorf("value = %d, want 0", value)
		}
	})

	t.Run("Незавершенная", func(t *testing.T) {
		data := []byte{0x81, 0x80}

		value, bytesRead := readBase128FromBytes(data)

		if bytesRead != 0 {
			t.Errorf("bytesRead = %d, want 0", bytesRead)
		}
		if value != 0 {
			t.Errorf("value = %d, want 0", value)
		}
	})

	t.Run("Пустая", func(t *testing.T) {
		data := []byte{}

		value, bytesRead := readBase128FromBytes(data)

		if bytesRead != 0 {
			t.Errorf("bytesRead = %d, want 0", bytesRead)
		}
		if value != 0 {
			t.Errorf("value = %d, want 0", value)
		}
	})
}

func TestReadBase128FromBytesNoAllocations(t *testing.T) {
	data := []byte{0x81, 0x7F}

	allocs := testing.AllocsPerRun(1000, func() {
		_, _ = readBase128FromBytes(data)
	})

	if allocs != 0 {
		t.Errorf("readBase128FromBytes: %f allocs, want 0", allocs)
	}
}

// Бенчмарк
func BenchmarkReadBase128FromBytes(b *testing.B) {
	data := []byte{0x81, 0x7F}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = readBase128FromBytes(data)
	}
}

func TestAppendBase128Value(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{
			name:     "0",
			value:    0,
			expected: []byte{0x00},
		},
		{
			name:     "1",
			value:    1,
			expected: []byte{0x01},
		},
		{
			name:     "127",
			value:    127,
			expected: []byte{0x7F},
		},
		{
			name:     "128",
			value:    128,
			expected: []byte{0x81, 0x00},
		},
		{
			name:     "129",
			value:    129,
			expected: []byte{0x81, 0x01},
		},
		{
			name:     "255",
			value:    255,
			expected: []byte{0x81, 0x7F},
		},
		{
			name:     "256",
			value:    256,
			expected: []byte{0x82, 0x00},
		},
		{
			name:     "16383",
			value:    16383,
			expected: []byte{0xFF, 0x7F},
		},
		{
			name:     "16384",
			value:    16384,
			expected: []byte{0x81, 0x80, 0x00},
		},
		{
			name:     "MaxOIDComponent",
			value:    MaxOIDComponent,
			expected: []byte{0xFF, 0xFF, 0xFF, 0x7F},
		},
		{
			name:     "MaxUint32",
			value:    ^uint32(0),
			expected: []byte{0x8F, 0xFF, 0xFF, 0xFF, 0x7F},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendBase128Value(nil, tt.value)

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("appendBase128Value(%d) = %x, want %x",
					tt.value, result, tt.expected)
			}
		})
	}
}

func TestAppendBase128ValueAppendToExisting(t *testing.T) {
	dst := []byte{0xAA, 0xBB}

	result := appendBase128Value(dst, 128)

	expected := []byte{0xAA, 0xBB, 0x81, 0x00}

	if !bytes.Equal(result, expected) {
		t.Errorf("appendBase128Value = %x, want %x", result, expected)
	}
}

func TestAppendBase128ValueRoundTrip(t *testing.T) {
	values := []uint32{
		0, 1, 127, 128, 255, 256, 16383, 16384,
		MaxOIDComponent, ^uint32(0),
	}

	for _, value := range values {
		t.Run(fmt.Sprintf("%d", value), func(t *testing.T) {
			// Кодируем
			data := appendBase128Value(nil, value)

			// Декодируем
			decoded, bytesRead := readBase128FromBytes(data)

			if bytesRead != len(data) {
				t.Errorf("bytesRead = %d, want %d", bytesRead, len(data))
			}

			if decoded != value {
				t.Errorf("Round trip: %d -> %x -> %d", value, data, decoded)
			}
		})
	}
}

func TestAppendBase128ValueConsistency(t *testing.T) {
	t.Run("Соответствует base128Size", func(t *testing.T) {
		values := []uint32{
			0, 1, 127, 128, 16383, 16384,
			MaxOIDComponent, ^uint32(0),
		}

		for _, value := range values {
			size := base128Size(value)
			data := appendBase128Value(nil, value)

			if size != len(data) {
				t.Errorf("base128Size(%d) = %d, len = %d",
					value, size, len(data))
			}
		}
	})

	t.Run("Детерминированность", func(t *testing.T) {
		value := uint32(12345)

		result1 := appendBase128Value(nil, value)
		result2 := appendBase128Value(nil, value)

		if !bytes.Equal(result1, result2) {
			t.Error("appendBase128Value должен быть детерминированным")
		}
	})
}

func TestAppendBase128ValueNoAllocations(t *testing.T) {
	dst := make([]byte, 0, 10)

	allocs := testing.AllocsPerRun(1000, func() {
		_ = appendBase128Value(dst, 12345)
	})

	// append может аллоцировать, если dst заполнен
	if allocs > 1 {
		t.Errorf("appendBase128Value: %f allocs", allocs)
	}
}

// Бенчмарк
func BenchmarkAppendBase128Value(b *testing.B) {
	dst := make([]byte, 0, 10)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		dst = dst[:0]
		_ = appendBase128Value(dst, 12345)
	}
}

// Бенчмарк для разных значений
func BenchmarkAppendBase128ValueValues(b *testing.B) {
	values := []uint32{0, 127, 128, 16383, MaxOIDComponent, ^uint32(0)}

	for _, value := range values {
		b.Run(fmt.Sprintf("%d", value), func(b *testing.B) {
			dst := make([]byte, 0, 10)

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				dst = dst[:0]
				_ = appendBase128Value(dst, value)
			}
		})
	}
}
