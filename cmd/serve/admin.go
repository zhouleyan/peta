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
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"net/http"
	"peta.io/peta/pkg/server"
	"peta.io/peta/pkg/server/options"
	"peta.io/peta/pkg/signals"
)

func NewServerAdminCommand(o *options.APIServerOptions, nfs options.NamedFlagSets) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Start the peta admin server.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := options.LoadConfig(o.ConfigFile)
			if err != nil {
				klog.Fatal("Failed to load config from disk: %v", err)
			}
			o.Merge(cmd.Flags(), c)
			return Run(signals.SetupSignalHandler(), o)
		},
		SilenceUsage: true,
	}

	o.AuditingOptions.AddFlags(nfs.Insert("auditing", 1))

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
