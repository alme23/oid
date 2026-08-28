// oid/database_test.go
package oid

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
)

// ============================================
// ТЕСТЫ
// ============================================

func TestOIDValue(t *testing.T) {
	tests := []struct {
		name    string
		oid     OID
		wantVal driver.Value
		wantErr bool
	}{
		{
			name:    "Корректный OID",
			oid:     MustParseOID("1.3.6.1.4.1"),
			wantVal: "1.3.6.1.4.1",
			wantErr: false,
		},
		{
			name:    "Пустой OID",
			oid:     OID{},
			wantVal: nil,
			wantErr: false,
		},
		{
			name:    "Nil OID",
			oid:     nil,
			wantVal: nil,
			wantErr: false,
		},
		{
			name:    "Невалидный OID",
			oid:     OID{3, 1},
			wantVal: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.oid.Value()

			if tt.wantErr {
				if err == nil {
					t.Error("Value(): ожидалась ошибка")
				}
				return
			}

			if err != nil {
				t.Errorf("Value(): неожиданная ошибка: %v", err)
			}

			if val != tt.wantVal {
				t.Errorf("Value() = %v, ожидалось %v", val, tt.wantVal)
			}
		})
	}
}

func TestOIDScan(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantOID OID
		wantErr bool
	}{
		{
			name:    "Строка",
			input:   "1.3.6.1.4.1",
			wantOID: MustParseOID("1.3.6.1.4.1"),
			wantErr: false,
		},
		{
			name:    "Байты",
			input:   []byte("1.3.6.1.4.1"),
			wantOID: MustParseOID("1.3.6.1.4.1"),
			wantErr: false,
		},
		{
			name:    "NULL",
			input:   nil,
			wantOID: nil,
			wantErr: false,
		},
		{
			name:    "Некорректная строка",
			input:   "invalid",
			wantOID: nil,
			wantErr: true,
		},
		{
			name:    "Неподдерживаемый тип",
			input:   123,
			wantOID: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OID
			err := oid.Scan(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Scan(): ожидалась ошибка")
				}
				return
			}

			if err != nil {
				t.Errorf("Scan(): неожиданная ошибка: %v", err)
			}

			if !oid.Equal(tt.wantOID) {
				t.Errorf("Scan() = %v, ожидалось %v", oid, tt.wantOID)
			}
		})
	}
}

func TestNullOID(t *testing.T) {
	// Тест Value
	nullOID := NullOID{Valid: false}
	val, err := nullOID.Value()
	if err != nil {
		t.Errorf("NullOID.Value(): ошибка: %v", err)
	}
	if val != nil {
		t.Errorf("NullOID.Value() = %v, ожидалось nil", val)
	}

	validOID := NullOID{
		OID:   MustParseOID("1.3.6.1.4.1"),
		Valid: true,
	}
	val, err = validOID.Value()
	if err != nil {
		t.Errorf("Valid NullOID.Value(): ошибка: %v", err)
	}
	if val != "1.3.6.1.4.1" {
		t.Errorf("Valid NullOID.Value() = %v, ожидалось '1.3.6.1.4.1'", val)
	}

	// Тест Scan
	var scanned NullOID
	err = scanned.Scan("1.3.6.1.4.1")
	if err != nil {
		t.Errorf("NullOID.Scan(): ошибка: %v", err)
	}
	if !scanned.Valid {
		t.Error("NullOID.Scan(): Valid должно быть true")
	}
	if !scanned.OID.Equal(MustParseOID("1.3.6.1.4.1")) {
		t.Errorf("NullOID.Scan() = %v, ожидалось 1.3.6.1.4.1", scanned.OID)
	}

	// Тест Scan NULL
	err = scanned.Scan(nil)
	if err != nil {
		t.Errorf("NullOID.Scan(NULL): ошибка: %v", err)
	}
	if scanned.Valid {
		t.Error("NullOID.Scan(NULL): Valid должно быть false")
	}
}

func TestOIDArray(t *testing.T) {
	// Тест Value
	arr := OIDArray{
		MustParseOID("1.3.6.1.4.1"),
		MustParseOID("2.100.3"),
	}

	val, err := arr.Value()
	if err != nil {
		t.Errorf("OIDArray.Value(): ошибка: %v", err)
	}
	if val != "{1.3.6.1.4.1,2.100.3}" {
		t.Errorf("OIDArray.Value() = %v, ожидалось '{1.3.6.1.4.1,2.100.3}'", val)
	}

	// Тест Scan
	var scanned OIDArray
	err = scanned.Scan("{1.3.6.1.4.1,2.100.3}")
	if err != nil {
		t.Errorf("OIDArray.Scan(): ошибка: %v", err)
	}
	if len(scanned) != 2 {
		t.Errorf("OIDArray.Scan(): длина = %d, ожидалось 2", len(scanned))
	}
	if !scanned[0].Equal(MustParseOID("1.3.6.1.4.1")) {
		t.Errorf("OIDArray.Scan()[0] = %v, ожидалось 1.3.6.1.4.1", scanned[0])
	}

	// Тест пустого массива
	var emptyArr OIDArray
	err = emptyArr.Scan("{}")
	if err != nil {
		t.Errorf("OIDArray.Scan({}): ошибка: %v", err)
	}
	if len(emptyArr) != 0 {
		t.Errorf("OIDArray.Scan({}): длина = %d, ожидалось 0", len(emptyArr))
	}
}

func TestNullOIDJSON(t *testing.T) {
	// Тест Marshal
	nullOID := NullOID{Valid: false}
	data, err := json.Marshal(nullOID)
	if err != nil {
		t.Errorf("NullOID.MarshalJSON(): ошибка: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("NullOID.MarshalJSON() = %s, ожидалось 'null'", data)
	}

	validOID := NullOID{
		OID:   MustParseOID("1.3.6.1.4.1"),
		Valid: true,
	}
	data, err = json.Marshal(validOID)
	if err != nil {
		t.Errorf("Valid NullOID.MarshalJSON(): ошибка: %v", err)
	}
	if string(data) != `"1.3.6.1.4.1"` {
		t.Errorf("Valid NullOID.MarshalJSON() = %s, ожидалось '\"1.3.6.1.4.1\"'", data)
	}

	// Тест Unmarshal
	var unmarshaled NullOID
	err = json.Unmarshal([]byte("null"), &unmarshaled)
	if err != nil {
		t.Errorf("NullOID.UnmarshalJSON(null): ошибка: %v", err)
	}
	if unmarshaled.Valid {
		t.Error("NullOID.UnmarshalJSON(null): Valid должно быть false")
	}

	err = json.Unmarshal([]byte(`"1.3.6.1.4.1"`), &unmarshaled)
	if err != nil {
		t.Errorf("NullOID.UnmarshalJSON(): ошибка: %v", err)
	}
	if !unmarshaled.Valid {
		t.Error("NullOID.UnmarshalJSON(): Valid должно быть true")
	}
}

func TestNullOIDExtended(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		nullOID := NullOID{Valid: false}
		if nullOID.String() != "" {
			t.Errorf("NullOID.String() = %q, ожидалась пустая строка", nullOID.String())
		}

		validOID := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
		if validOID.String() != "1.3.6.1" {
			t.Errorf("Valid NullOID.String() = %q, ожидалось '1.3.6.1'", validOID.String())
		}
	})

	t.Run("Equal", func(t *testing.T) {
		null1 := NullOID{Valid: false}
		null2 := NullOID{Valid: false}
		if !null1.Equal(null2) {
			t.Error("Два NULL должны быть равны")
		}

		valid1 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
		valid2 := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
		if !valid1.Equal(valid2) {
			t.Error("Два одинаковых OID должны быть равны")
		}

		if null1.Equal(valid1) {
			t.Error("NULL и OID не должны быть равны")
		}
	})

	t.Run("FromOID", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1")
		n := FromOID(oid)

		if !n.Valid {
			t.Error("FromOID: Valid должно быть true")
		}
		if !n.OID.Equal(oid) {
			t.Error("FromOID: OID не совпадает")
		}

		emptyOID := OID{}
		n = FromOID(emptyOID)
		if n.Valid {
			t.Error("FromOID(empty): Valid должно быть false")
		}
	})

	t.Run("FromString", func(t *testing.T) {
		n, err := FromString("1.3.6.1")
		if err != nil {
			t.Fatalf("FromString: ошибка: %v", err)
		}
		if !n.Valid {
			t.Error("FromString: Valid должно быть true")
		}

		n, err = FromString("")
		if err != nil {
			t.Fatalf("FromString(empty): ошибка: %v", err)
		}
		if n.Valid {
			t.Error("FromString(empty): Valid должно быть false")
		}

		_, err = FromString("invalid")
		if err == nil {
			t.Error("FromString(invalid): ожидалась ошибка")
		}
	})

	t.Run("JSON", func(t *testing.T) {
		nullOID := NullOID{Valid: false}
		data, err := json.Marshal(nullOID)
		if err != nil {
			t.Fatalf("Marshal NULL: ошибка: %v", err)
		}
		if string(data) != "null" {
			t.Errorf("Marshal NULL = %s, ожидалось 'null'", data)
		}

		validOID := NullOID{OID: MustParseOID("1.3.6.1"), Valid: true}
		data, err = json.Marshal(validOID)
		if err != nil {
			t.Fatalf("Marshal OID: ошибка: %v", err)
		}
		if string(data) != `"1.3.6.1"` {
			t.Errorf("Marshal OID = %s, ожидалось '\"1.3.6.1\"'", data)
		}

		var unmarshaled NullOID
		err = json.Unmarshal([]byte("null"), &unmarshaled)
		if err != nil {
			t.Fatalf("Unmarshal null: ошибка: %v", err)
		}
		if unmarshaled.Valid {
			t.Error("Unmarshal null: Valid должно быть false")
		}

		err = json.Unmarshal([]byte(`"1.3.6.1"`), &unmarshaled)
		if err != nil {
			t.Fatalf("Unmarshal OID: ошибка: %v", err)
		}
		if !unmarshaled.Valid {
			t.Error("Unmarshal OID: Valid должно быть true")
		}
	})
}

func TestOIDArrayExtended(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		arr := OIDArray{
			MustParseOID("1.3.6.1"),
			MustParseOID("2.100.3"),
		}

		data, err := json.Marshal(arr)
		if err != nil {
			t.Fatalf("Marshal: ошибка: %v", err)
		}
		expected := `["1.3.6.1","2.100.3"]`
		if string(data) != expected {
			t.Errorf("Marshal = %s, ожидалось %s", data, expected)
		}

		var unmarshaled OIDArray
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("Unmarshal: ошибка: %v", err)
		}
		if !arr.Equal(unmarshaled) {
			t.Errorf("Unmarshal = %v, ожидалось %v", unmarshaled, arr)
		}
	})

	t.Run("Contains", func(t *testing.T) {
		arr := OIDArray{
			MustParseOID("1.3.6.1"),
			MustParseOID("2.100.3"),
		}

		if !arr.Contains(MustParseOID("1.3.6.1")) {
			t.Error("Contains: OID должен быть найден")
		}

		if arr.Contains(MustParseOID("1.3.6.2")) {
			t.Error("Contains: OID не должен быть найден")
		}
	})

	t.Run("Append", func(t *testing.T) {
		arr := OIDArray{MustParseOID("1.3.6.1")}
		extended := arr.Append(
			MustParseOID("2.100.3"),
			MustParseOID("0.39.1"),
		)

		if len(extended) != 3 {
			t.Errorf("Append: длина = %d, ожидалось 3", len(extended))
		}
		if len(arr) != 1 {
			t.Error("Append: оригинальный массив не должен измениться")
		}
	})

	t.Run("String", func(t *testing.T) {
		arr := OIDArray{
			MustParseOID("1.3.6.1"),
			MustParseOID("2.100.3"),
		}

		expected := "[1.3.6.1, 2.100.3]"
		if arr.String() != expected {
			t.Errorf("String = %q, ожидалось %q", arr.String(), expected)
		}

		empty := OIDArray{}
		if empty.String() != "[]" {
			t.Errorf("Empty String = %q, ожидалось '[]'", empty.String())
		}
	})
}

func TestSplitPostgresArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Простые элементы",
			input:    "1.3.6.1,2.100.3",
			expected: []string{"1.3.6.1", "2.100.3"},
		},
		{
			name:     "С кавычками",
			input:    `"1.3.6.1","2.100.3"`,
			expected: []string{"1.3.6.1", "2.100.3"},
		},
		{
			name:     "Один элемент",
			input:    "1.3.6.1",
			expected: []string{"1.3.6.1"},
		},
		{
			name:     "Пустой",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitPostgresArray(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("splitPostgresArray(%q) = %v, ожидалось %v",
					tt.input, result, tt.expected)
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitPostgresArray(%q)[%d] = %q, ожидалось %q",
						tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// ============================================
// БЕНЧМАРКИ
// ============================================

func BenchmarkOIDValue(b *testing.B) {
	oid := MustParseOID("1.3.6.1.4.1.99999.1.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := oid.Value()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOIDScan(b *testing.B) {
	oidStr := "1.3.6.1.4.1.99999.1.1"

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var oid OID
		if err := oid.Scan(oidStr); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNullOIDValue(b *testing.B) {
	nullOID := NullOID{
		OID:   MustParseOID("1.3.6.1.4.1.99999.1.1"),
		Valid: true,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := nullOID.Value()
		if err != nil {
			b.Fatal(err)
		}
	}
}
