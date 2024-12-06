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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log"
	"peta.io/peta/pkg/server/options"
	"peta.io/peta/pkg/signals"
	"runtime"
)

func NewServePublicCommand(o *options.APIServerOptions) *cobra.Command {
	nfs := o.Flags()

	cmd := &cobra.Command{
		Use:   "public",
		Short: "Start the peta public server.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := signals.SetupSignalHandler()

			// `ctx` controls the entire program, cancel when get os termination signal
			// `ctx => serverCtx` controls the server life-cycle, cancel when receive reload signal(config changed)
			return RunPublic(ctx, cmd.Flags(), o)
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

func RunPublic(ctx context.Context, fs *pflag.FlagSet, o *options.APIServerOptions) error {
	log.Println(runtime.NumGoroutine())
	return nil
}
