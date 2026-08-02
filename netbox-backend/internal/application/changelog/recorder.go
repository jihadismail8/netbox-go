package changelog

import "context"

type Recorder interface {
	Record(context.Context, Change) error
}
