#!/usr/bin/env bash

#
#  This file is part of PETA.
#  Copyright (C) 2024 The PETA Authors.
#  PETA is free software: you can redistribute it and/or modify
#  it under the terms of the GNU Affero General Public License as published by
#  the Free Software Foundation, either version 3 of the License, or
#  (at your option) any later version.
#
#  PETA is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
#  GNU Affero General Public License for more details.
#
#  You should have received a copy of the GNU Affero General Public License
#  along with PETA. If not, see <https://www.gnu.org/licenses/>.
#

GO_PACKAGE=peta.io/peta

#version::get_version_vars() {
#
#}

# Prints the value that needs to be passed to the -ldflags parameter of go build
version::ldflags() {
  local -a ldflags
  function add_ldflag() {
    local key=${1}
    local val=${2}

    ldflags+=(
      "-X '${GO_PACKAGE}/pkg/version.${key}=${val}'"
    )
  }

  add_ldflag "buildDate" "$(date ${SOURCE_DATE_EPOCH:+"--date=@${SOURCE_DATE_EPOCH}"} -u +'%Y-%m-%dT%H:%M:%SZ')"

  # The -ldflags parameter takes a single string, so join the output.
  echo "${ldflags[*]-}"
}
