package transaction

import "context"

type Manager interface {
	Within(ctx context.Context, fn func(context.Context) error) error
}
