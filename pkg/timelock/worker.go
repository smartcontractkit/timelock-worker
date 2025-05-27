package timelock

import "context"

type Worker interface {
	Listen(ctx context.Context) error
}
