package main

import (
	"context"
	"flag"

	"github.com/spf13/cobra"
	"github.com/starine/aim/logger"
	"github.com/starine/aim/services/gateway"
	"github.com/starine/aim/services/router"
	"github.com/starine/aim/services/server"
	"github.com/starine/aim/services/service"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "aim",
		Version: version,
		Short:   "King IM Cloud",
	}
	ctx := context.Background()

	root.AddCommand(gateway.NewServerStartCmd(ctx, version))
	root.AddCommand(server.NewServerStartCmd(ctx, version))
	root.AddCommand(service.NewServerStartCmd(ctx, version))
	root.AddCommand(router.NewServerStartCmd(ctx, version))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
