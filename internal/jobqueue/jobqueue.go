// Package jobqueue provides a generic interface to a simple job queue.
//
// Jobs are pushed to the queue with Enqueue(). Workers call Dequeue() to
// receive a job and FinishJob() to report one as finished.
//
// Each job has a type and arguments corresponding to this type. These are
// opaque to the job queue, but it mandates that the arguments must be
// serializable to JSON. Similarly, a job's result has opaque result arguments
// that are determined by its type.
//
// A job can have dependencies. It is not run until all its dependencies have
// finished.
package jobqueue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// JobQueue is an interface to a simple job queue. It is safe for concurrent use.
type JobQueue interface {
	// Enqueues a job.
	//
	// `args` must be JSON-serializable and fit the given `jobType`, i.e., a worker
	// that is running that job must know the format of `args`.
	//
	// All dependencies must already exist, but the job isn't run until all of them
	// have finished.
	//
	// Returns the id of the new job, or an error.
	Enqueue(jobType string, args interface{}, dependencies []uuid.UUID) (uuid.UUID, error)

	// Dequeues a job, blocking until one is available.
	//
	// Waits until a job with a type of any of `jobTypes` is available, or `ctx` is
	// canceled.
	//
	// Returns the job's id, dependencies, type, and arguments, or an error. Arguments
	// can be unmarshaled to the type given in Enqueue().
	Dequeue(ctx context.Context, jobTypes []string) (uuid.UUID, []uuid.UUID, string, json.RawMessage, error)

	// Mark the job with `id` as finished. `result` must fit the associated
	// job type and must be serializable to JSON.
	FinishJob(id uuid.UUID, result interface{}) error

	// Cancel a job. Does nothing if the job has already finished.
	CancelJob(id uuid.UUID) error

	// If the job has finished, returns the result as raw JSON.
	//
	// Returns the current status of the job, in the form of three times:
	// queued, started, and finished. `started` and `finished` might be the
	// zero time (check with t.IsZero()), when the job is not running or
	// finished, respectively.
	//
	// Lastly, the IDs of the jobs dependencies are returned.
	JobStatus(id uuid.UUID) (result json.RawMessage, queued, started, finished time.Time, canceled bool, deps []uuid.UUID, err error)

	// Returns the job's arguments in Raw form.
	JobArgs(id uuid.UUID) (args json.RawMessage, err error)
}

var (
	ErrNotExist   = errors.New("job does not exist")
	ErrNotRunning = errors.New("job is not running")
	ErrCanceled   = errors.New("job ws canceled")
)
