// oid/global.go
package oid

import (
	"fmt"
)

// defaultRegistry — глобальный синглтон реестра
var defaultRegistry = NewRegistry()

// Register регистрирует OID в глобальном реестре
func Register(name string, oid OID) error {
	return defaultRegistry.Register(name, oid)
}

// MustRegister регистрирует OID и паникует при ошибке
func MustRegister(name string, oid OID) {
	if err := Register(name, oid); err != nil {
		panic(fmt.Sprintf("ошибка регистрации OID '%s': %v", name, err))
	}
}

// BatchRegister регистрирует несколько OID атомарно
func BatchRegister(entries map[string]OID) error {
	return defaultRegistry.BatchRegister(entries)
}

// MustBatchRegister регистрирует несколько OID и паникует при ошибке
func MustBatchRegister(entries map[string]OID) {
	if err := BatchRegister(entries); err != nil {
		panic(fmt.Sprintf("ошибка пакетной регистрации OID: %v", err))
	}
}

// LookupByName возвращает OID по имени
func LookupByName(name string) (OID, bool) {
	return defaultRegistry.LookupByName(name)
}

// LookupByOID возвращает имя по OID
func LookupByOID(oid OID) (string, bool) {
	return defaultRegistry.LookupByOID(oid)
}

// Remove удаляет запись по имени
func Remove(name string) bool {
	return defaultRegistry.Remove(name)
}

// List возвращает копию всех записей
func List() map[string]OID {
	return defaultRegistry.List()
}

// Size возвращает количество элементов
func Size() int {
	return defaultRegistry.Size()
}

// Contains проверяет существование имени
func Contains(name string) bool {
	return defaultRegistry.Contains(name)
}

// Clear полностью очищает реестр
func Clear() {
	defaultRegistry.Clear()
}

// Names возвращает все имена
func Names() []string {
	return defaultRegistry.Names()
}

// OIDs возвращает все OID
func OIDs() []OID {
	return defaultRegistry.OIDs()
}

// GetRegistry возвращает глобальный реестр
func GetRegistry() *Registry {
	return defaultRegistry
}

// ResetRegistry сбрасывает глобальный реестр
func ResetRegistry() {
	defaultRegistry = NewRegistry()
}

// Snapshot возвращает снимок реестра
func Snapshot() map[string]OID {
	return defaultRegistry.List()
}

// Diff возвращает разницу между реестром и снимком
func Diff(snapshot map[string]OID) (added, removed, changed map[string]OID) {
	current := List()

	added = make(map[string]OID)
	removed = make(map[string]OID)
	changed = make(map[string]OID)

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

// LookupByNameNoCopy возвращает прямую ссылку на OID из глобального реестра без аллокаций.
func LookupByNameNoCopy(name string) (OID, bool) {
	return defaultRegistry.LookupByNameNoCopy(name)
}

// ListNoCopy возвращает карту ссылок на OID глобального реестра без аллокаций.
func ListNoCopy() map[string]OID {
	return defaultRegistry.ListNoCopy()
}

// OIDsNoCopy возвращает срез ссылок на все OID глобального реестра без аллокаций.
func OIDsNoCopy() []OID {
	return defaultRegistry.OIDsNoCopy()
}
