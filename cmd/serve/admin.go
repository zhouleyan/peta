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
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"net/http"
	"peta.io/peta/pkg/server"
	"peta.io/peta/pkg/server/options"
	"peta.io/peta/pkg/signals"
)

func NewServerAdminCommand(o *options.APIServerOptions, nfs *options.NamedFlagSets) *cobra.Command {
	AddFlags(o, nfs)

	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Start the peta admin server.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := MergeConfig(cmd.Flags(), o)
			if err != nil {
				return fmt.Errorf("misconfiguration \n%v", err)
			}
			return Run(signals.SetupSignalHandler(), o)
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

func MergeConfig(fs *pflag.FlagSet, o *options.APIServerOptions) (*options.APIServerOptions, error) {
	c, err := options.LoadConfig(o.ConfigFile)
	if err != nil {
		klog.Fatalf("failed to load config from disk: %v", err)
	}
	o.Merge(fs, c)
	return o, errors.Join(o.Validate()...)
}

func AddFlags(o *options.APIServerOptions, nfs *options.NamedFlagSets) {
	o.AuditingOptions.AddFlags(nfs.Insert("auditing", 1))
	o.MetricsOptions.AddFlags(nfs.Insert("metrics", 1))
	o.DatabaseOptions.AddFlags(nfs.Insert("database", 1))
}
