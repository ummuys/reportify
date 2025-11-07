package errs

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Common normalized PostgreSQL / pgx error types.
var (
	// === Logical & data-level errors ===
	ErrNotFound             = errors.New("not_found")
	ErrDuplicate            = errors.New("duplicate")
	ErrForeignKey           = errors.New("foreign_key_violation")
	ErrInvalidInput         = errors.New("invalid_input")
	ErrConstraint           = errors.New("constraint_violation")
	ErrNullViolation        = errors.New("not_null_violation")
	ErrCheckViolation       = errors.New("check_violation")
	ErrExclusionViolation   = errors.New("exclusion_violation")
	ErrSerializationFailure = errors.New("serialization_failure")
	ErrDeadlock             = errors.New("deadlock_detected")

	// === Data / formatting ===
	ErrNumericOutOfRange  = errors.New("numeric_value_out_of_range")
	ErrInvalidTextFormat  = errors.New("invalid_text_representation")
	ErrStringTooLong      = errors.New("string_data_right_truncation")
	ErrInvalidDatetime    = errors.New("invalid_datetime_format")
	ErrDivisionByZero     = errors.New("division_by_zero")
	ErrUntranslatableChar = errors.New("untranslatable_character")

	// === Access / privileges ===
	ErrInsufficientPrivilege = errors.New("insufficient_privilege")
	ErrInvalidPassword       = errors.New("invalid_password")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrUndefinedTable        = errors.New("undefined_table")
	ErrUndefinedColumn       = errors.New("undefined_column")

	// === Transaction & connection ===
	ErrConnection     = errors.New("connection_failure")
	ErrTxState        = errors.New("invalid_transaction_state")
	ErrInFailedTx     = errors.New("in_failed_sql_transaction")
	ErrIdleTxTimeout  = errors.New("idle_in_transaction_timeout")
	ErrConnectionLost = errors.New("connection_lost")

	// === Resource / system ===
	ErrOutOfMemory           = errors.New("out_of_memory")
	ErrDiskFull              = errors.New("disk_full")
	ErrTooManyConnections    = errors.New("too_many_connections")
	ErrSystemIO              = errors.New("io_error")
	ErrSystemInternal        = errors.New("system_error")
	ErrDataCorrupted         = errors.New("data_corrupted")
	ErrIndexCorrupted        = errors.New("index_corrupted")
	ErrConfigurationExceeded = errors.New("configuration_limit_exceeded")

	// === Syntax / schema ===
	ErrSyntaxError      = errors.New("syntax_error")
	ErrDatatypeMismatch = errors.New("datatype_mismatch")
	ErrDuplicateObject  = errors.New("duplicate_object")
	ErrInvalidSchema    = errors.New("invalid_schema_definition")
	ErrInvalidTableDef  = errors.New("invalid_table_definition")

	// === Operator / cancel ===
	ErrQueryCanceled   = errors.New("query_canceled")
	ErrAdminShutdown   = errors.New("admin_shutdown")
	ErrCrashShutdown   = errors.New("crash_shutdown")
	ErrDatabaseDropped = errors.New("database_dropped")

	// === Internal fallback ===
	ErrInternal = errors.New("internal_error")
	ErrUnknown  = errors.New("unknown")
)

func ParsePgError(err error) error {
	if err == nil {
		return nil
	}

	// 1. Check for logical pgx errors
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if errors.Is(err, pgx.ErrTxClosed) {
		return ErrTxState
	}

	// 2. Try PostgreSQL-specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {

		// === Class 23 — Integrity Constraint Violation ===
		case "23505":
			return ErrDuplicate
		case "23503":
			return ErrForeignKey
		case "23502":
			return ErrNullViolation
		case "23514":
			return ErrCheckViolation
		case "23P01":
			return ErrExclusionViolation

		// === Class 22 — Data Exception ===
		case "22P02":
			return ErrInvalidInput
		case "22003":
			return ErrNumericOutOfRange
		case "22001":
			return ErrStringTooLong
		case "22007":
			return ErrInvalidDatetime
		case "22012":
			return ErrDivisionByZero
		case "22P05":
			return ErrUntranslatableChar

		// === Class 08 — Connection Exception ===
		case "08006", "08001", "08003":
			return ErrConnection
		case "57P03":
			return ErrConnectionLost

		// === Class 25 — Invalid Transaction State ===
		case "25000":
			return ErrTxState
		case "25P02":
			return ErrInFailedTx
		case "25P03":
			return ErrIdleTxTimeout

		// === Class 40 — Transaction Rollback ===
		case "40001":
			return ErrSerializationFailure
		case "40P01":
			return ErrDeadlock

		// === Class 42 — Syntax Error or Access Rule Violation ===
		case "42601":
			return ErrSyntaxError
		case "42804":
			return ErrDatatypeMismatch
		case "42701", "42702", "42703", "42704":
			return ErrDuplicateObject
		case "42P06":
			return ErrInvalidSchema
		case "42P17":
			return ErrInvalidTableDef
		case "42501":
			return ErrInsufficientPrivilege
		case "28P01":
			return ErrInvalidPassword

		// === Class 53 — Resource Issues ===
		case "53100":
			return ErrDiskFull
		case "53200":
			return ErrOutOfMemory
		case "53300":
			return ErrTooManyConnections
		case "53400":
			return ErrConfigurationExceeded

		// === Class 55 — Object State ===
		case "55P03":
			return ErrSystemIO
		case "55000":
			return ErrSystemInternal

		// === Class 57 — Operator Intervention ===
		case "57014":
			return ErrQueryCanceled
		case "57P01":
			return ErrAdminShutdown
		case "57P02":
			return ErrCrashShutdown
		case "57P04":
			return ErrDatabaseDropped

		// === Class 58 — System Errors ===
		case "58030":
			return ErrSystemIO
		case "XX000":
			return ErrInternal
		case "XX001":
			return ErrDataCorrupted
		case "XX002":
			return ErrIndexCorrupted

		default:
			return ErrUnknown
		}
	}

	// 3. Fallback for unrecognized errors
	return ErrUnknown
}
