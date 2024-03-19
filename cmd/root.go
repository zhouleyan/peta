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
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// NewPetaCommand creates a new peta root command.
func NewPetaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peta",
		Short: "Run and manage PETA",
		Long:  "Run and manage PETA...",
	}
	RegisterCommandRecursive(cmd)

	return cmd
}

func RegisterCommandRecursive(parent *cobra.Command) {
	initCmd := NewInitCmd()
	parent.AddCommand(initCmd)
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := NewPetaCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
