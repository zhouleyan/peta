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

version::get_version_vars() {
	local git=(git --work-tree "${ROOT}")

	if [[ -n ${GIT_COMMIT-} ]] || GIT_COMMIT=$("${git[@]}" rev-parse "HEAD^{commit}" 2>/dev/null); then
		if [[ -z ${GIT_TREE_STATE-} ]]; then
			# Check if the tree is dirty.  default to dirty
			if git_status=$("${git[@]}" status --porcelain 2>/dev/null) && [[ -z ${git_status} ]]; then
				GIT_TREE_STATE="clean"
			else
				GIT_TREE_STATE="dirty"
			fi
		fi

		# Use git describe to find the version based on tags.
		if [[ -n ${GIT_VERSION-} ]] || GIT_VERSION=$("${git[@]}" describe --tags --match='v*' --abbrev=14 "${GIT_COMMIT}^{commit}" 2>/dev/null); then
			# This translates the "git describe" to an actual semver.org
			# compatible semantic version that looks something like this:
			#   v1.1.0-alpha.0.6+84c76d1142ea4d
			#
			# TODO: We continue calling this "git version" because so many
			# downstream consumers are expecting it there.
			#
			# These regexes are painful enough in sed...
			# We don't want to do them in pure shell, so disable SC2001
			# shellcheck disable=SC2001
			DASHES_IN_VERSION=$(echo "${GIT_VERSION}" | sed "s/[^-]//g")
			if [[ "${DASHES_IN_VERSION}" == "---" ]]; then
				# shellcheck disable=SC2001
				# We have distance to subversion (v1.1.0-subversion-1-gCommitHash)
				GIT_VERSION=$(echo "${GIT_VERSION}" | sed "s/-\([0-9]\{1,\}\)-g\([0-9a-f]\{14\}\)$/.\1\+\2/")
			elif [[ "${DASHES_IN_VERSION}" == "--" ]]; then
				# shellcheck disable=SC2001
				# We have distance to base tag (v1.1.0-1-gCommitHash)
				GIT_VERSION=$(echo "${GIT_VERSION}" | sed "s/-g\([0-9a-f]\{14\}\)$/+\1/")
			fi
			if [[ "${GIT_TREE_STATE}" == "dirty" ]]; then
				# git describe --dirty only considers changes to existing files, but
				# that is problematic since new untracked .go files affect the build,
				# so use our idea of "dirty" from git status instead.
				GIT_VERSION+="-dirty"
			fi

			# Try to match the "git describe" output to a regex to try to extract
			# the "major" and "minor" versions and whether this is the exact tagged
			# version or whether the tree is between two tagged versions.
			if [[ "${GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)?([-].*)?([+].*)?$ ]]; then
				# shellcheck disable=SC2034
				GIT_MAJOR=${BASH_REMATCH[1]}
				GIT_MINOR=${BASH_REMATCH[2]}
				if [[ -n "${BASH_REMATCH[4]}" ]]; then
					GIT_MINOR+="+"
				fi
			fi

			# If GIT_VERSION is not a valid Semantic Version, then refuse to build.
			if ! [[ "${GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)?(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
				echo "GIT_VERSION should be a valid Semantic Version. Current value: ${GIT_VERSION}"
				echo "Please see more details here: https://semver.org"
				exit 1
			fi
		fi
	fi
}

# Prints the value that needs to be passed to the -ldflags parameter of go build
version::ldflags() {
	version::get_version_vars

	local -a ldflags
	function add_ldflag() {
		local key=${1}
		local val=${2}

		ldflags+=(
			"-X '${GO_PACKAGE}/pkg/version.${key}=${val}'"
		)
	}

	add_ldflag "buildDate" "$(date ${SOURCE_DATE_EPOCH:+"--date=@${SOURCE_DATE_EPOCH}"} -u +'%Y-%m-%dT%H:%M:%SZ')"
	if [[ -n ${GIT_COMMIT-} ]]; then
		add_ldflag "gitCommit" "${GIT_COMMIT}"
		add_ldflag "gitTreeState" "${GIT_TREE_STATE}"
	fi

	if [[ -n ${GIT_VERSION-} ]]; then
		add_ldflag "gitVersion" "${GIT_VERSION}"
	fi

	if [[ -n ${GIT_MAJOR-} && -n ${GIT_MINOR-} ]]; then
		add_ldflag "gitMajor" "${GIT_MAJOR}"
		add_ldflag "gitMinor" "${GIT_MINOR}"
	fi
	# The -ldflags parameter takes a single string, so join the output.
	echo "${ldflags[*]-}"
}
