package kafgo

import (
	"context"
	"errors"
	"fmt"
	tele "social-network/shared/go/telemetry"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Record is a type that helps with commiting after processing a record
//
// # MAKE SURE THE PROCESSING OF THE RECORD IS INSIDE A TRANSACTION
//
// # AND AFTER PROCESSING THE RECORD MAKE SURE TO COMMIT!
//
// Usage:
//
// Record.Data() -> get your data
//
// err := Record.Commit() -> commit the result, make sure this is done
//
// # ^--- MAKE SURE TO HANDLE THE ERROR, IMPORTANT!
type Record struct {
	monotinicId   uint64
	rec           *kgo.Record
	commitChannel chan<- (*Record)
	context       context.Context
}

var ErrBadArgs = errors.New("bro, you passed bad arguments")

// newRecord creates a new Record instance
func newRecord(ctx context.Context, record *kgo.Record, commitChannel chan<- (*Record), monotonicId uint64) (*Record, error) {
	if record == nil {
		err := fmt.Errorf("%w record: %v", ErrBadArgs, record)
		tele.Error(ctx, "new record @1", "error", err.Error())
		return nil, err
	}
	return &Record{
		rec:           record,
		commitChannel: commitChannel,
		context:       ctx,
		monotinicId:   monotonicId,
	}, nil
}

// Data returns the data payload
func (r *Record) Data(ctx context.Context) []byte {
	if r.rec == nil {
		tele.Warn(ctx, "empty kafka record")
		return []byte("no data found")
	}
	return r.rec.Value
}

var ErrEmptyRecord = errors.New("empty record")

var ErrContextExpired = errors.New("context expired")

// Commit marks the record as processed in the Kafka client.
// MAKE SURE THIS IS AFTER PROCESSING THE DATA, DONT BE COMMITING THINGS YOU DIDNT PROCESS!!
func (r *Record) Commit(ctx context.Context) error {
	if r.rec == nil {
		tele.Error(ctx, "committing nil record")
		return ErrEmptyRecord
	}
	select {
	case r.commitChannel <- r:
	case <-r.context.Done():
		//listening to the context in case the consumer is shutting down
		tele.Warn(ctx, "record context done")
		return ErrContextExpired
	}

	return nil
}
