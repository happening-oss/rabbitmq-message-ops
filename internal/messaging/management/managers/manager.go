package managers

import (
	"context"
)

type Manager interface {
	Manage(ctx context.Context, srcQueue string) (err error)
}
