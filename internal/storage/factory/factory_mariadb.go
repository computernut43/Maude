package factory

import (
	"context"

	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/mariadb"
)

func init() {
	RegisterBackend(configfile.BackendMariaDB, func(ctx context.Context, path string, opts Options) (storage.Storage, error) {
		store, err := mariadb.New(ctx, &mariadb.Config{
			Host:     opts.ServerHost,
			Port:     opts.ServerPort,
			User:     opts.ServerUser,
			Database: opts.Database,
			ReadOnly: opts.ReadOnly,
		})
		if err != nil {
			return nil, err
		}
		return store, nil
	})
}
