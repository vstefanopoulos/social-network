package commonerrors

import (
	"errors"

	"google.golang.org/grpc/codes"
)

var (
	// ErrOK indicates successful completion.
	// This error should generally not be returned; use nil instead.
	ErrOK = errors.New("ok")

	// ErrCanceled indicates the operation was canceled by the caller.
	ErrCanceled = errors.New("canceled")

	// ErrUnknown indicates an unknown error.
	ErrUnknown = errors.New("unknown error")

	// ErrInvalidArgument indicates the client specified an invalid argument.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrDeadlineExceeded indicates the operation timed out.
	ErrDeadlineExceeded = errors.New("deadline exceeded")

	// ErrNotFound indicates a requested entity was not found.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates an attempt to create an entity that already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrPermissionDenied indicates the caller lacks permission.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrResourceExhausted indicates resource limits have been exceeded.
	ErrResourceExhausted = errors.New("resource exhausted")

	// ErrFailedPrecondition indicates the system is in an invalid state.
	ErrFailedPrecondition = errors.New("failed precondition")

	// ErrAborted indicates the operation was aborted.
	ErrAborted = errors.New("aborted")

	// ErrOutOfRange indicates a value is outside the valid range.
	ErrOutOfRange = errors.New("out of range")

	// ErrUnimplemented indicates the operation is not implemented.
	ErrUnimplemented = errors.New("unimplemented")

	// ErrInternal indicates an internal server error.
	ErrInternal = errors.New("internal error")

	// ErrUnavailable indicates the service is currently unavailable.
	ErrUnavailable = errors.New("service unavailable")

	// ErrDataLoss indicates unrecoverable data corruption or loss.
	ErrDataLoss = errors.New("data loss")

	// ErrUnauthenticated indicates missing or invalid authentication.
	ErrUnauthenticated = errors.New("unauthenticated")
)

var classToGRPC = map[error]codes.Code{
	ErrOK:                 codes.OK,
	ErrCanceled:           codes.Canceled,
	ErrUnknown:            codes.Unknown,
	ErrInvalidArgument:    codes.InvalidArgument,
	ErrDeadlineExceeded:   codes.DeadlineExceeded,
	ErrNotFound:           codes.NotFound,
	ErrAlreadyExists:      codes.AlreadyExists,
	ErrPermissionDenied:   codes.PermissionDenied,
	ErrResourceExhausted:  codes.ResourceExhausted,
	ErrFailedPrecondition: codes.FailedPrecondition,
	ErrAborted:            codes.Aborted,
	ErrOutOfRange:         codes.OutOfRange,
	ErrUnimplemented:      codes.Unimplemented,
	ErrInternal:           codes.Internal,
	ErrUnavailable:        codes.Unavailable,
	ErrDataLoss:           codes.DataLoss,
	ErrUnauthenticated:    codes.Unauthenticated,
}

var grpcToErrorClass = map[codes.Code]error{
	codes.OK:                 ErrOK,
	codes.Canceled:           ErrCanceled,
	codes.Unknown:            ErrUnknown,
	codes.InvalidArgument:    ErrInvalidArgument,
	codes.DeadlineExceeded:   ErrDeadlineExceeded,
	codes.NotFound:           ErrNotFound,
	codes.AlreadyExists:      ErrAlreadyExists,
	codes.PermissionDenied:   ErrPermissionDenied,
	codes.ResourceExhausted:  ErrResourceExhausted,
	codes.FailedPrecondition: ErrFailedPrecondition,
	codes.Aborted:            ErrAborted,
	codes.OutOfRange:         ErrOutOfRange,
	codes.Unimplemented:      ErrUnimplemented,
	codes.Internal:           ErrInternal,
	codes.Unavailable:        ErrUnavailable,
	codes.DataLoss:           ErrDataLoss,
	codes.Unauthenticated:    ErrUnauthenticated,
}
