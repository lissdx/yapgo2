package pipeline

import "context"

type Runnable[T, RS any] interface {
	Run(context.Context, T) RS
}
