/*
 *  This file is part of PETA.
 *  Copyright (C) 2025 The PETA Authors.
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

package initialize

import (
	"github.com/spf13/cobra"
	"peta.io/peta/pkg/server/options"
)

func NewServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the peta environment.",
	}
}

func RegisterCommands(parent *cobra.Command) {
	o := options.NewAPIServerOptions()
	nfs := o.Flags()
	cmd := NewServeCommand()
	fs := cmd.PersistentFlags()
	for _, f := range nfs.FlagSets {
		fs.AddFlagSet(f)
	}

	options.SetUsageAndHelpFunc(cmd, nfs)

	parent.AddCommand(cmd)
	cmd.AddCommand(NewInitOSCommand(o))
}
