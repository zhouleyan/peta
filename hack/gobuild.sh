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

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT}/hack/version.sh"

VERBOSE=${VERBOSE:-"0"}
# V=""
if [[ "${VERBOSE}" == "1" ]];then
    # V="-x"
    set -x
fi

OUTPUT_DIR=bin
BUILDPATH=./${1:?"path to build"}
OUT=${OUTPUT_DIR}/${1:?"output path"}

BUILD_GOOS=${GOOS:-$(go env GOOS)}
BUILD_GOARCH=${GOARCH:-$(go env GOARCH)}
GO_BINARY=${GOBINARY:-go}
LDFLAGS=$(version::ldflags)

GOOS=${BUILD_GOOS} CGO_ENABLED=0 GOARCH=${BUILD_GOARCH} ${GO_BINARY} build \
        -ldflags="${LDFLAGS}" \
        -o "${OUT}" \
        "${BUILDPATH}"
