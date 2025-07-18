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

package pg

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"peta.io/peta/pkg/log"
	"peta.io/peta/pkg/types"
	"peta.io/peta/pkg/utils/errutils"
)

const (
	defaultBlueprintName = "blueprint"
)

func NewPGCreateCommand() *cobra.Command {
	blueprint := ""
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create Postgres instance.",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			errutils.CheckErr(Run(blueprint))
		},
	}

	cmd.Flags().StringVarP(&blueprint, "blueprint", "b", "blueprint.yml", "Specify a blueprint file")

	return cmd
}

func Run(blueprint string) error {
	b, err := loadFromBlueprint(blueprint)
	log.Infoln(b)
	return err
}

func loadFromBlueprint(blueprint string) (b *types.Blueprint, err error) {
	b = &types.Blueprint{}
	fp, err := filepath.Abs(blueprint)
	if err != nil {
		return nil, err
	}

	_, err = os.Open(fp)
	if err != nil {
		log.Errorln(errors.Wrap(err, "unable to open the given blueprint file"))
		return nil, errors.Wrap(err, "unable to open the given blueprint file")
	}

	return b, err
}
