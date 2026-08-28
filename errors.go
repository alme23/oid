package oid

import "errors"

// Статические ошибки пакета
var (
	// ErrEmptyOID - пустой OID
	ErrEmptyOID = errors.New("пустой OID")

	// ErrInvalidOID - некорректный OID
	ErrInvalidOID = errors.New("некорректный OID")

	// ErrOIDTooShort - OID содержит менее 2 компонентов
	ErrOIDTooShort = errors.New("OID должен содержать минимум 2 компонента")

	// ErrFirstComponentTooBig - первый компонент > 2
	ErrFirstComponentTooBig = errors.New("первый компонент OID должен быть 0, 1 или 2")

	// ErrSecondComponentTooBig - второй компонент > 39 при первом 0 или 1
	ErrSecondComponentTooBig = errors.New("второй компонент должен быть <= 39 при первом компоненте 0 или 1")

	// ErrComponentTooBig - компонент превышает максимальное значение ASN.1
	ErrComponentTooBig = errors.New("компонент превышает максимальное значение ASN.1")

	// ErrNoParent - у OID нет родителя
	ErrNoParent = errors.New("у OID нет родителя")

	// ErrDataTooShort - данные слишком короткие
	ErrDataTooShort = errors.New("данные слишком короткие для OID")

	// ErrInvalidASN1Tag - некорректный тег ASN.1
	ErrInvalidASN1Tag = errors.New("некорректный тег ASN.1")

	// ErrInvalidASN1Length - некорректная длина ASN.1
	ErrInvalidASN1Length = errors.New("некорректная длина ASN.1")

	// ErrInsufficientData - недостаточно данных
	ErrInsufficientData = errors.New("недостаточно данных для OID")

	// ErrInvalidFirstComponent - некорректный первый компонент
	ErrInvalidFirstComponent = errors.New("некорректный первый компонент OID")

	// ErrInvalidComponent - некорректный компонент OID
	ErrInvalidComponent = errors.New("некорректный компонент OID")

	// ErrInvalidJSONType - некорректный JSON тип
	ErrInvalidJSONType = errors.New("некорректный JSON тип для OID")
)
