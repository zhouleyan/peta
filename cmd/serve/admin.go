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
	"net/http"
	"peta.io/peta/pkg/config"
	"peta.io/peta/pkg/server"
	"peta.io/peta/pkg/signals"
)

func NewServerAdminCommand() *cobra.Command {
	c := config.NewServerAdminConfig()

	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Start the peta admin server.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(signals.SetupSignalHandler(), c)
		},
	}

	return cmd
}

func Run(ctx context.Context, c *config.Config) error {

	apiServer, err := server.NewAPIServer(ctx)
	if err != nil {
		return err
	}

	if err = apiServer.PreRun(); err != nil {
		return err
	}

	if errors.Is(apiServer.Run(ctx), http.ErrServerClosed) {
		return nil
	}
	// inner ctx is to control the life cycle of the peta admin
	//ctx, cancel := context.WithCancel(ctx)
	//defer cancel()
	//
	//go func() {
	//	<-ctx.Done()
	//}()
	//iCtx, cancelFunc := context.WithCancel(context.TODO())

	// The ctx(signals.SetupSignalHandler()) is to control the entire program life cycle,
	// The iCtx(internal context) is created here to control the life cycle of the peta admin serve
	//for {
	//	select {
	//	case <-ctx.Done():
	//		cancelFunc()
	//		return nil
	//	}
	//}
	return err
}
