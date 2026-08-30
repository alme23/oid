// oid/global_register_test.go
package oid

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
)

func TestDefaultRegistry(t *testing.T) {
	t.Run("Инициализирован", func(t *testing.T) {
		if defaultRegistry == nil {
			t.Fatal("defaultRegistry не инициализирован")
		}
	})

	t.Run("Тип *Registry", func(t *testing.T) {
		var _ *Registry = defaultRegistry
	})

	t.Run("Пустой по умолчанию", func(t *testing.T) {
		ResetRegistry()

		if defaultRegistry.Size() != 0 {
			t.Errorf("Size = %d, want 0", defaultRegistry.Size())
		}
	})
}

func TestDefaultRegistryUsedByGlobalFunctions(t *testing.T) {
	ResetRegistry()

	// Глобальная функция использует defaultRegistry
	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	// Проверяем через defaultRegistry напрямую
	if defaultRegistry.Size() != 1 {
		t.Errorf("defaultRegistry.Size = %d, want 1", defaultRegistry.Size())
	}

	if !defaultRegistry.Contains("test") {
		t.Error("defaultRegistry должен содержать 'test'")
	}
}

func TestDefaultRegistryReset(t *testing.T) {
	ResetRegistry()

	// Сохраняем старый
	oldReg := defaultRegistry

	// Сбрасываем
	ResetRegistry()

	// Проверяем, что новый
	if defaultRegistry == oldReg {
		t.Error("ResetRegistry должен создавать новый реестр")
	}
}

func TestDefaultRegistryIsolation(t *testing.T) {
	ResetRegistry()

	// Регистрируем через глобальную функцию
	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	// Создаем отдельный реестр
	separateReg := NewRegistry()

	// Отдельный реестр не должен видеть записи defaultRegistry
	if separateReg.Size() != 0 {
		t.Error("Отдельный реестр должен быть пустым")
	}

	// defaultRegistry должен видеть запись
	if defaultRegistry.Size() != 1 {
		t.Error("defaultRegistry должен содержать запись")
	}
}

func TestDefaultRegistryConcurrent(t *testing.T) {
	ResetRegistry()

	// Конкурентная регистрация через глобальные функции
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			oid := MustParseOID(fmt.Sprintf("1.3.6.%d", id+1))
			Register(fmt.Sprintf("oid-%d", id), oid)
		}(i)
	}

	wg.Wait()

	if Size() != 10 {
		t.Errorf("Size = %d, want 10", Size())
	}
}

// Бенчмарк
func BenchmarkDefaultRegistryAccess(b *testing.B) {
	ResetRegistry()
	MustRegister("test", MustParseOID("1.3.6.1"))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = defaultRegistry.Size()
	}
}

func TestGlobalRegister(t *testing.T) {
	// Очищаем перед тестом
	ResetRegistry()

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
			err := Register(tt.oidName, tt.oid)

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

			if Size() != tt.expectedLen {
				t.Errorf("Size = %d, want %d", Size(), tt.expectedLen)
			}
		})
	}
}

func TestGlobalRegisterValidation(t *testing.T) {
	ResetRegistry()

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
			err := Register(tt.oidName, tt.oid)

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

func TestGlobalRegisterCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	Register("test", oid)

	// Изменяем оригинал
	oid[0] = 99

	// Реестр не должен измениться
	if o, exists := LookupByName("test"); exists {
		if o[0] != 1 {
			t.Error("Register должен создать копию")
		}
	}
}

func TestGlobalRegisterOverwrite(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	Register("test", oid1)
	Register("test", oid2)

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}

	if o, exists := LookupByName("test"); exists {
		if !o.Equal(oid2) {
			t.Error("Должен сохраниться второй OID")
		}
	}
}

func TestGlobalRegisterAfterReset(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	Register("test", oid)

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}

	// Сбрасываем
	ResetRegistry()

	if Size() != 0 {
		t.Errorf("Size после Reset = %d, want 0", Size())
	}

	// Можно регистрировать снова
	if err := Register("test", oid); err != nil {
		t.Errorf("Register после Reset: %v", err)
	}
}

func TestGlobalRegisterAfterClear(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	Register("test", oid)

	Clear()

	if Size() != 0 {
		t.Errorf("Size после Clear = %d, want 0", Size())
	}

	// Можно регистрировать снова
	if err := Register("test", oid); err != nil {
		t.Errorf("Register после Clear: %v", err)
	}
}

// Пример использования
func ExampleRegister() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	Register("test", oid)

	if o, exists := LookupByName("test"); exists {
		fmt.Println(o)
	}
	// Output: 1.3.6.1
}

// Пример с ошибкой
func ExampleRegister_error() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	Register("test", oid)

	err := Register("other", oid)
	fmt.Println(errors.Is(err, ErrOIDAlreadyRegistered))
	// Output: true
}

// Бенчмарк
func BenchmarkGlobalRegister(b *testing.B) {
	ResetRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ResetRegistry()
		_ = Register("test", oid)
	}
}

func TestMustRegister(t *testing.T) {
	ResetRegistry()

	tests := []struct {
		name        string
		oidName     string
		oid         OID
		expectedLen int
	}{
		{
			name:        "Первая регистрация",
			oidName:     "test",
			oid:         MustParseOID("1.3.6.1"),
			expectedLen: 1,
		},
		{
			name:        "Вторая регистрация",
			oidName:     "second",
			oid:         MustParseOID("2.100.3"),
			expectedLen: 2,
		},
		{
			name:        "Перезапись",
			oidName:     "test",
			oid:         MustParseOID("1.3.6.2"),
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MustRegister(tt.oidName, tt.oid)

			if Size() != tt.expectedLen {
				t.Errorf("Size = %d, want %d", Size(), tt.expectedLen)
			}
		})
	}
}

func TestMustRegisterPanic(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	tests := []struct {
		name    string
		oidName string
		oid     OID
	}{
		{
			name:    "Дубликат OID",
			oidName: "other",
			oid:     oid,
		},
		{
			name:    "Пустой OID",
			oidName: "empty",
			oid:     OID{},
		},
		{
			name:    "Невалидный OID",
			oidName: "invalid",
			oid:     OID{3, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("MustRegister(%q): expected panic", tt.oidName)
				}
			}()

			MustRegister(tt.oidName, tt.oid)
		})
	}
}

func TestMustRegisterPanicMessage(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	t.Run("Паника содержит имя OID", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic")
				return
			}

			msg := fmt.Sprintf("%v", r)
			if msg == "" {
				t.Error("Сообщение паники пустое")
			}

			t.Logf("Panic: %s", msg)
		}()

		MustRegister("other", oid)
	})
}

func TestMustRegisterAfterReset(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}

	// Сбрасываем
	ResetRegistry()

	// Можно регистрировать снова
	MustRegister("test", oid)

	if Size() != 1 {
		t.Errorf("Size после Reset = %d, want 1", Size())
	}
}

func TestMustRegisterAfterClear(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Clear()

	// Можно регистрировать снова
	MustRegister("test", oid)

	if Size() != 1 {
		t.Errorf("Size после Clear = %d, want 1", Size())
	}
}

// Пример использования
func ExampleMustRegister() {
	ResetRegistry()

	MustRegister("test", MustParseOID("1.3.6.1"))

	if o, exists := LookupByName("test"); exists {
		fmt.Println(o)
	}
	// Output: 1.3.6.1
}

// Пример с паникой
func ExampleMustRegister_panic() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Паника поймана")
		}
	}()

	MustRegister("other", oid)
	// Output: Паника поймана
}

// Бенчмарк
func BenchmarkMustRegister(b *testing.B) {
	ResetRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ResetRegistry()
		MustRegister("test", oid)
	}
}

func TestGlobalBatchRegister(t *testing.T) {
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
			// Сбрасываем реестр перед каждым тестом
			ResetRegistry()

			err := BatchRegister(tt.entries)

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

			if Size() != tt.expectedLen {
				t.Errorf("Size = %d, want %d", Size(), tt.expectedLen)
			}
		})
	}
}

func TestGlobalBatchRegisterDuplicateOID(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")

	entries := map[string]OID{
		"first":  oid,
		"second": oid, // Дубликат OID
	}

	err := BatchRegister(entries)

	if err == nil {
		t.Fatal("BatchRegister: expected error")
	}

	if !errors.Is(err, ErrDuplicateOIDInBatch) {
		t.Errorf("BatchRegister = %v, want ErrDuplicateOIDInBatch", err)
	}

	// Атомарность: ничего не должно быть зарегистрировано
	if Size() != 0 {
		t.Errorf("Size = %d, want 0 (атомарность)", Size())
	}
}

func TestGlobalBatchRegisterNameConflict(t *testing.T) {
	ResetRegistry()

	// Существующая запись
	existingOID := MustParseOID("1.3.6.1")
	MustRegister("existing", existingOID)

	entries := map[string]OID{
		"existing": MustParseOID("1.3.6.2"), // Конфликт имени
		"new":      MustParseOID("2.100.3"),
	}

	err := BatchRegister(entries)

	if err == nil {
		t.Fatal("BatchRegister: expected error")
	}

	if !errors.Is(err, ErrNameAlreadyExists) {
		t.Errorf("BatchRegister = %v, want ErrNameAlreadyExists", err)
	}

	// Атомарность: new не должен быть зарегистрирован
	if _, exists := LookupByName("new"); exists {
		t.Error("new не должен быть зарегистрирован")
	}
}

func TestGlobalBatchRegisterOIDConflict(t *testing.T) {
	ResetRegistry()

	existingOID := MustParseOID("1.3.6.1")
	MustRegister("existing", existingOID)

	entries := map[string]OID{
		"new":      MustParseOID("2.100.3"),
		"conflict": existingOID, // Конфликт OID
	}

	err := BatchRegister(entries)

	if err == nil {
		t.Fatal("BatchRegister: expected error")
	}

	if !errors.Is(err, ErrOIDAlreadyRegistered) {
		t.Errorf("BatchRegister = %v, want ErrOIDAlreadyRegistered", err)
	}

	if _, exists := LookupByName("new"); exists {
		t.Error("new не должен быть зарегистрирован")
	}
}

func TestGlobalBatchRegisterInvalidOID(t *testing.T) {
	ResetRegistry()

	entries := map[string]OID{
		"valid":   MustParseOID("1.3.6.1"),
		"invalid": OID{3, 1},
	}

	err := BatchRegister(entries)

	if err == nil {
		t.Fatal("BatchRegister: expected error")
	}

	if !errors.Is(err, ErrFirstComponentTooBig) {
		t.Errorf("BatchRegister = %v, want ErrFirstComponentTooBig", err)
	}

	if _, exists := LookupByName("valid"); exists {
		t.Error("valid не должен быть зарегистрирован")
	}
}

func TestGlobalBatchRegisterAtomicity(t *testing.T) {
	ResetRegistry()

	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
		"third":  MustParseOID("0.39.1"),
	}

	err := BatchRegister(entries)

	if err != nil {
		t.Fatalf("BatchRegister: %v", err)
	}

	if Size() != 3 {
		t.Errorf("Size = %d, want 3", Size())
	}

	for name, oid := range entries {
		got, exists := LookupByName(name)
		if !exists {
			t.Errorf("%s not found", name)
			continue
		}
		if !got.Equal(oid) {
			t.Errorf("%s = %v, want %v", name, got, oid)
		}
	}
}

// Пример использования
func ExampleBatchRegister() {
	ResetRegistry()

	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
	}

	err := BatchRegister(entries)
	if err != nil {
		panic(err)
	}

	fmt.Println(Size())
	// Output: 2
}

// Пример с ошибкой
func ExampleBatchRegister_error() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")

	entries := map[string]OID{
		"first":  oid,
		"second": oid,
	}

	err := BatchRegister(entries)
	fmt.Println(errors.Is(err, ErrDuplicateOIDInBatch))
	// Output: true
}

// Бенчмарк
func BenchmarkGlobalBatchRegister(b *testing.B) {
	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
		"third":  MustParseOID("0.39.1"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ResetRegistry()
		_ = BatchRegister(entries)
	}
}

func TestMustBatchRegister(t *testing.T) {
	tests := []struct {
		name        string
		entries     map[string]OID
		expectedLen int
	}{
		{
			name:        "Пустой map",
			entries:     nil,
			expectedLen: 0,
		},
		{
			name: "Одна запись",
			entries: map[string]OID{
				"first": MustParseOID("1.3.6.1"),
			},
			expectedLen: 1,
		},
		{
			name: "Две записи",
			entries: map[string]OID{
				"first":  MustParseOID("1.3.6.1"),
				"second": MustParseOID("2.100.3"),
			},
			expectedLen: 2,
		},
		{
			name: "Три записи",
			entries: map[string]OID{
				"first":  MustParseOID("1.3.6.1"),
				"second": MustParseOID("2.100.3"),
				"third":  MustParseOID("0.39.1"),
			},
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetRegistry()

			MustBatchRegister(tt.entries)

			if Size() != tt.expectedLen {
				t.Errorf("Size = %d, want %d", Size(), tt.expectedLen)
			}
		})
	}
}

func TestMustBatchRegisterPanic(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() // Функция для подготовки
		entries map[string]OID
	}{
		{
			name: "Дубликат OID",
			entries: map[string]OID{
				"first":  MustParseOID("1.3.6.1"),
				"second": MustParseOID("1.3.6.1"),
			},
		},
		{
			name: "Невалидный OID",
			entries: map[string]OID{
				"valid":   MustParseOID("1.3.6.1"),
				"invalid": OID{3, 1},
			},
		},
		{
			name: "Конфликт с существующим",
			setup: func() {
				MustRegister("existing", MustParseOID("1.3.6.1"))
			},
			entries: map[string]OID{
				"existing": MustParseOID("1.3.6.2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetRegistry()

			// Выполняем подготовку
			if tt.setup != nil {
				tt.setup()
			}

			defer func() {
				if r := recover(); r == nil {
					t.Errorf("MustBatchRegister: expected panic")
				}
			}()

			MustBatchRegister(tt.entries)
		})
	}
}

func TestMustBatchRegisterAtomicity(t *testing.T) {
	ResetRegistry()

	t.Run("Успешная регистрация", func(t *testing.T) {
		entries := map[string]OID{
			"first":  MustParseOID("1.3.6.1"),
			"second": MustParseOID("2.100.3"),
		}

		MustBatchRegister(entries)

		if Size() != 2 {
			t.Errorf("Size = %d, want 2", Size())
		}
	})

	t.Run("Паника не регистрирует ничего", func(t *testing.T) {
		ResetRegistry()

		entries := map[string]OID{
			"valid":   MustParseOID("1.3.6.1"),
			"invalid": OID{3, 1},
		}

		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic")
				}
			}()

			MustBatchRegister(entries)
		}()

		// Ничего не должно быть зарегистрировано
		if Size() != 0 {
			t.Errorf("Size = %d, want 0 (атомарность)", Size())
		}
	})
}

func TestMustBatchRegisterAfterReset(t *testing.T) {
	ResetRegistry()

	entries := map[string]OID{
		"first": MustParseOID("1.3.6.1"),
	}

	MustBatchRegister(entries)

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}

	// Сбрасываем
	ResetRegistry()

	// Можно регистрировать снова
	MustBatchRegister(entries)

	if Size() != 1 {
		t.Errorf("Size после Reset = %d, want 1", Size())
	}
}

// Пример использования
func ExampleMustBatchRegister() {
	ResetRegistry()

	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
	}

	MustBatchRegister(entries)

	fmt.Println(Size())
	// Output: 2
}

// Пример с паникой
func ExampleMustBatchRegister_panic() {
	ResetRegistry()

	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("1.3.6.1"),
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Паника поймана")
		}
	}()

	MustBatchRegister(entries)
	// Output: Паника поймана
}

// Бенчмарк
func BenchmarkMustBatchRegister(b *testing.B) {
	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ResetRegistry()
		MustBatchRegister(entries)
	}
}

func TestGlobalLookupByName(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

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
			got, exists := LookupByName(tt.lookup)

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

func TestGlobalLookupByNameCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	got, exists := LookupByName("test")
	if !exists {
		t.Fatal("test not found")
	}

	// Изменяем копию
	got[0] = 99

	// Реестр не должен измениться
	if o, _ := LookupByName("test"); o[0] != 1 {
		t.Error("LookupByName должен вернуть копию")
	}
}

func TestGlobalLookupByNameEachCallNewCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	got1, _ := LookupByName("test")
	got2, _ := LookupByName("test")

	got1[0] = 99

	if got2[0] != 1 {
		t.Error("Каждый вызов должен возвращать новую копию")
	}
}

func TestGlobalLookupByNameNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	LookupByName("test")
	after := Size()

	if before != after {
		t.Error("LookupByName не должен изменять реестр")
	}
}

func TestGlobalLookupByNameAfterRemove(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Remove("test")

	_, exists := LookupByName("test")
	if exists {
		t.Error("После Remove LookupByName должен вернуть false")
	}
}

func TestGlobalLookupByNameAfterClear(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Clear()

	_, exists := LookupByName("test")
	if exists {
		t.Error("После Clear LookupByName должен вернуть false")
	}
}

// Пример использования
func ExampleLookupByName() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	if got, exists := LookupByName("test"); exists {
		fmt.Println(got)
	}

	_, exists := LookupByName("nonexistent")
	fmt.Println(exists)
	// Output:
	// 1.3.6.1
	// false
}

// Бенчмарк
func BenchmarkGlobalLookupByName(b *testing.B) {
	ResetRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	MustRegister("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = LookupByName("test")
	}
}

func TestGlobalLookupByNameNoCopy(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

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
			got, exists := LookupByNameNoCopy(tt.lookup)

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

func TestGlobalLookupByNameNoCopySharedSlice(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	got, exists := LookupByNameNoCopy("test")
	if !exists {
		t.Fatal("test not found")
	}

	// Изменяем через полученную ссылку
	got[0] = 99

	// Реестр должен измениться (общий слайс)
	if o, _ := LookupByNameNoCopy("test"); o[0] != 99 {
		t.Error("LookupByNameNoCopy должен вернуть ссылку")
	}

	// Восстанавливаем
	got[0] = 1
}

func TestGlobalLookupByNameNoCopyNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	LookupByNameNoCopy("test")
	after := Size()

	if before != after {
		t.Error("LookupByNameNoCopy не должен изменять реестр")
	}
}

func TestGlobalLookupByNameVsNoCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	// LookupByName - копия
	copied, _ := LookupByName("test")
	copied[0] = 99
	if o, _ := LookupByName("test"); o[0] != 1 {
		t.Error("LookupByName должен вернуть копию")
	}

	// LookupByNameNoCopy - ссылка
	referenced, _ := LookupByNameNoCopy("test")
	referenced[0] = 99
	if o, _ := LookupByNameNoCopy("test"); o[0] != 99 {
		t.Error("LookupByNameNoCopy должен вернуть ссылку")
	}

	// Восстанавливаем
	referenced[0] = 1
}

// Пример использования
func ExampleLookupByNameNoCopy() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	if got, exists := LookupByNameNoCopy("test"); exists {
		fmt.Println(got)
	}
	// Output: 1.3.6.1
}

// Бенчмарк
func BenchmarkGlobalLookupByNameNoCopy(b *testing.B) {
	ResetRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	MustRegister("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = LookupByNameNoCopy("test")
	}
}

// Сравнение LookupByName vs LookupByNameNoCopy
func BenchmarkGlobalLookupByNameVsNoCopy(b *testing.B) {
	ResetRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	MustRegister("test", oid)

	b.Run("LookupByName", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = LookupByName("test")
		}
	})

	b.Run("LookupByNameNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = LookupByNameNoCopy("test")
		}
	})
}

func TestGlobalLookupByOID(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

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
			got, exists := LookupByOID(tt.oid)

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

func TestGlobalLookupByOIDNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	LookupByOID(oid)
	after := Size()

	if before != after {
		t.Error("LookupByOID не должен изменять реестр")
	}
}

func TestGlobalLookupByOIDRoundTrip(t *testing.T) {
	ResetRegistry()

	oids := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
		"third":  MustParseOID("0.39.1"),
	}

	for name, oid := range oids {
		MustRegister(name, oid)
	}

	for expectedName, oid := range oids {
		gotName, exists := LookupByOID(oid)
		if !exists {
			t.Errorf("OID %v not found", oid)
			continue
		}
		if gotName != expectedName {
			t.Errorf("got %q, want %q", gotName, expectedName)
		}
	}
}

func TestGlobalLookupByOIDAfterOverwrite(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	MustRegister("test", oid1)
	MustRegister("test", oid2)

	// oid1 должен указывать на test (старый)
	if name, exists := LookupByOID(oid1); exists {
		if name != "test" {
			t.Errorf("oid1 -> %q, want 'test'", name)
		}
	} else {
		t.Error("oid1 should still be found")
	}

	// oid2 должен указывать на test (новый)
	if name, exists := LookupByOID(oid2); exists {
		if name != "test" {
			t.Errorf("oid2 -> %q, want 'test'", name)
		}
	} else {
		t.Error("oid2 should be found")
	}
}

// Пример использования
func ExampleLookupByOID() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	if name, exists := LookupByOID(oid); exists {
		fmt.Println(name)
	}

	_, exists := LookupByOID(MustParseOID("1.3.6.99"))
	fmt.Println(exists)
	// Output:
	// test
	// false
}

// Бенчмарк
func BenchmarkGlobalLookupByOID(b *testing.B) {
	ResetRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	MustRegister("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = LookupByOID(oid)
	}
}

func TestGlobalRemove(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

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
			removed := Remove(tt.removeName)

			if removed != tt.expectedRem {
				t.Errorf("removed = %v, want %v", removed, tt.expectedRem)
			}

			if Size() != tt.expectedLen {
				t.Errorf("Size = %d, want %d", Size(), tt.expectedLen)
			}
		})
	}
}

func TestGlobalRemoveFromEmpty(t *testing.T) {
	ResetRegistry()

	if Remove("nonexistent") {
		t.Error("Remove из пустого реестра должен вернуть false")
	}

	if Size() != 0 {
		t.Errorf("Size = %d, want 0", Size())
	}
}

func TestGlobalRemoveThenLookup(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Remove("test")

	// LookupByName не должен найти
	if _, exists := LookupByName("test"); exists {
		t.Error("LookupByName должен вернуть false после Remove")
	}

	// LookupByOID не должен найти
	if _, exists := LookupByOID(oid); exists {
		t.Error("LookupByOID должен вернуть false после Remove")
	}

	// Contains не должен найти
	if Contains("test") {
		t.Error("Contains должен вернуть false после Remove")
	}
}

func TestGlobalRemoveThenReRegister(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Remove("test")

	// Можно регистрировать снова
	if err := Register("test", oid); err != nil {
		t.Errorf("Register после Remove: %v", err)
	}

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}
}

func TestGlobalRemoveEmptyName(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("", oid)

	if !Remove("") {
		t.Error("Remove должен вернуть true для пустого имени")
	}

	if Size() != 0 {
		t.Errorf("Size = %d, want 0", Size())
	}
}

func TestGlobalRemoveOverwritten(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	MustRegister("test", oid1)
	MustRegister("test", oid2)

	Remove("test")

	// oid2 должен быть удален
	if _, exists := LookupByOID(oid2); exists {
		t.Error("oid2 должен быть удален")
	}
}

// Пример использования
func ExampleRemove() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	fmt.Println(Remove("test"))
	fmt.Println(Remove("test"))
	// Output:
	// true
	// false
}

// Бенчмарк
func BenchmarkGlobalRemove(b *testing.B) {
	oid := MustParseOID("1.3.6.1")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ResetRegistry()
		MustRegister("test", oid)
		Remove("test")
	}
}

func TestGlobalList(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	MustRegister("first", oid1)
	MustRegister("second", oid2)
	MustRegister("third", oid3)

	t.Run("Содержит все записи", func(t *testing.T) {
		list := List()

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
		ResetRegistry()

		list := List()

		if len(list) != 0 {
			t.Errorf("len = %d, want 0", len(list))
		}
	})
}

func TestGlobalListDeepCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	list := List()

	// Изменяем OID в списке
	list["test"][0] = 99

	// Реестр не должен измениться
	if o, _ := LookupByName("test"); o[0] != 1 {
		t.Error("List должен вернуть глубокую копию")
	}
}

func TestGlobalListEachCallNewMap(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	list1 := List()
	list2 := List()

	list1["new"] = MustParseOID("2.100.3")

	if _, exists := list2["new"]; exists {
		t.Error("Каждый вызов должен возвращать новый map")
	}
}

func TestGlobalListNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	list := List()
	after := Size()

	if before != after {
		t.Error("List не должен изменять реестр")
	}

	// Изменяем список
	list["test"][0] = 99
	list["new"] = MustParseOID("2.100.3")

	// Реестр не должен измениться
	if Size() != 1 {
		t.Error("Реестр не должен измениться")
	}
	if o, _ := LookupByName("test"); o[0] != 1 {
		t.Error("OID в реестре не должен измениться")
	}
}

func TestGlobalListAfterRemove(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

	Remove("first")

	list := List()

	if len(list) != 1 {
		t.Errorf("len = %d, want 1", len(list))
	}

	if !list["second"].Equal(oid2) {
		t.Error("list[second] неверный")
	}
}

func TestGlobalListAfterClear(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Clear()

	list := List()

	if len(list) != 0 {
		t.Errorf("len = %d, want 0", len(list))
	}
}

// Пример использования
func ExampleList() {
	ResetRegistry()

	MustRegister("first", MustParseOID("1.3.6.1"))
	MustRegister("second", MustParseOID("2.100.3"))

	list := List()

	fmt.Println(len(list))
	// Output: 2
}

// Бенчмарк
func BenchmarkGlobalList(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = List()
	}
}

func TestGlobalListNoCopy(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	MustRegister("first", oid1)
	MustRegister("second", oid2)
	MustRegister("third", oid3)

	t.Run("Содержит все записи", func(t *testing.T) {
		list := ListNoCopy()

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
		ResetRegistry()

		list := ListNoCopy()

		if len(list) != 0 {
			t.Errorf("len = %d, want 0", len(list))
		}
	})
}

func TestGlobalListNoCopySharedSlices(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	list := ListNoCopy()

	// Изменяем OID в списке
	list["test"][0] = 99

	// Реестр должен измениться (общий слайс)
	if o, _ := LookupByNameNoCopy("test"); o[0] != 99 {
		t.Error("ListNoCopy должен вернуть ссылки")
	}

	// Восстанавливаем
	list["test"][0] = 1
}

func TestGlobalListNoCopyEachCallNewMap(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	list1 := ListNoCopy()
	list2 := ListNoCopy()

	list1["new"] = MustParseOID("2.100.3")

	if _, exists := list2["new"]; exists {
		t.Error("Каждый вызов должен возвращать новый map")
	}
}

func TestGlobalListNoCopyNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	list := ListNoCopy()
	after := Size()

	if before != after {
		t.Error("ListNoCopy не должен изменять реестр")
	}

	// Добавляем запись в список (не влияет на реестр)
	list["new"] = MustParseOID("2.100.3")

	if Size() != 1 {
		t.Error("Реестр не должен измениться")
	}
}

func TestGlobalListVsListNoCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	// List - глубокая копия
	list := List()
	list["test"][0] = 99
	if o, _ := LookupByName("test"); o[0] != 1 {
		t.Error("List должен вернуть копию")
	}

	// ListNoCopy - ссылка
	listNoCopy := ListNoCopy()
	listNoCopy["test"][0] = 99
	if o, _ := LookupByNameNoCopy("test"); o[0] != 99 {
		t.Error("ListNoCopy должен вернуть ссылку")
	}

	// Восстанавливаем
	listNoCopy["test"][0] = 1
}

func TestGlobalListNoCopyAfterRemove(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

	Remove("first")

	list := ListNoCopy()

	if len(list) != 1 {
		t.Errorf("len = %d, want 1", len(list))
	}

	if !list["second"].Equal(oid2) {
		t.Error("list[second] неверный")
	}
}

// Пример использования
func ExampleListNoCopy() {
	ResetRegistry()

	MustRegister("first", MustParseOID("1.3.6.1"))
	MustRegister("second", MustParseOID("2.100.3"))

	list := ListNoCopy()

	fmt.Println(len(list))
	// Output: 2
}

// Бенчмарк
func BenchmarkGlobalListNoCopy(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = ListNoCopy()
	}
}

// Сравнение List vs ListNoCopy
func BenchmarkGlobalListVsListNoCopy(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	b.Run("List", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = List()
		}
	})

	b.Run("ListNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = ListNoCopy()
		}
	})
}

func TestGlobalSize(t *testing.T) {
	ResetRegistry()

	t.Run("Пустой реестр", func(t *testing.T) {
		if Size() != 0 {
			t.Errorf("Size = %d, want 0", Size())
		}
	})

	t.Run("Одна запись", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1")
		MustRegister("first", oid)

		if Size() != 1 {
			t.Errorf("Size = %d, want 1", Size())
		}
	})

	t.Run("Две записи", func(t *testing.T) {
		oid := MustParseOID("2.100.3")
		MustRegister("second", oid)

		if Size() != 2 {
			t.Errorf("Size = %d, want 2", Size())
		}
	})

	t.Run("Три записи", func(t *testing.T) {
		oid := MustParseOID("0.39.1")
		MustRegister("third", oid)

		if Size() != 3 {
			t.Errorf("Size = %d, want 3", Size())
		}
	})
}

func TestGlobalSizeAfterRemove(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

	if Size() != 2 {
		t.Errorf("Size = %d, want 2", Size())
	}

	Remove("first")

	if Size() != 1 {
		t.Errorf("Size после Remove = %d, want 1", Size())
	}

	Remove("second")

	if Size() != 0 {
		t.Errorf("Size после Remove = %d, want 0", Size())
	}
}

func TestGlobalSizeAfterClear(t *testing.T) {
	ResetRegistry()

	for i := 0; i < 5; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	if Size() != 5 {
		t.Errorf("Size = %d, want 5", Size())
	}

	Clear()

	if Size() != 0 {
		t.Errorf("Size после Clear = %d, want 0", Size())
	}
}

func TestGlobalSizeAfterOverwrite(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	MustRegister("test", oid1)

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}

	MustRegister("test", oid2)

	if Size() != 1 {
		t.Errorf("Size после перезаписи = %d, want 1", Size())
	}
}

func TestGlobalSizeAfterReset(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}

	ResetRegistry()

	if Size() != 0 {
		t.Errorf("Size после Reset = %d, want 0", Size())
	}
}

func TestGlobalSizeNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	_ = Size()
	after := Size()

	if before != after {
		t.Error("Size не должен изменять реестр")
	}
}

// Пример использования
func ExampleSize() {
	ResetRegistry()

	fmt.Println(Size())

	MustRegister("test", MustParseOID("1.3.6.1"))

	fmt.Println(Size())
	// Output:
	// 0
	// 1
}

// Бенчмарк
func BenchmarkGlobalSize(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = Size()
	}
}

func TestGlobalContains(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

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
			exists := Contains(tt.lookup)

			if exists != tt.expected {
				t.Errorf("Contains(%q) = %v, want %v",
					tt.lookup, exists, tt.expected)
			}
		})
	}
}

func TestGlobalContainsEmptyRegistry(t *testing.T) {
	ResetRegistry()

	if Contains("any") {
		t.Error("Пустой реестр не должен содержать записи")
	}
}

func TestGlobalContainsAfterRemove(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	if !Contains("test") {
		t.Error("Contains должен вернуть true после регистрации")
	}

	Remove("test")

	if Contains("test") {
		t.Error("Contains должен вернуть false после Remove")
	}
}

func TestGlobalContainsAfterClear(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	if !Contains("test") {
		t.Error("Contains должен вернуть true после регистрации")
	}

	Clear()

	if Contains("test") {
		t.Error("Contains должен вернуть false после Clear")
	}
}

func TestGlobalContainsAfterOverwrite(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("1.3.6.2")

	MustRegister("test", oid1)

	if !Contains("test") {
		t.Error("Contains должен вернуть true")
	}

	MustRegister("test", oid2)

	if !Contains("test") {
		t.Error("Contains должен вернуть true после перезаписи")
	}
}

func TestGlobalContainsEmptyName(t *testing.T) {
	ResetRegistry()

	if Contains("") {
		t.Error("Contains(\"\") должен вернуть false для пустого реестра")
	}

	oid := MustParseOID("1.3.6.1")
	MustRegister("", oid)

	if !Contains("") {
		t.Error("Contains(\"\") должен вернуть true после регистрации")
	}
}

func TestGlobalContainsNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	Contains("test")
	after := Size()

	if before != after {
		t.Error("Contains не должен изменять реестр")
	}
}

func TestGlobalContainsNoAllocations(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	allocs := testing.AllocsPerRun(1000, func() {
		_ = Contains("test")
	})

	if allocs != 0 {
		t.Errorf("Contains: %f allocs, want 0", allocs)
	}
}

// Пример использования
func ExampleContains() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	fmt.Println(Contains("test"))
	fmt.Println(Contains("nonexistent"))
	// Output:
	// true
	// false
}

// Бенчмарк
func BenchmarkGlobalContains(b *testing.B) {
	ResetRegistry()
	oid := MustParseOID("1.3.6.1.4.1.99999")
	MustRegister("test", oid)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = Contains("test")
	}
}

func TestGlobalClear(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	MustRegister("first", oid1)
	MustRegister("second", oid2)
	MustRegister("third", oid3)

	// Проверяем, что записи есть
	if Size() != 3 {
		t.Errorf("Size = %d, want 3", Size())
	}

	// Очищаем
	Clear()

	// Проверяем, что все пусто
	if Size() != 0 {
		t.Errorf("Size после Clear = %d, want 0", Size())
	}

	if len(List()) != 0 {
		t.Error("List после Clear должен быть пустым")
	}

	if len(ListNoCopy()) != 0 {
		t.Error("ListNoCopy после Clear должен быть пустым")
	}
}

func TestGlobalClearEmptyRegistry(t *testing.T) {
	ResetRegistry()

	// Очищаем пустой реестр
	Clear()

	if Size() != 0 {
		t.Errorf("Size = %d, want 0", Size())
	}
}

func TestGlobalClearThenReuse(t *testing.T) {
	ResetRegistry()

	// Регистрируем и очищаем
	oid1 := MustParseOID("1.3.6.1")
	MustRegister("test", oid1)
	Clear()

	// Регистрируем снова
	oid2 := MustParseOID("2.100.3")
	if err := Register("test", oid2); err != nil {
		t.Errorf("Register после Clear: %v", err)
	}

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}

	if o, exists := LookupByName("test"); exists {
		if !o.Equal(oid2) {
			t.Error("LookupByName вернул неверный OID")
		}
	}
}

func TestGlobalClearThenLookup(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Clear()

	// Проверяем, что не можем найти
	if _, exists := LookupByName("test"); exists {
		t.Error("LookupByName должен вернуть false после Clear")
	}

	if _, exists := LookupByOID(oid); exists {
		t.Error("LookupByOID должен вернуть false после Clear")
	}

	if Contains("test") {
		t.Error("Contains должен вернуть false после Clear")
	}
}

func TestGlobalClearMultipleTimes(t *testing.T) {
	ResetRegistry()

	// Первая очистка
	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)
	Clear()

	// Вторая очистка (пустого)
	Clear()

	if Size() != 0 {
		t.Errorf("Size = %d, want 0", Size())
	}

	// Третья очистка после повторной регистрации
	MustRegister("test", oid)
	Clear()

	if Size() != 0 {
		t.Errorf("Size = %d, want 0", Size())
	}
}

func TestGlobalClearAfterBatchRegister(t *testing.T) {
	ResetRegistry()

	entries := map[string]OID{
		"first":  MustParseOID("1.3.6.1"),
		"second": MustParseOID("2.100.3"),
	}

	MustBatchRegister(entries)

	if Size() != 2 {
		t.Errorf("Size = %d, want 2", Size())
	}

	Clear()

	if Size() != 0 {
		t.Errorf("Size после Clear = %d, want 0", Size())
	}
}

// Пример использования
func ExampleClear() {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	fmt.Println(Size())

	Clear()

	fmt.Println(Size())
	// Output:
	// 1
	// 0
}

// Бенчмарк
func BenchmarkGlobalClear(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ResetRegistry()
		for i := 0; i < 10; i++ {
			oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
			MustRegister(fmt.Sprintf("oid-%d", i), oid)
		}
		Clear()
	}
}

func TestGlobalNames(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	MustRegister("first", oid1)
	MustRegister("second", oid2)
	MustRegister("third", oid3)

	t.Run("Содержит все имена", func(t *testing.T) {
		names := Names()

		if len(names) != 3 {
			t.Errorf("len = %d, want 3", len(names))
		}

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
		ResetRegistry()

		names := Names()

		if len(names) != 0 {
			t.Errorf("len = %d, want 0", len(names))
		}
	})
}

func TestGlobalNamesEachCallNewSlice(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	names1 := Names()
	names2 := Names()

	names1[0] = "modified"

	if names2[0] != "test" {
		t.Error("Каждый вызов должен возвращать новый слайс")
	}
}

func TestGlobalNamesNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	names := Names()
	after := Size()

	if before != after {
		t.Error("Names не должен изменять реестр")
	}

	// Изменяем слайс
	names[0] = "modified"

	// Реестр не должен измениться
	if !Contains("test") {
		t.Error("Реестр не должен измениться")
	}
}

func TestGlobalNamesAfterRemove(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

	Remove("first")

	names := Names()

	if len(names) != 1 {
		t.Errorf("len = %d, want 1", len(names))
	}

	if names[0] != "second" {
		t.Errorf("names[0] = %q, want 'second'", names[0])
	}
}

func TestGlobalNamesAfterClear(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Clear()

	names := Names()

	if len(names) != 0 {
		t.Errorf("len = %d, want 0", len(names))
	}
}

func TestGlobalNamesEmptyName(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("", oid)

	names := Names()

	if len(names) != 1 {
		t.Errorf("len = %d, want 1", len(names))
	}

	if names[0] != "" {
		t.Errorf("names[0] = %q, want empty", names[0])
	}
}

func TestGlobalNamesSorted(t *testing.T) {
	ResetRegistry()

	// Регистрируем с разными OID
	MustRegister("zebra", MustParseOID("1.3.6.1"))
	MustRegister("apple", MustParseOID("1.3.6.2"))
	MustRegister("mango", MustParseOID("1.3.6.3"))

	names := Names()

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
func ExampleNames() {
	ResetRegistry()

	MustRegister("first", MustParseOID("1.3.6.1"))
	MustRegister("second", MustParseOID("2.100.3"))

	names := Names()

	fmt.Println(len(names))
	// Output: 2
}

// Бенчмарк
func BenchmarkGlobalNames(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = Names()
	}
}

func TestGlobalOIDs(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	MustRegister("first", oid1)
	MustRegister("second", oid2)
	MustRegister("third", oid3)

	t.Run("Содержит все OID", func(t *testing.T) {
		oids := OIDs()

		if len(oids) != 3 {
			t.Errorf("len = %d, want 3", len(oids))
		}

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
		ResetRegistry()

		oids := OIDs()

		if len(oids) != 0 {
			t.Errorf("len = %d, want 0", len(oids))
		}
	})
}

func TestGlobalOIDsDeepCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	oids := OIDs()

	// Изменяем OID в слайсе
	oids[0][0] = 99

	// Реестр не должен измениться
	if o, _ := LookupByName("test"); o[0] != 1 {
		t.Error("OIDs должен вернуть глубокую копию")
	}
}

func TestGlobalOIDsEachCallNewSlice(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	oids1 := OIDs()
	oids2 := OIDs()

	oids1[0][0] = 99

	if oids2[0][0] != 1 {
		t.Error("Каждый вызов должен возвращать новый слайс")
	}
}

func TestGlobalOIDsNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	oids := OIDs()
	after := Size()

	if before != after {
		t.Error("OIDs не должен изменять реестр")
	}

	// Изменяем OID в слайсе
	oids[0][0] = 99

	// Реестр не должен измениться
	if o, _ := LookupByName("test"); o[0] != 1 {
		t.Error("Реестр не должен измениться")
	}
}

func TestGlobalOIDsAfterRemove(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

	Remove("first")

	oids := OIDs()

	if len(oids) != 1 {
		t.Errorf("len = %d, want 1", len(oids))
	}

	if !oids[0].Equal(oid2) {
		t.Errorf("oids[0] = %v, want %v", oids[0], oid2)
	}
}

func TestGlobalOIDsAfterClear(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Clear()

	oids := OIDs()

	if len(oids) != 0 {
		t.Errorf("len = %d, want 0", len(oids))
	}
}

// Пример использования
func ExampleOIDs() {
	ResetRegistry()

	MustRegister("first", MustParseOID("1.3.6.1"))
	MustRegister("second", MustParseOID("2.100.3"))

	oids := OIDs()

	fmt.Println(len(oids))
	// Output: 2
}

// Бенчмарк
func BenchmarkGlobalOIDs(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = OIDs()
	}
}

func TestGlobalOIDsNoCopy(t *testing.T) {
	ResetRegistry()

	// Регистрируем OID
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	MustRegister("first", oid1)
	MustRegister("second", oid2)
	MustRegister("third", oid3)

	t.Run("Содержит все OID", func(t *testing.T) {
		oids := OIDsNoCopy()

		if len(oids) != 3 {
			t.Errorf("len = %d, want 3", len(oids))
		}

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
		ResetRegistry()

		oids := OIDsNoCopy()

		if len(oids) != 0 {
			t.Errorf("len = %d, want 0", len(oids))
		}
	})
}

func TestGlobalOIDsNoCopySharedSlices(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	oids := OIDsNoCopy()

	// Изменяем OID в слайсе
	oids[0][0] = 99

	// Реестр должен измениться (общий слайс)
	if o, _ := LookupByNameNoCopy("test"); o[0] != 99 {
		t.Error("OIDsNoCopy должен вернуть ссылки")
	}

	// Восстанавливаем
	oids[0][0] = 1
}

func TestGlobalOIDsNoCopyEachCallNewSlice(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	oids1 := OIDsNoCopy()
	oids2 := OIDsNoCopy()

	// Добавляем в первый слайс (не влияет на реестр)
	oids1 = append(oids1, MustParseOID("2.100.3"))

	if len(oids2) != 1 {
		t.Error("Каждый вызов должен возвращать новый слайс")
	}
}

func TestGlobalOIDsNoCopyNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	oids := OIDsNoCopy()
	after := Size()

	if before != after {
		t.Error("OIDsNoCopy не должен изменять реестр")
	}

	// Добавляем в слайс (не влияет на реестр)
	oids = append(oids, MustParseOID("2.100.3"))

	if Size() != 1 {
		t.Error("Реестр не должен измениться")
	}
}

func TestGlobalOIDsVsOIDsNoCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	// OIDs - глубокая копия
	oids := OIDs()
	oids[0][0] = 99
	if o, _ := LookupByName("test"); o[0] != 1 {
		t.Error("OIDs должен вернуть копию")
	}

	// OIDsNoCopy - ссылка
	oidsNoCopy := OIDsNoCopy()
	oidsNoCopy[0][0] = 99
	if o, _ := LookupByNameNoCopy("test"); o[0] != 99 {
		t.Error("OIDsNoCopy должен вернуть ссылку")
	}

	// Восстанавливаем
	oidsNoCopy[0][0] = 1
}

func TestGlobalOIDsNoCopyAfterRemove(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

	Remove("first")

	oids := OIDsNoCopy()

	if len(oids) != 1 {
		t.Errorf("len = %d, want 1", len(oids))
	}

	if !oids[0].Equal(oid2) {
		t.Errorf("oids[0] = %v, want %v", oids[0], oid2)
	}
}

// Пример использования
func ExampleOIDsNoCopy() {
	ResetRegistry()

	MustRegister("first", MustParseOID("1.3.6.1"))
	MustRegister("second", MustParseOID("2.100.3"))

	oids := OIDsNoCopy()

	fmt.Println(len(oids))
	// Output: 2
}

// Бенчмарк
func BenchmarkGlobalOIDsNoCopy(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = OIDsNoCopy()
	}
}

// Сравнение OIDs vs OIDsNoCopy
func BenchmarkGlobalOIDsVsNoCopy(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	b.Run("OIDs", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = OIDs()
		}
	})

	b.Run("OIDsNoCopy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = OIDsNoCopy()
		}
	})
}

func TestGetRegistry(t *testing.T) {
	ResetRegistry()

	t.Run("Возвращает не nil", func(t *testing.T) {
		reg := GetRegistry()

		if reg == nil {
			t.Fatal("GetRegistry: nil")
		}
	})

	t.Run("Возвращает тот же реестр", func(t *testing.T) {
		reg1 := GetRegistry()
		reg2 := GetRegistry()

		if reg1 != reg2 {
			t.Error("GetRegistry должен возвращать один и тот же реестр")
		}
	})

	t.Run("Возвращает реестр с данными", func(t *testing.T) {
		oid := MustParseOID("1.3.6.1")
		MustRegister("test", oid)

		reg := GetRegistry()

		if reg.Size() != 1 {
			t.Errorf("Size = %d, want 1", reg.Size())
		}
	})
}

func TestGetRegistryDirectAccess(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")

	// Регистрируем через глобальную функцию
	MustRegister("test", oid)

	// Получаем прямой доступ
	reg := GetRegistry()

	// Проверяем через прямой доступ
	if !reg.Contains("test") {
		t.Error("Contains должен вернуть true")
	}

	got, exists := reg.LookupByName("test")
	if !exists || !got.Equal(oid) {
		t.Error("LookupByName вернул неверный результат")
	}
}

func TestGetRegistryModification(t *testing.T) {
	ResetRegistry()

	reg := GetRegistry()

	// Регистрируем через прямой доступ
	oid := MustParseOID("1.3.6.1")
	if err := reg.Register("direct", oid); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Проверяем через глобальные функции
	if !Contains("direct") {
		t.Error("Contains должен вернуть true")
	}

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}
}

func TestGetRegistryAfterReset(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	oldReg := GetRegistry()

	// Сбрасываем
	ResetRegistry()

	newReg := GetRegistry()

	if oldReg == newReg {
		t.Error("После Reset должен быть новый реестр")
	}

	if newReg.Size() != 0 {
		t.Errorf("Size = %d, want 0", newReg.Size())
	}
}

func TestGetRegistryConsistency(t *testing.T) {
	ResetRegistry()

	reg := GetRegistry()

	// Регистрируем через глобальную функцию
	oid1 := MustParseOID("1.3.6.1")
	MustRegister("first", oid1)

	// Регистрируем через прямой доступ
	oid2 := MustParseOID("2.100.3")
	reg.Register("second", oid2)

	// Оба должны быть видны
	if !Contains("first") {
		t.Error("first не найден")
	}
	if !Contains("second") {
		t.Error("second не найден")
	}

	if Size() != 2 {
		t.Errorf("Size = %d, want 2", Size())
	}
}

// Пример использования
func ExampleGetRegistry() {
	ResetRegistry()

	MustRegister("test", MustParseOID("1.3.6.1"))

	reg := GetRegistry()

	fmt.Println(reg.Size())
	// Output: 1
}

// Бенчмарк
func BenchmarkGetRegistry(b *testing.B) {
	ResetRegistry()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = GetRegistry()
	}
}

func TestResetRegistry(t *testing.T) {
	t.Run("Очищает все записи", func(t *testing.T) {
		ResetRegistry()

		// Регистрируем OID
		oid1 := MustParseOID("1.3.6.1")
		oid2 := MustParseOID("2.100.3")

		MustRegister("first", oid1)
		MustRegister("second", oid2)

		if Size() != 2 {
			t.Errorf("Size = %d, want 2", Size())
		}

		// Сбрасываем
		ResetRegistry()

		if Size() != 0 {
			t.Errorf("Size после Reset = %d, want 0", Size())
		}
	})

	t.Run("Очищает пустой реестр", func(t *testing.T) {
		ResetRegistry()

		// Сбрасываем пустой реестр
		ResetRegistry()

		if Size() != 0 {
			t.Errorf("Size = %d, want 0", Size())
		}
	})
}

func TestResetRegistryCreatesNewInstance(t *testing.T) {
	ResetRegistry()

	oldReg := GetRegistry()

	// Сбрасываем
	ResetRegistry()

	newReg := GetRegistry()

	if oldReg == newReg {
		t.Error("ResetRegistry должен создавать новый реестр")
	}
}

func TestResetRegistryThenRegister(t *testing.T) {
	ResetRegistry()

	// Регистрируем
	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	// Сбрасываем
	ResetRegistry()

	// Можно регистрировать снова
	if err := Register("test", oid); err != nil {
		t.Errorf("Register после Reset: %v", err)
	}

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}
}

func TestResetRegistryThenLookup(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	// Сбрасываем
	ResetRegistry()

	// Не можем найти
	if _, exists := LookupByName("test"); exists {
		t.Error("LookupByName должен вернуть false после Reset")
	}

	if _, exists := LookupByOID(oid); exists {
		t.Error("LookupByOID должен вернуть false после Reset")
	}

	if Contains("test") {
		t.Error("Contains должен вернуть false после Reset")
	}
}

func TestResetRegistryThenBatchRegister(t *testing.T) {
	ResetRegistry()

	// Регистрируем
	entries := map[string]OID{
		"first": MustParseOID("1.3.6.1"),
	}
	MustBatchRegister(entries)

	// Сбрасываем
	ResetRegistry()

	// Можно регистрировать снова
	if err := BatchRegister(entries); err != nil {
		t.Errorf("BatchRegister после Reset: %v", err)
	}

	if Size() != 1 {
		t.Errorf("Size = %d, want 1", Size())
	}
}

func TestResetRegistryMultipleTimes(t *testing.T) {
	// Многократный сброс
	for i := 0; i < 5; i++ {
		ResetRegistry()

		oid := MustParseOID("1.3.6.1")
		MustRegister("test", oid)

		if Size() != 1 {
			t.Errorf("Итерация %d: Size = %d, want 1", i, Size())
		}
	}

	// Финальный сброс
	ResetRegistry()

	if Size() != 0 {
		t.Errorf("Size = %d, want 0", Size())
	}
}

func TestResetRegistryIsolation(t *testing.T) {
	ResetRegistry()

	// Регистрируем
	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	// Сохраняем старый реестр
	oldReg := GetRegistry()

	// Сбрасываем
	ResetRegistry()

	// Старый реестр не должен измениться
	if oldReg.Size() != 1 {
		t.Errorf("oldReg.Size = %d, want 1", oldReg.Size())
	}

	// Новый реестр пустой
	if Size() != 0 {
		t.Errorf("Size = %d, want 0", Size())
	}
}

// Пример использования
func ExampleResetRegistry() {
	ResetRegistry()

	MustRegister("test", MustParseOID("1.3.6.1"))

	fmt.Println(Size())

	ResetRegistry()

	fmt.Println(Size())
	// Output:
	// 1
	// 0
}

// Бенчмарк
func BenchmarkResetRegistry(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ResetRegistry()
	}
}

func TestSnapshot(t *testing.T) {
	ResetRegistry()

	t.Run("Пустой реестр", func(t *testing.T) {
		snapshot := Snapshot()

		if len(snapshot) != 0 {
			t.Errorf("len = %d, want 0", len(snapshot))
		}
	})

	t.Run("С записями", func(t *testing.T) {
		oid1 := MustParseOID("1.3.6.1")
		oid2 := MustParseOID("2.100.3")

		MustRegister("first", oid1)
		MustRegister("second", oid2)

		snapshot := Snapshot()

		if len(snapshot) != 2 {
			t.Errorf("len = %d, want 2", len(snapshot))
		}

		if !snapshot["first"].Equal(oid1) {
			t.Error("snapshot[first] неверный")
		}
		if !snapshot["second"].Equal(oid2) {
			t.Error("snapshot[second] неверный")
		}
	})
}

func TestSnapshotDeepCopy(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	snapshot := Snapshot()

	// Изменяем OID в снимке
	snapshot["test"][0] = 99

	// Реестр не должен измениться
	if o, _ := LookupByName("test"); o[0] != 1 {
		t.Error("Snapshot должен вернуть глубокую копию")
	}
}

func TestSnapshotEachCallNewMap(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	snapshot1 := Snapshot()
	snapshot2 := Snapshot()

	snapshot1["new"] = MustParseOID("2.100.3")

	if _, exists := snapshot2["new"]; exists {
		t.Error("Каждый вызов должен возвращать новый map")
	}
}

func TestSnapshotNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	before := Size()
	snapshot := Snapshot()
	after := Size()

	if before != after {
		t.Error("Snapshot не должен изменять реестр")
	}

	// Изменяем снимок
	snapshot["test"][0] = 99
	snapshot["new"] = MustParseOID("2.100.3")

	// Реестр не должен измениться
	if Size() != 1 {
		t.Error("Реестр не должен измениться")
	}
}

func TestSnapshotAfterRemove(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

	Remove("first")

	snapshot := Snapshot()

	if len(snapshot) != 1 {
		t.Errorf("len = %d, want 1", len(snapshot))
	}

	if !snapshot["second"].Equal(oid2) {
		t.Error("snapshot[second] неверный")
	}
}

func TestSnapshotAfterClear(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	Clear()

	snapshot := Snapshot()

	if len(snapshot) != 0 {
		t.Errorf("len = %d, want 0", len(snapshot))
	}
}

func TestSnapshotConsistency(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	// Snapshot должен совпадать с List
	snapshot := Snapshot()
	list := List()

	if len(snapshot) != len(list) {
		t.Error("Snapshot и List должны совпадать")
	}

	for name, oid := range snapshot {
		if !list[name].Equal(oid) {
			t.Errorf("snapshot[%s] != list[%s]", name, name)
		}
	}
}

// Пример использования
func ExampleSnapshot() {
	ResetRegistry()

	MustRegister("test", MustParseOID("1.3.6.1"))

	snapshot := Snapshot()

	fmt.Println(len(snapshot))
	// Output: 1
}

// Бенчмарк
func BenchmarkSnapshot(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = Snapshot()
	}
}

func TestDiffEmpty(t *testing.T) {
	ResetRegistry()

	snapshot := Snapshot()

	added, removed, changed := Diff(snapshot)

	if len(added) != 0 {
		t.Errorf("added = %d, want 0", len(added))
	}
	if len(removed) != 0 {
		t.Errorf("removed = %d, want 0", len(removed))
	}
	if len(changed) != 0 {
		t.Errorf("changed = %d, want 0", len(changed))
	}
}

func TestDiffAdded(t *testing.T) {
	ResetRegistry()

	// Пустой снимок
	snapshot := Snapshot()

	// Добавляем записи
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

	added, removed, changed := Diff(snapshot)

	if len(added) != 2 {
		t.Errorf("added = %d, want 2", len(added))
	}
	if len(removed) != 0 {
		t.Errorf("removed = %d, want 0", len(removed))
	}
	if len(changed) != 0 {
		t.Errorf("changed = %d, want 0", len(changed))
	}

	if !added["first"].Equal(oid1) {
		t.Error("added[first] неверный")
	}
	if !added["second"].Equal(oid2) {
		t.Error("added[second] неверный")
	}
}

func TestDiffRemoved(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")

	MustRegister("first", oid1)
	MustRegister("second", oid2)

	// Снимок с двумя записями
	snapshot := Snapshot()

	// Удаляем одну
	Remove("first")

	added, removed, changed := Diff(snapshot)

	if len(added) != 0 {
		t.Errorf("added = %d, want 0", len(added))
	}
	if len(removed) != 1 {
		t.Errorf("removed = %d, want 1", len(removed))
	}
	if len(changed) != 0 {
		t.Errorf("changed = %d, want 0", len(changed))
	}

	if !removed["first"].Equal(oid1) {
		t.Error("removed[first] неверный")
	}
}

func TestDiffChanged(t *testing.T) {
	ResetRegistry()

	oid1 := MustParseOID("1.3.6.1")
	MustRegister("test", oid1)

	// Снимок с оригинальным OID
	snapshot := Snapshot()

	// Перезаписываем другим OID
	oid2 := MustParseOID("1.3.6.2")
	MustRegister("test", oid2)

	added, removed, changed := Diff(snapshot)

	if len(added) != 0 {
		t.Errorf("added = %d, want 0", len(added))
	}
	if len(removed) != 0 {
		t.Errorf("removed = %d, want 0", len(removed))
	}
	if len(changed) != 1 {
		t.Errorf("changed = %d, want 1", len(changed))
	}

	if !changed["test"].Equal(oid2) {
		t.Error("changed[test] неверный")
	}
}

func TestDiffMixed(t *testing.T) {
	ResetRegistry()

	// Изначальное состояние
	oid1 := MustParseOID("1.3.6.1")
	oid2 := MustParseOID("2.100.3")
	oid3 := MustParseOID("0.39.1")

	MustRegister("first", oid1)
	MustRegister("second", oid2)
	MustRegister("third", oid3)

	snapshot := Snapshot()

	// Изменяем: first удален, second изменен, fourth добавлен
	Remove("first")
	MustRegister("second", MustParseOID("2.100.4"))
	MustRegister("fourth", MustParseOID("1.3.6.4"))

	added, removed, changed := Diff(snapshot)

	if len(added) != 1 {
		t.Errorf("added = %d, want 1", len(added))
	}
	if len(removed) != 1 {
		t.Errorf("removed = %d, want 1", len(removed))
	}
	if len(changed) != 1 {
		t.Errorf("changed = %d, want 1", len(changed))
	}

	if !added["fourth"].Equal(MustParseOID("1.3.6.4")) {
		t.Error("added[fourth] неверный")
	}
	if !removed["first"].Equal(oid1) {
		t.Error("removed[first] неверный")
	}
	if !changed["second"].Equal(MustParseOID("2.100.4")) {
		t.Error("changed[second] неверный")
	}
}

func TestDiffNotModifyRegistry(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	snapshot := Snapshot()

	before := Size()
	Diff(snapshot)
	after := Size()

	if before != after {
		t.Error("Diff не должен изменять реестр")
	}
}

func TestDiffNilSnapshot(t *testing.T) {
	ResetRegistry()

	oid := MustParseOID("1.3.6.1")
	MustRegister("test", oid)

	added, removed, changed := Diff(nil)

	if len(added) != 1 {
		t.Errorf("added = %d, want 1", len(added))
	}
	if len(removed) != 0 {
		t.Errorf("removed = %d, want 0", len(removed))
	}
	if len(changed) != 0 {
		t.Errorf("changed = %d, want 0", len(changed))
	}
}

// Пример использования
func ExampleDiff() {
	ResetRegistry()

	// Снимок пустого реестра
	snapshot := Snapshot()

	// Добавляем запись
	MustRegister("test", MustParseOID("1.3.6.1"))

	added, removed, changed := Diff(snapshot)

	fmt.Println(len(added))
	fmt.Println(len(removed))
	fmt.Println(len(changed))
	// Output:
	// 1
	// 0
	// 0
}

// Бенчмарк
func BenchmarkDiff(b *testing.B) {
	ResetRegistry()
	for i := 0; i < 10; i++ {
		oid := MustParseOID(fmt.Sprintf("1.3.6.%d", i+1))
		MustRegister(fmt.Sprintf("oid-%d", i), oid)
	}

	snapshot := Snapshot()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _, _ = Diff(snapshot)
	}
}
