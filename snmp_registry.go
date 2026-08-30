package oid

// SNMPRegistry хранит скалярные и колумнарные OID
type SNMPRegistry struct {
	scalars map[string]ScalarOID
	columns map[string]ColumnarOID
}

// NewSNMPRegistry создает реестр SNMP OID
func NewSNMPRegistry() *SNMPRegistry {
	return &SNMPRegistry{
		scalars: make(map[string]ScalarOID),
		columns: make(map[string]ColumnarOID),
	}
}

// RegisterScalar регистрирует скалярный OID
func (r *SNMPRegistry) RegisterScalar(name string, oid ScalarOID) {
	r.scalars[name] = oid
}

// RegisterColumn регистрирует колумнарный OID
func (r *SNMPRegistry) RegisterColumn(name string, oid ColumnarOID) {
	r.columns[name] = oid
}

// GetScalar возвращает скалярный OID по имени
func (r *SNMPRegistry) GetScalar(name string) (ScalarOID, bool) {
	oid, exists := r.scalars[name]
	return oid, exists
}

// GetColumn возвращает колумнарный OID по имени
func (r *SNMPRegistry) GetColumn(name string) (ColumnarOID, bool) {
	oid, exists := r.columns[name]
	return oid, exists
}

// GetColumnWithIndexes возвращает колумнарный OID с индексами
func (r *SNMPRegistry) GetColumnWithIndexes(name string, indexes ...uint32) (OID, bool) {
	col, exists := r.columns[name]
	if !exists {
		return nil, false
	}
	return col.WithIndexes(indexes...).FullOID(), true
}
