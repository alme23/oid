package oid

import "errors"

// Статические ошибки пакета
var (
	// Основные ошибки OID
	ErrEmptyOID              = errors.New("пустой OID")
	ErrInvalidOID            = errors.New("некорректный OID")
	ErrOIDTooShort           = errors.New("OID должен содержать минимум 2 компонента")
	ErrFirstComponentTooBig  = errors.New("первый компонент OID должен быть 0, 1 или 2")
	ErrSecondComponentTooBig = errors.New("второй компонент должен быть <= 39 при первом компоненте 0 или 1")
	ErrComponentTooBig       = errors.New("компонент превышает максимальное значение ASN.1")
	ErrNoParent              = errors.New("у OID нет родителя")
	ErrDataTooShort          = errors.New("данные слишком короткие для OID")
	ErrInvalidASN1Tag        = errors.New("некорректный тег ASN.1")
	ErrInvalidASN1Length     = errors.New("некорректная длина ASN.1")
	ErrInsufficientData      = errors.New("недостаточно данных для OID")
	ErrInvalidFirstComponent = errors.New("некорректный первый компонент OID")
	ErrInvalidComponent      = errors.New("некорректный компонент OID")
	ErrInvalidJSONType       = errors.New("некорректный JSON тип для OID")

	// Ошибки BER
	ErrInvalidLength        = errors.New("некорректная кодировка длины в BER")
	ErrEmptyContent         = errors.New("пустой контент BER OID")
	ErrFirstComponentFailed = errors.New("некорректный первый компонент BER OID")
	ErrComponentFailed      = errors.New("некорректный компонент BER OID")
	ErrComponentOverflow    = errors.New("переполнение компонента OID при декодировании BER")

	// Ошибки Registry
	ErrOIDAlreadyRegistered = errors.New("OID уже зарегистрирован")
	ErrNameAlreadyExists    = errors.New("имя уже зарегистрировано")
	ErrDuplicateNameInBatch = errors.New("дублирование имени внутри пакета")
	ErrDuplicateOIDInBatch  = errors.New("дублирование OID внутри пакета")

	// Ошибки Database
	ErrUnsupportedScanType = errors.New("неподдерживаемый тип для конвертации в OID")
	ErrInvalidArrayFormat  = errors.New("некорректный формат массива OID")
	ErrSaveValidation      = errors.New("невалидный OID перед сохранением в БД")
	ErrDatabaseParse       = errors.New("ошибка парсинга OID из БД")
	ErrInvalidArrayOID     = errors.New("невалидный OID в массиве")
	ErrJSONDecodeArray     = errors.New("ошибка декодирования JSON массива")

	// Ошибки ColumnarOID
	ErrOIDNotInBase        = errors.New("OID не принадлежит базе")
	ErrNotEnoughComponents = errors.New("недостаточно компонентов для колумнарного OID")
	ErrNoIndexes           = errors.New("нет индексов")
	ErrColumnNotFound      = errors.New("колонка не найдена")

	// Ошибки TableOID
	ErrTableEmpty   = errors.New("таблица не содержит колонок")
	ErrNotInTable   = errors.New("OID не принадлежит таблице")
	ErrInvalidTable = errors.New("некорректная таблица")
)
