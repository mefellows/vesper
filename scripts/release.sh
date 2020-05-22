#!/bin/bash

set -e
trap cleanup TERM

date=$(date "+%d %B %Y")

function step {
  echo " -----> $1"
}

function log {
  echo "        $1"
}

function get_version() {
  tag=` git tag -n1 | grep "chore(release)" | tail -n 1 | cut -d ' ' -f1`
  echo "${tag}"
}

# usage: increment_version current_version major|minor|patch
function increment_version() {
  current_version="${1}"
  increment="${2}"

  old_ifs=${IFS}
  IFS='.' read -a vers <<< "$current_version"
  IFS=${old_ifs}

  major=${vers[0]}
  minor=${vers[1]}
  patch=${vers[2]}

  case $increment in
    "major")
      ((major+=1))
      minor=0
      patch=0
      ;;
    "minor")
      ((minor+=1))
      patch=0
      ;;
    "patch")
      ((patch+=1))
      ;;
  esac

  echo "v$major.$minor.$patch"
}

function determine_increment() {
  changelog=${1}
  step="patch"
  [[ "${1}" =~ (feat:|feat\() ]] && step="minor"
  [[ "${1}" =~ (BREAKING|breaking) ]] && step="major"

  echo $step
}

function generate_changelog() {
  release_version=$1
  log=$(git log --pretty=format:'  * [%h](https://github.com/pact-foundation/pact-go/commit/%h) - %s (%an, %ad)' ${release_version}..HEAD | egrep -v "wip(:|\()" | grep -v "docs(" | grep -v "chore(" | grep -v Merge | grep -v "test(")

  log "Updating CHANGELOG.md"
  ed CHANGELOG.md << END
7i

### $release_version ($date)
$log
.
w
q
END
}

function cleanup() {
  log "ERROR generating release, please check your git logs and working tree to ensure things are in order."
}

step "Releasing Vesper 🚀 "
log "Finding current version"
current_version=$(get_version)

full_log=$(git log ${current_version}..HEAD)
inc=$(determine_increment "${full_log}")
version=$(increment_version "${current_version/v/}" "${inc}")
log "Increment '${inc}' version from ${current_version} to ${version}"

step "Generating changelog"
generate_changelog "${current_version}"
log "Changelog updated"

step "Committing changes"
log "unstaging files"
git reset HEAD
log "adding CHANGELOG.md"
git add CHANGELOG.md
log "commiting"
git commit -m "chore(release): release ${version}"

step "Creating tag ${version}"
git tag ${version} -m "chore(release): release ${version}"

log "Done - check your git logs, CHANGELOG, and then run 'git push --follow-tags'."