#!/usr/bin/env bash
#
# Decides which upstream hashtree commit the update workflow builds, and whether
# that build is to be published as a release.
#
# Normally that is the tip of upstream main. But when upstream has published a
# version tag that this repository has not released yet, the tagged commit is
# built instead, so the bindings release for that version is built from exactly
# the upstream release. The next run then carries on with upstream main.
#
# Usage: pick-upstream-target.sh <hashtree checkout> <stored commit> <our tags>
#   <our tags> is a newline-separated list of the tags of this repository.
#
# Prints `commit=<sha>` and `release_tag=<tag or empty>` (GITHUB_OUTPUT format).

set -euo pipefail

upstream="$1"
stored="$2"
our_tags="$3"

version_tags() { { grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' || true; } | sort -V; }

latest_ours="$(printf '%s\n' "$our_tags" | version_tags | tail -1)"
commit="$(git -C "$upstream" rev-parse origin/main)"
release_tag=""

for tag in $(git -C "$upstream" tag --list 'v*' --merged origin/main | version_tags); do
	# Already released here.
	if printf '%s\n' "$our_tags" | grep -qxF "$tag"; then
		continue
	fi

	# Older than our latest release: a version we skipped, not a new one.
	if [ -n "$latest_ours" ] &&
		[ "$(printf '%s\n%s\n' "$tag" "$latest_ours" | sort -V | tail -1)" = "$latest_ours" ]; then
		continue
	fi

	tag_commit="$(git -C "$upstream" rev-list -n 1 "$tag")"

	# The libs we ship are already newer than this release, so it can no longer
	# be built from the tagged commit without going backwards.
	if [ -n "$stored" ] && [ "$tag_commit" != "$stored" ] &&
		git -C "$upstream" merge-base --is-ancestor "$tag_commit" "$stored" 2>/dev/null; then
		echo "::warning::upstream $tag (${tag_commit:0:7}) is older than the libs already" \
			"shipped (${stored:0:7}); not released automatically" >&2
		continue
	fi

	commit="$tag_commit"
	release_tag="$tag"
	break
done

echo "commit=$commit"
echo "release_tag=$release_tag"
