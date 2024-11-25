/*
 *  This file is part of PETA.
 *  Copyright (C) 2024 The PETA Authors.
 *  PETA is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  PETA is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with PETA. If not, see <https://www.gnu.org/licenses/>.
 */

package serve

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"peta.io/peta/pkg/server"
	"peta.io/peta/pkg/server/options"
	"peta.io/peta/pkg/signals"
)

func NewServeAdminCommand(o *options.APIServerOptions) *cobra.Command {
	nfs := o.Flags()
	o.AddFlags(nfs)

	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Start the peta admin server.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := options.MergeConfig(cmd.Flags(), o)
			ctx := signals.SetupSignalHandler()
			options.WatchConfig(ctx, o)
			if err != nil {
				return fmt.Errorf("misconfiguration \n%v", err)
			}
			return Run(ctx, o)
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	for _, f := range nfs.FlagSets {
		fs.AddFlagSet(f)
	}

	options.SetUsageAndHelpFunc(cmd, nfs)
	return cmd
}

func Run(ctx context.Context, o *options.APIServerOptions) error {

	apiServer, err := server.NewAPIServer(ctx, o)
	if err != nil {
		return err
	}

	if err = apiServer.PreRun(); err != nil {
		return err
	}

	if errors.Is(apiServer.Run(ctx), http.ErrServerClosed) {
		return nil
	}

	return err
}
