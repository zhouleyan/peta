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
	"flag"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"peta.io/peta/cmd/initialize"
	"peta.io/peta/cmd/pg"
	"peta.io/peta/cmd/serve"
	"peta.io/peta/cmd/version"
	"peta.io/peta/pkg/log"
	"peta.io/peta/pkg/server/options"
)

// NewPetaCommand creates a new peta root command.
func NewPetaCommand() *cobra.Command {
	nfs := new(options.NamedFlagSets)
	fs := nfs.FlagSet("log")
	local := flag.NewFlagSet("local", flag.ExitOnError)
	log.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})
	cmd := &cobra.Command{
		Use:   "peta",
		Short: "Run and manage PETA",
		Long:  "Run and manage PETA...",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Setup()
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			log.Flush()
		},
	}

	pfs := cmd.PersistentFlags()
	for _, f := range nfs.FlagSets {
		pfs.AddFlagSet(f)
	}
	options.SetUsageAndHelpFunc(cmd, nfs)
	RegisterCommandRecursive(cmd)

	return cmd
}

func RegisterCommandRecursive(cmd *cobra.Command) {

	initialize.RegisterCommands(cmd)
	serve.RegisterCommands(cmd)
	version.RegisterCommands(cmd)
	pg.RegisterCommands(cmd)
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := NewPetaCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
