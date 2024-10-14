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

package cmd

import (
	"github.com/spf13/cobra"
	"peta.io/peta/pkg/server"
)

func NewServeAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Serves PETA Admin APIs",
		Long: `This command opens one port and listens to API requests. The exposed API handles administrative
requests like managing users, roles, groups and others.`,
		RunE:         server.RunServeAdmin(),
		SilenceUsage: true,
	}
	return cmd
}
