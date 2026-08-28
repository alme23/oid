package oid

import (
	"fmt"
)

// defaultRegistry — глобальный синглтон реестра.
// Используйте с осторожностью в конкурентной среде.
var defaultRegistry = NewRegistry()

// Register регистрирует OID в глобальном реестре.
// Возвращает ошибку, если OID уже зарегистрирован под другим именем.
func Register(name string, oid OID) error {
	return defaultRegistry.Register(name, oid)
}

// MustRegister регистрирует OID и паникует при ошибке.
// Удобно для инициализации в init() или в начале main().
func MustRegister(name string, oid OID) {
	if err := Register(name, oid); err != nil {
		panic(fmt.Sprintf("ошибка регистрации OID '%s': %v", name, err))
	}
}

// BatchRegister регистрирует несколько OID атомарно.
// Если возникает любой конфликт, изменения не применяются.
func BatchRegister(entries map[string]OID) error {
	return defaultRegistry.BatchRegister(entries)
}

// MustBatchRegister регистрирует несколько OID и паникует при ошибке.
func MustBatchRegister(entries map[string]OID) {
	if err := BatchRegister(entries); err != nil {
		panic(fmt.Sprintf("ошибка пакетной регистрации OID: %v", err))
	}
}

// LookupByName возвращает копию OID по имени.
// Безопасно для изменения возвращенного OID.
func LookupByName(name string) (OID, bool) {
	return defaultRegistry.LookupByName(name)
}

// LookupByNameNoCopy возвращает прямую ссылку на OID без копирования.
// ВАЖНО: Не изменяйте возвращенный OID!
// Используйте только для чтения в сценариях, где реестр не изменяется конкурентно.
func LookupByNameNoCopy(name string) (OID, bool) {
	return defaultRegistry.LookupByNameNoCopy(name)
}

// LookupByOID возвращает имя по OID.
// Строки в Go иммутабельны, поэтому возвращаемое имя безопасно.
func LookupByOID(oid OID) (string, bool) {
	return defaultRegistry.LookupByOID(oid)
}

// Remove удаляет запись по имени.
// Возвращает true, если запись была удалена.
func Remove(name string) bool {
	return defaultRegistry.Remove(name)
}

// List возвращает глубокую копию всех записей.
// Безопасно для изменения возвращенных OID.
func List() map[string]OID {
	return defaultRegistry.List()
}

// ListNoCopy возвращает карту ссылок на OID без копирования.
// ВАЖНО: Не изменяйте возвращенные OID!
// Используйте только для чтения.
func ListNoCopy() map[string]OID {
	return defaultRegistry.ListNoCopy()
}

// Size возвращает количество элементов в реестре.
func Size() int {
	return defaultRegistry.Size()
}

// Contains проверяет существование имени в реестре.
// Не выполняет аллокаций.
func Contains(name string) bool {
	return defaultRegistry.Contains(name)
}

// Clear полностью очищает реестр.
func Clear() {
	defaultRegistry.Clear()
}

// Names возвращает все имена в реестре.
// Строки иммутабельны, поэтому возвращаемый срез безопасен.
func Names() []string {
	return defaultRegistry.Names()
}

// OIDs возвращает глубокую копию всех OID в реестре.
// Безопасно для изменения возвращенных OID.
func OIDs() []OID {
	return defaultRegistry.OIDs()
}

// OIDsNoCopy возвращает срез ссылок на все OID без копирования.
// ВАЖНО: Не изменяйте возвращенные OID!
// Используйте только для чтения.
func OIDsNoCopy() []OID {
	return defaultRegistry.OIDsNoCopy()
}

// GetRegistry возвращает глобальный реестр.
// Используйте с осторожностью: прямой доступ к внутреннему состоянию.
func GetRegistry() *Registry {
	return defaultRegistry
}

// ResetRegistry сбрасывает глобальный реестр к пустому состоянию.
// Полезно для тестов и изоляции.
func ResetRegistry() {
	defaultRegistry = NewRegistry()
}

// Snapshot возвращает снимок реестра (глубокая копия).
// Полезно для отладки и мониторинга.
func Snapshot() map[string]OID {
	return defaultRegistry.List()
}

// Diff возвращает разницу между текущим реестром и снимком.
// Возвращает: добавленные, удаленные, измененные OID.
func Diff(snapshot map[string]OID) (added, removed, changed map[string]OID) {
	current := List()

	added = make(map[string]OID)
	removed = make(map[string]OID)
	changed = make(map[string]OID)

	// Используем maps.Copy для эффективного копирования
	for name, oid := range current {
		if oldOID, exists := snapshot[name]; !exists {
			added[name] = oid
		} else if !oldOID.Equal(oid) {
			changed[name] = oid
		}
	}

	for name := range snapshot {
		if _, exists := current[name]; !exists {
			removed[name] = snapshot[name]
		}
	}

	return
}
