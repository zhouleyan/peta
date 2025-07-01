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
	"fmt"
	"github.com/spf13/cobra"
	"peta.io/peta/pkg/clients/ssh"
	"peta.io/peta/pkg/log"
	"peta.io/peta/pkg/server/options"
)

func NewInitOSCommand(o *options.APIServerOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "os",
		Short: "Start the peta admin server.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Setup(o.LogOptions)
			defer log.Flush()
			return Run()
		},
		SilenceUsage: true,
	}

	return cmd
}

func Run() error {

	// private key for root: "/Users/zhouleyan/.ssh/id_ed25519"
	// "zly2104718987" for "zhouleyan"
	client, err := ssh.New(
		"root",
		"10.1.1.31",
		22,
		"123456",
		"",
		"",
		"",
		0,
		true,
		true,
	)
	if err != nil {
		return err
	}

	output, err := client.Run("ip a")

	if err != nil {
		log.Errorf(string(output))
		return err
	}
	fmt.Println(string(output))

	return nil
}
