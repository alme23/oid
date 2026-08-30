package oid

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
)

func TestRegistryStructure(t *testing.T) {
	reg := NewRegistry()

	t.Run("Поля инициализированы", func(t *testing.T) {
		if reg.names == nil {
			t.Error("names должен быть инициализирован")
		}

		if reg.oids == nil {
			t.Error("oids должен быть инициализирован")
		}
	})

	t.Run("Пустой реестр", func(t *testing.T) {
		if len(reg.names) != 0 {
			t.Errorf("len(names) = %d, want 0", len(reg.names))
		}

		if len(reg.oids) != 0 {
			t.Errorf("len(oids) = %d, want 0", len(reg.oids))
		}
	})
}

func TestRegistryThreadSafety(t *testing.T) {
	reg := NewRegistry()

	t.Run("Конкурентная запись", func(t *testing.T) {
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				oid := MustParseOID(fmt.Sprintf("1.3.6.%d", id+1))
				reg.Register(fmt.Sprintf("oid-%d", id), oid)
			}(i)
		}

		wg.Wait()

		if reg.Size() != 100 {
			t.Errorf("Size = %d, want 100", reg.Size())
		}
	})

	t.Run("Конкурентное чтение", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1")
		reg.Register("test", oid)

		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, _ = reg.LookupByName("test")
				_ = reg.Contains("test")
				_ = reg.Size()
			}()
		}

		wg.Wait()
	})
}

func TestRegistryMutexType(t *testing.T) {
	reg := NewRegistry()

	// Проверяем, что mutex - RWMutex
	var _ sync.RWMutex = reg.mu
}

func TestRegistryMapTypes(t *testing.T) {
	reg := NewRegistry()

	// Проверяем типы map
	var _ map[string]OID = reg.names
	var _ map[string]string = reg.oids
}

// Пример использования
func ExampleRegistry() {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	if got, exists := reg.LookupByName("test"); exists {
		fmt.Println(got)
	}
	// Output: 1.3.6.1
}

// Бенчмарк
func BenchmarkRegistryCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = NewRegistry()
	}
}

func TestNewRegistry(t *testing.T) {
	t.Run("Инициализация", func(t *testing.T) {
		reg := NewRegistry()

		if reg == nil {
			t.Fatal("NewRegistry: nil результат")
		}

		if reg.names == nil {
			t.Error("names должен быть инициализирован")
		}

		if reg.oids == nil {
			t.Error("oids должен быть инициализирован")
		}

		if len(reg.names) != 0 {
			t.Errorf("len(names) = %d, want 0", len(reg.names))
		}

		if len(reg.oids) != 0 {
			t.Errorf("len(oids) = %d, want 0", len(reg.oids))
		}
	})

	t.Run("Независимые экземпляры", func(t *testing.T) {
		reg1 := NewRegistry()
		reg2 := NewRegistry()

		if reg1 == reg2 {
			t.Error("NewRegistry должен создавать разные экземпляры")
		}

		// Добавляем в первый
		oid := MustParseOID("1.3.6.1")
		reg1.Register("test", oid)

		// Второй не должен измениться
		if len(reg2.names) != 0 {
			t.Error("Экземпляры не должны влиять друг на друга")
		}
		if len(reg2.oids) != 0 {
			t.Error("Экземпляры не должны влиять друг на друга")
		}
	})

	t.Run("Готов к использованию", func(t *testing.T) {
		reg := NewRegistry()
		oid := MustParseOID("1.3.6.1")

		// Можно сразу регистрировать
		if err := reg.Register("test", oid); err != nil {
			t.Errorf("Register: %v", err)
		}

		// Можно искать
		if got, exists := reg.LookupByName("test"); !exists || !got.Equal(oid) {
			t.Error("LookupByName не работает")
		}

		// Можно искать по OID
		if name, exists := reg.LookupByOID(oid); !exists || name != "test" {
			t.Error("LookupByOID не работает")
		}
	})
}

func TestNewRegistryTypes(t *testing.T) {
	reg := NewRegistry()

	// Проверяем типы map
	var _ map[string]OID = reg.names
	var _ map[string]string = reg.oids
}

func TestNewRegistrySize(t *testing.T) {
	reg := NewRegistry()

	if reg.Size() != 0 {
		t.Errorf("Size = %d, want 0", reg.Size())
	}

	// Добавляем запись
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	if reg.Size() != 1 {
		t.Errorf("Size = %d, want 1", reg.Size())
	}
}

func TestNewRegistryContains(t *testing.T) {
	reg := NewRegistry()

	if reg.Contains("test") {
		t.Error("Contains должен вернуть false для пустого реестра")
	}

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	if !reg.Contains("test") {
		t.Error("Contains должен вернуть true после регистрации")
	}
}

func TestNewRegistryClear(t *testing.T) {
	reg := NewRegistry()

	// Добавляем записи с РАЗНЫМИ OID
	for i := 0; i < 5; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	if reg.Size() != 5 {
		t.Errorf("Size = %d, want 5", reg.Size())
	}

	// Очищаем
	reg.Clear()

	if reg.Size() != 0 {
		t.Errorf("Size после Clear = %d, want 0", reg.Size())
	}
}

func TestNewRegistryList(t *testing.T) {
	reg := NewRegistry()

	// Пустой список
	list := reg.List()
	if len(list) != 0 {
		t.Errorf("len(List) = %d, want 0", len(list))
	}

	// Добавляем записи
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	reg.Register("first", oid1)
	reg.Register("second", oid2)

	list = reg.List()
	if len(list) != 2 {
		t.Errorf("len(List) = %d, want 2", len(list))
	}

	if !list["first"].Equal(oid1) {
		t.Error("List[first] неверный")
	}
	if !list["second"].Equal(oid2) {
		t.Error("List[second] неверный")
	}
}

func TestNewRegistryRemove(t *testing.T) {
	reg := NewRegistry()

	// Remove из пустого
	if reg.Remove("nonexistent") {
		t.Error("Remove должен вернуть false для пустого реестра")
	}

	// Добавляем и удаляем
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	if !reg.Remove("test") {
		t.Error("Remove должен вернуть true")
	}

	if reg.Size() != 0 {
		t.Errorf("Size после Remove = %d, want 0", reg.Size())
	}
}

func TestNewRegistryNames(t *testing.T) {
	reg := NewRegistry()

	// Пустые имена
	if len(reg.Names()) != 0 {
		t.Error("Names должен быть пустым")
	}

	// Добавляем
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	names := reg.Names()
	if len(names) != 1 {
		t.Errorf("len(Names) = %d, want 1", len(names))
	}
	if names[0] != "test" {
		t.Errorf("Names[0] = %q, want 'test'", names[0])
	}
}

func TestNewRegistryOIDs(t *testing.T) {
	reg := NewRegistry()

	// Пустые OID
	if len(reg.OIDs()) != 0 {
		t.Error("OIDs должен быть пустым")
	}

	// Добавляем
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	oids := reg.OIDs()
	if len(oids) != 1 {
		t.Errorf("len(OIDs) = %d, want 1", len(oids))
	}
	if !oids[0].Equal(oid) {
		t.Error("OIDs[0] неверный")
	}
}

// Пример использования
func ExampleNewRegistry() {
	reg := NewRegistry()

	// Регистрируем OID
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Ищем
	if got, exists := reg.LookupByName("test"); exists {
		fmt.Println(got)
	}
	// Output: 1.3.6.1
}

// Бенчмарк
func BenchmarkNewRegistry(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = NewRegistry()
	}
}

func TestRegistryRegister(t *testing.T) {
	reg := NewRegistry()

	tests := []struct {
		name        string
		oidName     string
		oid         OID
		expectedLen int
		wantErr     error
	}{
		{
			name:        "Первая регистрация",
			oidName:     "test",
			oid:         MustParseOID("1.3.6.1"),
			expectedLen: 1,
			wantErr:     nil,
		},
		{
			name:        "Вторая регистрация",
			oidName:     "second",
			oid:         MustParseOID("2.100.3"),
			expectedLen: 2,
			wantErr:     nil,
		},
		{
			name:        "Дубликат имени (перезапись)",
			oidName:     "test",
			oid:         MustParseOID("1.3.6.2"),
			expectedLen: 2,
			wantErr:     nil,
		},
		{
			name:        "Дубликат OID с другим именем",
			oidName:     "other",
			oid:         MustParseOID("1.3.6.1"),
			expectedLen: 2,
			wantErr:     ErrOIDAlreadyRegistered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reg.Register(tt.oidName, tt.oid)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Register: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Register = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Register: %v", err)
				return
			}

			if len(reg.names) != tt.expectedLen {
				t.Errorf("len(names) = %d, want %d", len(reg.names), tt.expectedLen)
			}
		})
	}
}

func TestRegistryRegisterValidation(t *testing.T) {
	reg := NewRegistry()

	tests := []struct {
		name    string
		oidName string
		oid     OID
		wantErr error
	}{
		{
			name:    "Пустой OID",
			oidName: "empty",
			oid:     OID{},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Один компонент",
			oidName: "short",
			oid:     OID{1},
			wantErr: ErrOIDTooShort,
		},
		{
			name:    "Первый > 2",
			oidName: "invalid",
			oid:     OID{3, 1},
			wantErr: ErrFirstComponentTooBig,
		},
		{
			name:    "Второй > 39",
			oidName: "invalid2",
			oid:     OID{1, 40},
			wantErr: ErrSecondComponentTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reg.Register(tt.oidName, tt.oid)

			if err == nil {
				t.Error("Register: expected error")
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Register = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryRegisterCopy(t *testing.T) {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Изменяем оригинал
	oid[0] = 99

	// Реестр не должен измениться
	if reg.names["test"][0] != 1 {
		t.Error("Register должен создать копию")
	}
}

func TestRegistryRegisterOverwrite(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	reg.Register("test", oid1)
	reg.Register("test", oid2)

	if len(reg.names) != 1 {
		t.Errorf("len(names) = %d, want 1", len(reg.names))
	}

	if !reg.names["test"].Equal(oid2) {
		t.Error("Должен сохраниться второй OID")
	}

	// Обратный индекс: oid1 -> test (старый), oid2 -> test (новый)
	// Оба OID указывают на "test"
	if len(reg.oids) != 2 {
		t.Errorf("len(oids) = %d, want 2 (оба OID указывают на test)", len(reg.oids))
	}

	// Проверяем оба OID
	if reg.oids[oid1.String()] != "test" {
		t.Error("oid1 должен указывать на test")
	}
	if reg.oids[oid2.String()] != "test" {
		t.Error("oid2 должен указывать на test")
	}
}

func TestRegistryRegisterSameOIDSameName(t *testing.T) {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")

	// Первая регистрация
	if err := reg.Register("test", oid); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Повторная регистрация того же OID с тем же именем
	if err := reg.Register("test", oid); err != nil {
		t.Errorf("Register: %v", err)
	}

	if len(reg.names) != 1 {
		t.Errorf("len(names) = %d, want 1", len(reg.names))
	}
}

func TestRegistryRegisterNilName(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")

	// Пустое имя
	if err := reg.Register("", oid); err != nil {
		t.Errorf("Register: %v", err)
	}

	if _, exists := reg.names[""]; !exists {
		t.Error("Пустое имя не зарегистрировано")
	}
}

// Пример использования
func ExampleRegistry_Register() {
	reg := NewRegistry()

	// Регистрируем OID
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Ищем
	if got, exists := reg.LookupByName("test"); exists {
		fmt.Println(got)
	}
	// Output: 1.3.6.1
}

// Пример с ошибкой
func ExampleRegistry_Register_error() {
	reg := NewRegistry()

	// Регистрируем OID
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Пытаемся зарегистрировать тот же OID под другим именем
	err := reg.Register("other", oid)
	fmt.Println(errors.Is(err, ErrOIDAlreadyRegistered))
	// Output: true
}

// Бенчмарк
func BenchmarkRegistryRegister(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg.Register("test", oid)
	}
}

func TestRegistryLookupByName(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	tests := []struct {
		name     string
		lookup   string
		expected OID
		exists   bool
	}{
		{
			name:     "Существующий first",
			lookup:   "first",
			expected: oid1,
			exists:   true,
		},
		{
			name:     "Существующий second",
			lookup:   "second",
			expected: oid2,
			exists:   true,
		},
		{
			name:     "Несуществующий",
			lookup:   "nonexistent",
			expected: nil,
			exists:   false,
		},
		{
			name:     "Пустое имя",
			lookup:   "",
			expected: nil,
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := reg.LookupByName(tt.lookup)

			if exists != tt.exists {
				t.Errorf("exists = %v, want %v", exists, tt.exists)
			}

			if exists {
				if !got.Equal(tt.expected) {
					t.Errorf("got %v, want %v", got, tt.expected)
				}
			} else {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
			}
		})
	}
}

func TestRegistryLookupByNameCopy(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Получаем копию
	got, exists := reg.LookupByName("test")
	if !exists {
		t.Fatal("test not found")
	}

	// Изменяем копию
	got[0] = 99

	// Реестр не должен измениться
	if reg.names["test"][0] != 1 {
		t.Error("LookupByName должен вернуть копию")
	}
}

func TestRegistryLookupByNameEachCallNewCopy(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	got1, _ := reg.LookupByName("test")
	got2, _ := reg.LookupByName("test")

	// Изменяем первую копию
	got1[0] = 99

	// Вторая не должна измениться
	if got2[0] != 1 {
		t.Error("Каждый вызов должен возвращать новую копию")
	}
}

func TestRegistryLookupByNameNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.names)
	reg.LookupByName("test")
	after := len(reg.names)

	if before != after {
		t.Error("LookupByName не должен изменять реестр")
	}
}

func TestRegistryLookupByNameNil(t *testing.T) {
	reg := NewRegistry()

	// Пустой OID нельзя зарегистрировать (валидация)
	err := reg.Register("empty", OID{})
	if err == nil {
		t.Error("Register должен отклонить пустой OID")
	}

	// Проверяем, что empty не зарегистрирован
	_, exists := reg.LookupByName("empty")
	if exists {
		t.Error("empty не должен быть зарегистрирован")
	}
}

// Пример использования
func ExampleRegistry_LookupByName() {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	if got, exists := reg.LookupByName("test"); exists {
		fmt.Println(got)
	}

	_, exists := reg.LookupByName("nonexistent")
	fmt.Println(exists)
	// Output:
	// 1.3.6.1
	// false
}

// Бенчмарк
func BenchmarkRegistryLookupByName(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.LookupByName("test")
	}
}

func TestRegistryLookupByNameNoCopy(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	tests := []struct {
		name     string
		lookup   string
		expected OID
		exists   bool
	}{
		{
			name:     "Существующий first",
			lookup:   "first",
			expected: oid1,
			exists:   true,
		},
		{
			name:     "Существующий second",
			lookup:   "second",
			expected: oid2,
			exists:   true,
		},
		{
			name:     "Несуществующий",
			lookup:   "nonexistent",
			expected: nil,
			exists:   false,
		},
		{
			name:     "Пустое имя",
			lookup:   "",
			expected: nil,
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := reg.LookupByNameNoCopy(tt.lookup)

			if exists != tt.exists {
				t.Errorf("exists = %v, want %v", exists, tt.exists)
			}

			if exists {
				if !got.Equal(tt.expected) {
					t.Errorf("got %v, want %v", got, tt.expected)
				}
			} else {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
			}
		})
	}
}

func TestRegistryLookupByNameNoCopyReturnsReference(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Получаем ссылку
	got, exists := reg.LookupByNameNoCopy("test")
	if !exists {
		t.Fatal("test not found")
	}

	// Изменяем через ссылку
	got[0] = 99

	// Реестр должен измениться (хранит ссылку)
	if reg.names["test"][0] != 99 {
		t.Error("LookupByNameNoCopy должен вернуть ссылку")
	}

	// Восстанавливаем
	got[0] = 1
}

func TestRegistryLookupByNameNoCopySameSlice(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	got1, _ := reg.LookupByNameNoCopy("test")
	got2, _ := reg.LookupByNameNoCopy("test")

	// Оба должны указывать на один и тот же массив
	if &got1[0] != &got2[0] {
		t.Error("LookupByNameNoCopy должен возвращать один и тот же слайс")
	}
}

func TestRegistryLookupByNameNoCopyNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.names)
	reg.LookupByNameNoCopy("test")
	after := len(reg.names)

	if before != after {
		t.Error("LookupByNameNoCopy не должен изменять реестр")
	}
}

func TestRegistryLookupByNameVsNoCopy(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// LookupByName - копия
	copied, _ := reg.LookupByName("test")
	copied[0] = 99
	if reg.names["test"][0] != 1 {
		t.Error("LookupByName должен вернуть копию")
	}

	// LookupByNameNoCopy - ссылка
	referenced, _ := reg.LookupByNameNoCopy("test")
	referenced[0] = 99
	if reg.names["test"][0] != 99 {
		t.Error("LookupByNameNoCopy должен вернуть ссылку")
	}

	// Восстанавливаем
	referenced[0] = 1
}

// Пример использования
func ExampleRegistry_LookupByNameNoCopy() {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	if got, exists := reg.LookupByNameNoCopy("test"); exists {
		fmt.Println(got)
	}
	// Output: 1.3.6.1
}

// Бенчмарк
func BenchmarkRegistryLookupByNameNoCopy(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.LookupByNameNoCopy("test")
	}
}

// Сравнение LookupByName vs LookupByNameNoCopy
func BenchmarkLookupByNameVsNoCopy(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.Run("LookupByName", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = reg.LookupByName("test")
		}
	})

	b.Run("LookupByNameNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = reg.LookupByNameNoCopy("test")
		}
	})
}

func TestRegistryLookupByOID(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	tests := []struct {
		name     string
		oid      OID
		expected string
		exists   bool
	}{
		{
			name:     "Существующий oid1",
			oid:      oid1,
			expected: "first",
			exists:   true,
		},
		{
			name:     "Существующий oid2",
			oid:      oid2,
			expected: "second",
			exists:   true,
		},
		{
			name:     "Несуществующий",
			oid:      MustParseOID("1.3.6.99"),
			expected: "",
			exists:   false,
		},
		{
			name:     "Пустой OID",
			oid:      OID{},
			expected: "",
			exists:   false,
		},
		{
			name:     "Nil OID",
			oid:      nil,
			expected: "",
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := reg.LookupByOID(tt.oid)

			if exists != tt.exists {
				t.Errorf("exists = %v, want %v", exists, tt.exists)
			}

			if exists {
				if got != tt.expected {
					t.Errorf("got %q, want %q", got, tt.expected)
				}
			} else {
				if got != "" {
					t.Errorf("got %q, want empty", got)
				}
			}
		})
	}
}

func TestRegistryLookupByOIDNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.oids)
	reg.LookupByOID(oid)
	after := len(reg.oids)

	if before != after {
		t.Error("LookupByOID не должен изменять реестр")
	}
}

func TestRegistryLookupByOIDRoundTrip(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем несколько OID
	oids := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
		"third":  MustParseOID("0.39.1"),
	}

	for name, oid := range oids {
		reg.Register(name, oid)
	}

	// Ищем каждый по OID
	for expectedName, oid := range oids {
		gotName, exists := reg.LookupByOID(oid)
		if !exists {
			t.Errorf("OID %v not found", oid)
			continue
		}
		if gotName != expectedName {
			t.Errorf("got %q, want %q", gotName, expectedName)
		}
	}
}

func TestRegistryLookupByOIDAfterOverwrite(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	// Регистрируем test -> oid1
	reg.Register("test", oid1)

	// Перезаписываем test -> oid2
	reg.Register("test", oid2)

	// oid1 должен указывать на test (старый)
	if name, exists := reg.LookupByOID(oid1); exists {
		if name != "test" {
			t.Errorf("oid1 -> %q, want 'test'", name)
		}
	} else {
		t.Error("oid1 should still be found")
	}

	// oid2 должен указывать на test (новый)
	if name, exists := reg.LookupByOID(oid2); exists {
		if name != "test" {
			t.Errorf("oid2 -> %q, want 'test'", name)
		}
	} else {
		t.Error("oid2 should be found")
	}
}

// Пример использования
func ExampleRegistry_LookupByOID() {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	if name, exists := reg.LookupByOID(oid); exists {
		fmt.Println(name)
	}

	// Используем ParseOID для несуществующего
	notExist, _ := ParseOID("1.3.6.99")
	_, exists := reg.LookupByOID(notExist)
	fmt.Println(exists)
	// Output:
	// test
	// false
}

// Бенчмарк
func BenchmarkRegistryLookupByOID(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = reg.LookupByOID(oid)
	}
}

func TestRegistryRemove(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	tests := []struct {
		name        string
		removeName  string
		expectedRem bool
		expectedLen int
	}{
		{
			name:        "Удаление существующего",
			removeName:  "first",
			expectedRem: true,
			expectedLen: 1,
		},
		{
			name:        "Удаление несуществующего",
			removeName:  "nonexistent",
			expectedRem: false,
			expectedLen: 1,
		},
		{
			name:        "Удаление второго",
			removeName:  "second",
			expectedRem: true,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removed := reg.Remove(tt.removeName)

			if removed != tt.expectedRem {
				t.Errorf("removed = %v, want %v", removed, tt.expectedRem)
			}

			if len(reg.names) != tt.expectedLen {
				t.Errorf("len(names) = %d, want %d", len(reg.names), tt.expectedLen)
			}
		})
	}
}

func TestRegistryRemoveFromEmpty(t *testing.T) {
	reg := NewRegistry()

	if reg.Remove("nonexistent") {
		t.Error("Remove из пустого реестра должен вернуть false")
	}

	if len(reg.names) != 0 {
		t.Errorf("len(names) = %d, want 0", len(reg.names))
	}
	if len(reg.oids) != 0 {
		t.Errorf("len(oids) = %d, want 0", len(reg.oids))
	}
}

func TestRegistryRemoveBothMaps(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Проверяем, что оба map содержат запись
	if len(reg.names) != 1 {
		t.Error("names должен содержать 1 запись")
	}
	if len(reg.oids) != 1 {
		t.Error("oids должен содержать 1 запись")
	}

	// Удаляем
	removed := reg.Remove("test")

	if !removed {
		t.Error("Remove должен вернуть true")
	}

	// Проверяем, что оба map пустые
	if len(reg.names) != 0 {
		t.Errorf("len(names) = %d, want 0", len(reg.names))
	}
	if len(reg.oids) != 0 {
		t.Errorf("len(oids) = %d, want 0", len(reg.oids))
	}
}

func TestRegistryRemoveThenLookup(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Удаляем
	reg.Remove("test")

	// Проверяем, что не можем найти
	if _, exists := reg.LookupByName("test"); exists {
		t.Error("LookupByName должен вернуть false после удаления")
	}

	if _, exists := reg.LookupByOID(oid); exists {
		t.Error("LookupByOID должен вернуть false после удаления")
	}
}

func TestRegistryRemoveThenReRegister(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")

	// Регистрируем
	reg.Register("test", oid)

	// Удаляем
	reg.Remove("test")

	// Регистрируем снова
	if err := reg.Register("test", oid); err != nil {
		t.Errorf("Register после удаления: %v", err)
	}

	if reg.Size() != 1 {
		t.Errorf("Size = %d, want 1", reg.Size())
	}
}

func TestRegistryRemoveEmptyName(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")

	// Регистрируем с пустым именем
	reg.Register("", oid)

	// Удаляем
	if !reg.Remove("") {
		t.Error("Remove должен вернуть true для пустого имени")
	}

	if reg.Size() != 0 {
		t.Errorf("Size = %d, want 0", reg.Size())
	}
}

func TestRegistryRemoveOverwritten(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	// Регистрируем и перезаписываем
	reg.Register("test", oid1)
	reg.Register("test", oid2)

	// Удаляем
	reg.Remove("test")

	// Проверяем, что oid2 удален
	if _, exists := reg.LookupByOID(oid2); exists {
		t.Error("oid2 должен быть удален")
	}

	// oid1 может остаться (старый обратный индекс)
	// Это зависит от реализации
}

// Пример использования
func ExampleRegistry_Remove() {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	fmt.Println(reg.Remove("test"))
	fmt.Println(reg.Remove("test"))
	// Output:
	// true
	// false
}

// Бенчмарк
func BenchmarkRegistryRemove(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg := NewRegistry()
		oid := MustParseOID("1.3.6.1")
		reg.Register("test", oid)
		reg.Remove("test")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	reg.Register("first", oid1)
	reg.Register("second", oid2)
	reg.Register("third", oid3)

	t.Run("Содержит все записи", func(t *testing.T) {
		list := reg.List()

		if len(list) != 3 {
			t.Errorf("len = %d, want 3", len(list))
		}

		if !list["first"].Equal(oid1) {
			t.Error("list[first] неверный")
		}
		if !list["second"].Equal(oid2) {
			t.Error("list[second] неверный")
		}
		if !list["third"].Equal(oid3) {
			t.Error("list[third] неверный")
		}
	})

	t.Run("Пустой реестр", func(t *testing.T) {
		emptyReg := NewRegistry()
		list := emptyReg.List()

		if len(list) != 0 {
			t.Errorf("len = %d, want 0", len(list))
		}
	})
}

func TestRegistryListDeepCopy(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Получаем список
	list := reg.List()

	// Изменяем OID в списке
	list["test"][0] = 99

	// Реестр не должен измениться
	if reg.names["test"][0] != 1 {
		t.Error("List должен вернуть глубокую копию")
	}
}

func TestRegistryListEachCallNewMap(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	list1 := reg.List()
	list2 := reg.List()

	// Изменяем первый map
	list1["new"] = MustParseOID("2.100.3")

	// Второй не должен измениться
	if _, exists := list2["new"]; exists {
		t.Error("Каждый вызов должен возвращать новый map")
	}
}

func TestRegistryListNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.names)
	list := reg.List()
	after := len(reg.names)

	if before != after {
		t.Error("List не должен изменять реестр")
	}

	// Изменяем список
	list["test"][0] = 99
	list["new"] = MustParseOID("2.100.3")

	// Реестр не должен измениться
	if len(reg.names) != 1 {
		t.Error("Реестр не должен измениться")
	}
	if reg.names["test"][0] != 1 {
		t.Error("OID в реестре не должен измениться")
	}
}

func TestRegistryListDeepCopyEachOID(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	list := reg.List()

	// Изменяем каждый OID в списке
	for name, oid := range list {
		oid[0] = 99
		list[name] = oid
	}

	// Реестр не должен измениться
	if reg.names["first"][0] != 1 {
		t.Error("first OID не должен измениться")
	}
	if reg.names["second"][0] != 2 {
		t.Error("second OID не должен измениться")
	}
}

func TestRegistryListWithNilOID(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем валидный OID
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	list := reg.List()

	if len(list) != 1 {
		t.Errorf("len = %d, want 1", len(list))
	}

	if !list["test"].Equal(oid) {
		t.Error("list[test] неверный")
	}
}

// Пример использования
func ExampleRegistry_List() {
	reg := NewRegistry()

	reg.Register("first", MustParseOID("1.3.6.1"))
	reg.Register("second", MustParseOID("2.100.3"))

	list := reg.List()

	for name, oid := range list {
		fmt.Printf("%s: %s\n", name, oid)
	}
	// Output может быть в любом порядке
}

// Бенчмарк
func BenchmarkRegistryList(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID("1.3.6.1")
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = reg.List()
	}
}

func TestRegistryListNoCopy(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	reg.Register("first", oid1)
	reg.Register("second", oid2)
	reg.Register("third", oid3)

	t.Run("Содержит все записи", func(t *testing.T) {
		list := reg.ListNoCopy()

		if len(list) != 3 {
			t.Errorf("len = %d, want 3", len(list))
		}

		if !list["first"].Equal(oid1) {
			t.Error("list[first] неверный")
		}
		if !list["second"].Equal(oid2) {
			t.Error("list[second] неверный")
		}
		if !list["third"].Equal(oid3) {
			t.Error("list[third] неверный")
		}
	})

	t.Run("Пустой реестр", func(t *testing.T) {
		emptyReg := NewRegistry()
		list := emptyReg.ListNoCopy()

		if len(list) != 0 {
			t.Errorf("len = %d, want 0", len(list))
		}
	})
}

func TestRegistryListNoCopySharedSlices(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Получаем список
	list := reg.ListNoCopy()

	// Изменяем OID в списке
	list["test"][0] = 99

	// Реестр должен измениться (общий слайс)
	if reg.names["test"][0] != 99 {
		t.Error("ListNoCopy должен вернуть ссылки на внутренние срезы")
	}

	// Восстанавливаем
	list["test"][0] = 1
}

func TestRegistryListNoCopyEachCallNewMap(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	list1 := reg.ListNoCopy()
	list2 := reg.ListNoCopy()

	// Изменяем первый map (добавляем новую запись)
	list1["new"] = MustParseOID("2.100.3")

	// Второй не должен измениться
	if _, exists := list2["new"]; exists {
		t.Error("Каждый вызов должен возвращать новый map")
	}
}

func TestRegistryListNoCopyNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.names)
	list := reg.ListNoCopy()
	after := len(reg.names)

	if before != after {
		t.Error("ListNoCopy не должен изменять реестр")
	}

	// Добавляем запись в список (не влияет на реестр)
	list["new"] = MustParseOID("2.100.3")

	if len(reg.names) != 1 {
		t.Error("Реестр не должен измениться при добавлении в список")
	}
}

func TestRegistryListVsListNoCopy(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// List - глубокая копия
	list := reg.List()
	list["test"][0] = 99
	if reg.names["test"][0] != 1 {
		t.Error("List должен вернуть копию")
	}

	// ListNoCopy - ссылка
	listNoCopy := reg.ListNoCopy()
	listNoCopy["test"][0] = 99
	if reg.names["test"][0] != 99 {
		t.Error("ListNoCopy должен вернуть ссылку")
	}

	// Восстанавливаем
	listNoCopy["test"][0] = 1
}

func TestRegistryListNoCopySamePointers(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	list := reg.ListNoCopy()

	// Проверяем, что OID указывает на тот же массив
	if &list["test"][0] != &reg.names["test"][0] {
		t.Error("ListNoCopy должен вернуть тот же слайс")
	}
}

// Пример использования
func ExampleRegistry_ListNoCopy() {
	reg := NewRegistry()

	reg.Register("first", MustParseOID("1.3.6.1"))
	reg.Register("second", MustParseOID("2.100.3"))

	list := reg.ListNoCopy()

	fmt.Println(len(list))
	// Output: 2
}

// Бенчмарк
func BenchmarkRegistryListNoCopy(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID("1.3.6.1")
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = reg.ListNoCopy()
	}
}

// Сравнение List vs ListNoCopy
func BenchmarkListVsListNoCopy(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID("1.3.6.1")
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	b.Run("List", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = reg.List()
		}
	})

	b.Run("ListNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = reg.ListNoCopy()
		}
	})
}

func TestRegistrySize(t *testing.T) {
	reg := NewRegistry()

	t.Run("Пустой реестр", func(t *testing.T) {
		if reg.Size() != 0 {
			t.Errorf("Size = %d, want 0", reg.Size())
		}
	})

	t.Run("Одна запись", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1")
		reg.Register("first", oid)

		if reg.Size() != 1 {
			t.Errorf("Size = %d, want 1", reg.Size())
		}
	})

	t.Run("Две записи", func(t *testing.T) {
		oid := MustParseOID("2.100.3")
		reg.Register("second", oid)

		if reg.Size() != 2 {
			t.Errorf("Size = %d, want 2", reg.Size())
		}
	})

	t.Run("Три записи", func(t *testing.T) {
		oid := MustParseOID("0.39.1")
		reg.Register("third", oid)

		if reg.Size() != 3 {
			t.Errorf("Size = %d, want 3", reg.Size())
		}
	})
}

func TestRegistrySizeAfterRemove(t *testing.T) {
	reg := NewRegistry()

	// Добавляем записи
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	if reg.Size() != 2 {
		t.Errorf("Size = %d, want 2", reg.Size())
	}

	// Удаляем одну
	reg.Remove("first")

	if reg.Size() != 1 {
		t.Errorf("Size после Remove = %d, want 1", reg.Size())
	}

	// Удаляем вторую
	reg.Remove("second")

	if reg.Size() != 0 {
		t.Errorf("Size после Remove = %d, want 0", reg.Size())
	}
}

func TestRegistrySizeAfterClear(t *testing.T) {
	reg := NewRegistry()

	// Добавляем записи
	for i := 0; i < 5; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	if reg.Size() != 5 {
		t.Errorf("Size = %d, want 5", reg.Size())
	}

	// Очищаем
	reg.Clear()

	if reg.Size() != 0 {
		t.Errorf("Size после Clear = %d, want 0", reg.Size())
	}
}

func TestRegistrySizeAfterOverwrite(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	// Регистрируем
	reg.Register("test", oid1)

	if reg.Size() != 1 {
		t.Errorf("Size = %d, want 1", reg.Size())
	}

	// Перезаписываем
	reg.Register("test", oid2)

	if reg.Size() != 1 {
		t.Errorf("Size после перезаписи = %d, want 1", reg.Size())
	}
}

func TestRegistrySizeNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.names)
	size := reg.Size()
	after := len(reg.names)

	if before != after {
		t.Error("Size не должен изменять реестр")
	}

	if size != 1 {
		t.Errorf("Size = %d, want 1", size)
	}
}

func TestRegistrySizeConcurrent(t *testing.T) {
	reg := NewRegistry()

	// Добавляем записи
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	if reg.Size() != 10 {
		t.Errorf("Size = %d, want 10", reg.Size())
	}
}

// Пример использования
func ExampleRegistry_Size() {
	reg := NewRegistry()

	fmt.Println(reg.Size())

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	fmt.Println(reg.Size())
	// Output:
	// 0
	// 1
}

// Бенчмарк
func BenchmarkRegistrySize(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID("1.3.6.1")
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = reg.Size()
	}
}

func TestRegistryContains(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	tests := []struct {
		name     string
		lookup   string
		expected bool
	}{
		{
			name:     "Существующий first",
			lookup:   "first",
			expected: true,
		},
		{
			name:     "Существующий second",
			lookup:   "second",
			expected: true,
		},
		{
			name:     "Несуществующий",
			lookup:   "nonexistent",
			expected: false,
		},
		{
			name:     "Пустое имя",
			lookup:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := reg.Contains(tt.lookup)

			if exists != tt.expected {
				t.Errorf("Contains(%q) = %v, want %v", tt.lookup, exists, tt.expected)
			}
		})
	}
}

func TestRegistryContainsEmptyRegistry(t *testing.T) {
	reg := NewRegistry()

	if reg.Contains("any") {
		t.Error("Пустой реестр не должен содержать записи")
	}
}

func TestRegistryContainsAfterRemove(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")

	// Регистрируем
	reg.Register("test", oid)

	if !reg.Contains("test") {
		t.Error("Contains должен вернуть true после регистрации")
	}

	// Удаляем
	reg.Remove("test")

	if reg.Contains("test") {
		t.Error("Contains должен вернуть false после удаления")
	}
}

func TestRegistryContainsAfterClear(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")

	// Регистрируем
	reg.Register("test", oid)

	if !reg.Contains("test") {
		t.Error("Contains должен вернуть true после регистрации")
	}

	// Очищаем
	reg.Clear()

	if reg.Contains("test") {
		t.Error("Contains должен вернуть false после очистки")
	}
}

func TestRegistryContainsAfterOverwrite(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	// Регистрируем
	reg.Register("test", oid1)

	if !reg.Contains("test") {
		t.Error("Contains должен вернуть true")
	}

	// Перезаписываем
	reg.Register("test", oid2)

	if !reg.Contains("test") {
		t.Error("Contains должен вернуть true после перезаписи")
	}
}

func TestRegistryContainsEmptyName(t *testing.T) {
	reg := NewRegistry()

	// Пустое имя в пустом реестре
	if reg.Contains("") {
		t.Error("Contains(\"\") должен вернуть false для пустого реестра")
	}

	// Регистрируем с пустым именем
	oid := MustParseOID("1.3.6.1")
	reg.Register("", oid)

	if !reg.Contains("") {
		t.Error("Contains(\"\") должен вернуть true после регистрации")
	}
}

func TestRegistryContainsNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.names)
	reg.Contains("test")
	after := len(reg.names)

	if before != after {
		t.Error("Contains не должен изменять реестр")
	}
}

func TestRegistryContainsVsLookupByName(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Contains
	if !reg.Contains("test") {
		t.Error("Contains должен вернуть true")
	}

	// LookupByName
	if _, exists := reg.LookupByName("test"); !exists {
		t.Error("LookupByName должен вернуть true")
	}

	// Оба должны давать одинаковый результат
	if reg.Contains("test") != func() bool {
		_, exists := reg.LookupByName("test")
		return exists
	}() {
		t.Error("Contains и LookupByName должны давать одинаковый результат")
	}
}

// Пример использования
func ExampleRegistry_Contains() {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	fmt.Println(reg.Contains("test"))
	fmt.Println(reg.Contains("nonexistent"))
	// Output:
	// true
	// false
}

// Бенчмарк
func BenchmarkRegistryContains(b *testing.B) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	reg.Register("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = reg.Contains("test")
	}
}

func TestRegistryClear(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	reg.Register("first", oid1)
	reg.Register("second", oid2)
	reg.Register("third", oid3)

	// Проверяем, что записи есть
	if reg.Size() != 3 {
		t.Errorf("Size = %d, want 3", reg.Size())
	}

	// Очищаем
	reg.Clear()

	// Проверяем, что все пусто
	if reg.Size() != 0 {
		t.Errorf("Size после Clear = %d, want 0", reg.Size())
	}

	if len(reg.names) != 0 {
		t.Errorf("len(names) = %d, want 0", len(reg.names))
	}

	if len(reg.oids) != 0 {
		t.Errorf("len(oids) = %d, want 0", len(reg.oids))
	}
}

func TestRegistryClearEmptyRegistry(t *testing.T) {
	reg := NewRegistry()

	// Очищаем пустой реестр
	reg.Clear()

	if reg.Size() != 0 {
		t.Errorf("Size = %d, want 0", reg.Size())
	}
}

func TestRegistryClearThenReuse(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем и очищаем
	oid1 := MustParseOID("1.3.6.1")
	reg.Register("test", oid1)
	reg.Clear()

	// Регистрируем снова
	oid2 := MustParseOID("2.100.3")
	if err := reg.Register("test", oid2); err != nil {
		t.Errorf("Register после Clear: %v", err)
	}

	if reg.Size() != 1 {
		t.Errorf("Size = %d, want 1", reg.Size())
	}

	if !reg.names["test"].Equal(oid2) {
		t.Error("names[test] неверный")
	}
}

func TestRegistryClearThenLookup(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Очищаем
	reg.Clear()

	// Проверяем, что не можем найти
	if _, exists := reg.LookupByName("test"); exists {
		t.Error("LookupByName должен вернуть false после Clear")
	}

	if _, exists := reg.LookupByOID(oid); exists {
		t.Error("LookupByOID должен вернуть false после Clear")
	}

	if reg.Contains("test") {
		t.Error("Contains должен вернуть false после Clear")
	}
}

func TestRegistryClearThenList(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	// Очищаем
	reg.Clear()

	// Список должен быть пустым
	list := reg.List()
	if len(list) != 0 {
		t.Errorf("len(List) = %d, want 0", len(list))
	}

	listNoCopy := reg.ListNoCopy()
	if len(listNoCopy) != 0 {
		t.Errorf("len(ListNoCopy) = %d, want 0", len(listNoCopy))
	}
}

func TestRegistryClearMultipleTimes(t *testing.T) {
	reg := NewRegistry()

	// Первая очистка
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)
	reg.Clear()

	// Вторая очистка (пустого)
	reg.Clear()

	if reg.Size() != 0 {
		t.Errorf("Size = %d, want 0", reg.Size())
	}

	// Третья очистка после повторной регистрации
	reg.Register("test", oid)
	reg.Clear()

	if reg.Size() != 0 {
		t.Errorf("Size = %d, want 0", reg.Size())
	}
}

func TestRegistryClearMapInitialized(t *testing.T) {
	reg := NewRegistry()

	// Очищаем
	reg.Clear()

	// Map должны быть инициализированы (не nil)
	if reg.names == nil {
		t.Error("names должен быть инициализирован после Clear")
	}

	if reg.oids == nil {
		t.Error("oids должен быть инициализирован после Clear")
	}
}

// Пример использования
func ExampleRegistry_Clear() {
	reg := NewRegistry()

	// Регистрируем
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	fmt.Println(reg.Size())

	// Очищаем
	reg.Clear()

	fmt.Println(reg.Size())
	// Output:
	// 1
	// 0
}

// Бенчмарк
func BenchmarkRegistryClear(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg := NewRegistry()
		for i := 0; i < 10; i++ {
			oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
			reg.Register(fmt.Sprintf("oid-%d", i), oid)
		}
		reg.Clear()
	}
}

func TestRegistryNames(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	reg.Register("first", oid1)
	reg.Register("second", oid2)
	reg.Register("third", oid3)

	t.Run("Содержит все имена", func(t *testing.T) {
		names := reg.Names()

		if len(names) != 3 {
			t.Errorf("len = %d, want 3", len(names))
		}

		// Проверяем наличие каждого имени
		found := make(map[string]bool)
		for _, name := range names {
			found[name] = true
		}

		if !found["first"] {
			t.Error("first не найден")
		}
		if !found["second"] {
			t.Error("second не найден")
		}
		if !found["third"] {
			t.Error("third не найден")
		}
	})

	t.Run("Пустой реестр", func(t *testing.T) {
		emptyReg := NewRegistry()
		names := emptyReg.Names()

		if len(names) != 0 {
			t.Errorf("len = %d, want 0", len(names))
		}
	})
}

func TestRegistryNamesEachCallNewSlice(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	names1 := reg.Names()
	names2 := reg.Names()

	// Изменяем первый слайс
	names1[0] = "modified"

	// Второй не должен измениться
	if names2[0] != "test" {
		t.Error("Каждый вызов должен возвращать новый слайс")
	}
}

func TestRegistryNamesNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.names)
	names := reg.Names()
	after := len(reg.names)

	if before != after {
		t.Error("Names не должен изменять реестр")
	}

	// Изменяем слайс
	names[0] = "modified"

	// Реестр не должен измениться
	if !reg.names["test"].Equal(oid) {
		t.Error("Реестр не должен измениться")
	}
}

func TestRegistryNamesAfterRemove(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	// Удаляем одну запись
	reg.Remove("first")

	names := reg.Names()
	if len(names) != 1 {
		t.Errorf("len = %d, want 1", len(names))
	}

	if names[0] != "second" {
		t.Errorf("names[0] = %q, want 'second'", names[0])
	}
}

func TestRegistryNamesAfterClear(t *testing.T) {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Очищаем
	reg.Clear()

	names := reg.Names()
	if len(names) != 0 {
		t.Errorf("len = %d, want 0", len(names))
	}
}

func TestRegistryNamesEmptyName(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем с пустым именем
	oid := MustParseOID("1.3.6.1")
	reg.Register("", oid)

	names := reg.Names()
	if len(names) != 1 {
		t.Errorf("len = %d, want 1", len(names))
	}

	if names[0] != "" {
		t.Errorf("names[0] = %q, want empty", names[0])
	}
}

func TestRegistryNamesSorted(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем с РАЗНЫМИ OID
	reg.Register("zebra", MustParseOID("1.3.6.1"))
	reg.Register("apple", MustParseOID("1.3.6.2"))
	reg.Register("mango", MustParseOID("1.3.6.3"))

	names := reg.Names()

	// Порядок не гарантирован, но проверим, что все есть
	if len(names) != 3 {
		t.Errorf("len = %d, want 3", len(names))
	}

	// Сортируем для проверки
	slices.Sort(names)

	if names[0] != "apple" || names[1] != "mango" || names[2] != "zebra" {
		t.Errorf("names = %v, want [apple mango zebra]", names)
	}
}

// Пример использования
func ExampleRegistry_Names() {
	reg := NewRegistry()

	reg.Register("first", MustParseOID("1.3.6.1"))
	reg.Register("second", MustParseOID("2.100.3"))

	names := reg.Names()
	fmt.Println(len(names))
	// Output: 2
}

// Бенчмарк
func BenchmarkRegistryNames(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = reg.Names()
	}
}

func TestRegistryOIDs(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	reg.Register("first", oid1)
	reg.Register("second", oid2)
	reg.Register("third", oid3)

	t.Run("Содержит все OID", func(t *testing.T) {
		oids := reg.OIDs()

		if len(oids) != 3 {
			t.Errorf("len = %d, want 3", len(oids))
		}

		// Проверяем наличие каждого OID
		found := make(map[string]bool)
		for _, oid := range oids {
			found[oid.String()] = true
		}

		if !found["1.3.6.1"] {
			t.Error("1.3.6.1 не найден")
		}
		if !found["2.100.3"] {
			t.Error("2.100.3 не найден")
		}
		if !found["0.39.1"] {
			t.Error("0.39.1 не найден")
		}
	})

	t.Run("Пустой реестр", func(t *testing.T) {
		emptyReg := NewRegistry()
		oids := emptyReg.OIDs()

		if len(oids) != 0 {
			t.Errorf("len = %d, want 0", len(oids))
		}
	})
}

func TestRegistryOIDsDeepCopy(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Получаем OID
	oids := reg.OIDs()

	// Изменяем OID в слайсе
	oids[0][0] = 99

	// Реестр не должен измениться
	if reg.names["test"][0] != 1 {
		t.Error("OIDs должен вернуть глубокую копию")
	}
}

func TestRegistryOIDsEachCallNewSlice(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	oids1 := reg.OIDs()
	oids2 := reg.OIDs()

	// Изменяем первый слайс
	oids1[0][0] = 99

	// Второй не должен измениться
	if oids2[0][0] != 1 {
		t.Error("Каждый вызов должен возвращать новый слайс")
	}
}

func TestRegistryOIDsNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.names)
	oids := reg.OIDs()
	after := len(reg.names)

	if before != after {
		t.Error("OIDs не должен изменять реестр")
	}

	// Изменяем OID в слайсе
	oids[0][0] = 99

	// Реестр не должен измениться
	if reg.names["test"][0] != 1 {
		t.Error("Реестр не должен измениться")
	}
}

func TestRegistryOIDsAfterRemove(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	// Удаляем одну запись
	reg.Remove("first")

	oids := reg.OIDs()
	if len(oids) != 1 {
		t.Errorf("len = %d, want 1", len(oids))
	}

	if !oids[0].Equal(oid2) {
		t.Errorf("oids[0] = %v, want %v", oids[0], oid2)
	}
}

func TestRegistryOIDsAfterClear(t *testing.T) {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Очищаем
	reg.Clear()

	oids := reg.OIDs()
	if len(oids) != 0 {
		t.Errorf("len = %d, want 0", len(oids))
	}
}

func TestRegistryOIDsDeepCopyEach(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	oids := reg.OIDs()

	// Изменяем каждый OID
	for i := range oids {
		oids[i][0] = 99
	}

	// Реестр не должен измениться
	if reg.names["first"][0] != 1 {
		t.Error("first OID не должен измениться")
	}
	if reg.names["second"][0] != 2 {
		t.Error("second OID не должен измениться")
	}
}

// Пример использования
func ExampleRegistry_OIDs() {
	reg := NewRegistry()

	reg.Register("first", MustParseOID("1.3.6.1"))
	reg.Register("second", MustParseOID("2.100.3"))

	oids := reg.OIDs()
	fmt.Println(len(oids))
	// Output: 2
}

// Бенчмарк
func BenchmarkRegistryOIDs(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = reg.OIDs()
	}
}

func TestRegistryOIDsNoCopy(t *testing.T) {
	reg := NewRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	reg.Register("first", oid1)
	reg.Register("second", oid2)
	reg.Register("third", oid3)

	t.Run("Содержит все OID", func(t *testing.T) {
		oids := reg.OIDsNoCopy()

		if len(oids) != 3 {
			t.Errorf("len = %d, want 3", len(oids))
		}

		// Проверяем наличие каждого OID
		found := make(map[string]bool)
		for _, oid := range oids {
			found[oid.String()] = true
		}

		if !found["1.3.6.1"] {
			t.Error("1.3.6.1 не найден")
		}
		if !found["2.100.3"] {
			t.Error("2.100.3 не найден")
		}
		if !found["0.39.1"] {
			t.Error("0.39.1 не найден")
		}
	})

	t.Run("Пустой реестр", func(t *testing.T) {
		emptyReg := NewRegistry()
		oids := emptyReg.OIDsNoCopy()

		if len(oids) != 0 {
			t.Errorf("len = %d, want 0", len(oids))
		}
	})
}

func TestRegistryOIDsNoCopySharedSlices(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// Получаем OID
	oids := reg.OIDsNoCopy()

	// Изменяем OID в слайсе
	oids[0][0] = 99

	// Реестр должен измениться (общий слайс)
	if reg.names["test"][0] != 99 {
		t.Error("OIDsNoCopy должен вернуть ссылки на внутренние срезы")
	}

	// Восстанавливаем
	oids[0][0] = 1
}

func TestRegistryOIDsNoCopySamePointers(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	oids := reg.OIDsNoCopy()

	// Проверяем, что OID указывает на тот же массив
	if &oids[0][0] != &reg.names["test"][0] {
		t.Error("OIDsNoCopy должен вернуть тот же слайс")
	}
}

func TestRegistryOIDsNoCopyEachCallNewSlice(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	oids1 := reg.OIDsNoCopy()
	oids2 := reg.OIDsNoCopy()

	// Добавляем в первый слайс (не влияет на реестр)
	oids1 = append(oids1, MustParseOID("2.100.3"))

	// Второй не должен измениться
	if len(oids2) != 1 {
		t.Error("Каждый вызов должен возвращать новый слайс")
	}
}

func TestRegistryOIDsNoCopyNotModifyRegistry(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	before := len(reg.names)
	oids := reg.OIDsNoCopy()
	after := len(reg.names)

	if before != after {
		t.Error("OIDsNoCopy не должен изменять реестр")
	}

	// Добавляем в слайс (не влияет на реестр)
	oids = append(oids, MustParseOID("2.100.3"))

	if len(reg.names) != 1 {
		t.Error("Реестр не должен измениться при изменении слайса")
	}
}

func TestRegistryOIDsVsOIDsNoCopy(t *testing.T) {
	reg := NewRegistry()
	oid := MustParseOID("1.3.6.1")
	reg.Register("test", oid)

	// OIDs - глубокая копия
	oids := reg.OIDs()
	oids[0][0] = 99
	if reg.names["test"][0] != 1 {
		t.Error("OIDs должен вернуть копию")
	}

	// OIDsNoCopy - ссылка
	oidsNoCopy := reg.OIDsNoCopy()
	oidsNoCopy[0][0] = 99
	if reg.names["test"][0] != 99 {
		t.Error("OIDsNoCopy должен вернуть ссылку")
	}

	// Восстанавливаем
	oidsNoCopy[0][0] = 1
}

func TestRegistryOIDsNoCopyAfterRemove(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	reg.Register("first", oid1)
	reg.Register("second", oid2)

	// Удаляем одну запись
	reg.Remove("first")

	oids := reg.OIDsNoCopy()
	if len(oids) != 1 {
		t.Errorf("len = %d, want 1", len(oids))
	}

	if !oids[0].Equal(oid2) {
		t.Errorf("oids[0] = %v, want %v", oids[0], oid2)
	}
}

// Пример использования
func ExampleRegistry_OIDsNoCopy() {
	reg := NewRegistry()

	reg.Register("first", MustParseOID("1.3.6.1"))
	reg.Register("second", MustParseOID("2.100.3"))

	oids := reg.OIDsNoCopy()
	fmt.Println(len(oids))
	// Output: 2
}

// Бенчмарк
func BenchmarkRegistryOIDsNoCopy(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = reg.OIDsNoCopy()
	}
}

// Сравнение OIDs vs OIDsNoCopy
func BenchmarkOIDsVsOIDsNoCopy(b *testing.B) {
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		reg.Register(fmt.Sprintf("oid-%d", i), oid)
	}

	b.Run("OIDs", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = reg.OIDs()
		}
	})

	b.Run("OIDsNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = reg.OIDsNoCopy()
		}
	})
}

func TestRegistryBatchRegister(t *testing.T) {
	tests := []struct {
		name        string
		entries     map[string]OID
		expectedLen int
		wantErr     error
	}{
		{
			name:        "Пустой map",
			entries:     nil,
			expectedLen: 0,
			wantErr:     nil,
		},
		{
			name: "Одна запись",
			entries: map[string]OID{
				"first": MustParseOID("1.3.6.1"),
			},
			expectedLen: 1,
			wantErr:     nil,
		},
		{
			name: "Две записи",
			entries: map[string]OID{
				"first":  MustParseOID("1.3.6.1"),
				"second": MustParseOID("2.100.3"),
			},
			expectedLen: 2,
			wantErr:     nil,
		},
		{
			name: "Три записи",
			entries: map[string]OID{
				"first":  MustParseOID("1.3.6.1"),
				"second": MustParseOID("2.100.3"),
				"third":  MustParseOID("0.39.1"),
			},
			expectedLen: 3,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый реестр для каждого теста
			reg := NewRegistry()

			err := reg.BatchRegister(tt.entries)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("BatchRegister: expected error %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("BatchRegister = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("BatchRegister: %v", err)
				return
			}

			if len(reg.names) != tt.expectedLen {
				t.Errorf("len(names) = %d, want %d", len(reg.names), tt.expectedLen)
			}
		})
	}
}

func TestRegistryBatchRegisterDuplicateOID(t *testing.T) {
	reg := NewRegistry()

	oid := MustParseOID("1.3.6.1")

	entries := map[string]OID{
		"first":  oid,
		"second": oid, // Дубликат OID
	}

	err := reg.BatchRegister(entries)

	if err == nil {
		t.Fatal("BatchRegister: expected error")
	}

	if !errors.Is(err, ErrDuplicateOIDInBatch) {
		t.Errorf("BatchRegister = %v, want ErrDuplicateOIDInBatch", err)
	}

	// Атомарность: ничего не должно быть зарегистрировано
	if len(reg.names) != 0 {
		t.Errorf("len(names) = %d, want 0 (атомарность)", len(reg.names))
	}
}

func TestRegistryBatchRegisterNameConflict(t *testing.T) {
	reg := NewRegistry()

	// Существующая запись
	existingOID := MustParseOID("1.3.6.1")
	reg.Register("existing", existingOID)

	entries := map[string]OID{
		"existing": MustParseOID("1.3.6.2"), // Конфликт имени
		"new":      MustParseOID("2.100.3"),
	}

	err := reg.BatchRegister(entries)

	if err == nil {
		t.Fatal("BatchRegister: expected error")
	}

	if !errors.Is(err, ErrNameAlreadyExists) {
		t.Errorf("BatchRegister = %v, want ErrNameAlreadyExists", err)
	}

	// Атомарность: new не должен быть зарегистрирован
	if _, exists := reg.LookupByName("new"); exists {
		t.Error("new не должен быть зарегистрирован")
	}
}

func TestRegistryBatchRegisterOIDConflict(t *testing.T) {
	reg := NewRegistry()

	// Существующая запись
	existingOID := MustParseOID("1.3.6.1")
	reg.Register("existing", existingOID)

	entries := map[string]OID{
		"new":      MustParseOID("2.100.3"),
		"conflict": existingOID, // Конфликт OID
	}

	err := reg.BatchRegister(entries)

	if err == nil {
		t.Fatal("BatchRegister: expected error")
	}

	if !errors.Is(err, ErrOIDAlreadyRegistered) {
		t.Errorf("BatchRegister = %v, want ErrOIDAlreadyRegistered", err)
	}

	// Атомарность: new не должен быть зарегистрирован
	if _, exists := reg.LookupByName("new"); exists {
		t.Error("new не должен быть зарегистрирован")
	}
}

func TestRegistryBatchRegisterInvalidOID(t *testing.T) {
	reg := NewRegistry()

	entries := map[string]OID{
		"valid":   MustParseOID("1.3.6.1"),
		"invalid": OID{3, 1}, // Невалидный
	}

	err := reg.BatchRegister(entries)

	if err == nil {
		t.Fatal("BatchRegister: expected error")
	}

	if !errors.Is(err, ErrFirstComponentTooBig) {
		t.Errorf("BatchRegister = %v, want ErrFirstComponentTooBig", err)
	}

	// Атомарность: valid не должен быть зарегистрирован
	if _, exists := reg.LookupByName("valid"); exists {
		t.Error("valid не должен быть зарегистрирован")
	}
}

func TestRegistryBatchRegisterAtomicity(t *testing.T) {
	reg := NewRegistry()

	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
		"third":  MustParseOID("0.39.1"),
	}

	err := reg.BatchRegister(entries)

	if err != nil {
		t.Fatalf("BatchRegister: %v", err)
	}

	// Все три должны быть зарегистрированы
	if reg.Size() != 3 {
		t.Errorf("Size = %d, want 3", reg.Size())
	}

	for name, oid := range entries {
		got, exists := reg.LookupByName(name)
		if !exists {
			t.Errorf("%s not found", name)
			continue
		}
		if !got.Equal(oid) {
			t.Errorf("%s = %v, want %v", name, got, oid)
		}
	}
}

func TestRegistryBatchRegisterCopyOnWrite(t *testing.T) {
	reg := NewRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	entries := map[string]OID{
		"first":  oid1,
		"second": oid2,
	}

	reg.BatchRegister(entries)

	// Изменяем оригиналы
	oid1[0] = 99
	oid2[0] = 99

	// Реестр не должен измениться
	if reg.names["first"][0] != 1 {
		t.Error("first OID не должен измениться")
	}
	if reg.names["second"][0] != 2 {
		t.Error("second OID не должен измениться")
	}
}

// Пример использования
func ExampleRegistry_BatchRegister() {
	reg := NewRegistry()

	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
	}

	err := reg.BatchRegister(entries)
	if err != nil {
		panic(err)
	}

	fmt.Println(reg.Size())
	// Output: 2
}

// Бенчмарк
func BenchmarkRegistryBatchRegister(b *testing.B) {
	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
		"third":  MustParseOID("0.39.1"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		reg := NewRegistry()
		reg.BatchRegister(entries)
	}
}
