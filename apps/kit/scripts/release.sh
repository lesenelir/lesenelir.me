#!/usr/bin/env bash
# Releases Kit to GitHub: syncs the version, tests, builds a universal disk
# image signed with the local identity, tags kit-v<version>, pushes, and
# creates the release with the image and the CHANGELOG.md section of the
# version as its notes. The installed apps update from the kit-updates
# release, a fixed URL (mygo.json updates.url) that the GitHub API's rate
# limit does not touch: it gets the signed update archive, its deltas from
# earlier versions, and last the manifest that points the apps at them.
#
#   pnpm release:kit 0.1.0             # from the repo root
#   pnpm release:kit 0.1.0 --dry-run   # all but the tag, push and release
set -euo pipefail

usage() {
  echo "usage: scripts/release.sh <version> [--dry-run]" >&2
  exit 2
}
die() {
  echo "release: $*" >&2
  exit 1
}
step() { printf '\n\033[1m▸ %s\033[0m\n' "$*"; }

version=""
dry_run=0
for arg in "$@"; do
  case $arg in
  --dry-run) dry_run=1 ;;
  -*) usage ;;
  *) [[ -z $version ]] || usage; version=$arg ;;
  esac
done
[[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || usage

cd "$(dirname "$0")/.."
tag="kit-v$version"
identity="${KIT_SIGNING_IDENTITY:-Kit Local Signing}"
update_key="${KIT_UPDATE_KEY:-$HOME/Library/Application Support/mygo/update-keys/kit/mygo-update.key}"
feed_tag="kit-updates"
dmg="build/darwin-universal/Kit $version.dmg"
# GitHub turns the spaces of an uploaded file's name into dots; the release
# gets a copy named without them.
asset="build/darwin-universal/Kit-$version.dmg"

step "Checking"
[[ $(git rev-parse --abbrev-ref HEAD) == master ]] || die "releases are made from master"
[[ -z $(git status --porcelain) ]] || die "the working tree has changes; commit or stash them first"
if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
  die "the tag $tag exists already"
fi
if ((!dry_run)); then
  command -v gh >/dev/null || die "the gh CLI is needed: brew install gh"
  gh auth status >/dev/null 2>&1 || die "gh is not logged in: gh auth login"
  git fetch -q origin master
  [[ -z $(git rev-list HEAD..origin/master) ]] || die "origin/master has commits this branch lacks; pull first"
fi
notes=$(awk -v v="$version" '
  $0 == "## " v { found = 1; next }
  found && /^## / { exit }
  found { print }
' CHANGELOG.md)
[[ -n ${notes//[[:space:]]/} ]] || die "CHANGELOG.md has no \"## $version\" section"
[[ -n ${MYGO_UPDATER_PRIVATE_KEY:-} || -f $update_key ]] ||
  die "no update signing key at $update_key: installed apps only take updates signed with it"

step "Setting the version to $version"
for file in mygo.json package.json; do
  perl -0pi -e 's/"version": "[^"]*"/"version": "'"$version"'"/' "$file"
done
bumped=0
if [[ -n $(git status --porcelain mygo.json package.json) ]]; then
  bumped=1
  if ((dry_run)); then
    trap 'git checkout -q -- mygo.json package.json' EXIT
  else
    git commit -q -m "chore(kit): release $version" mygo.json package.json
  fi
fi

step "Testing"
CGO_ENABLED=0 go test ./...

step "Building"
./scripts/signing-identity.sh
rm -rf build/darwin-universal
if [[ -z ${MYGO_UPDATER_PRIVATE_KEY:-} ]]; then
  MYGO_UPDATER_PRIVATE_KEY=$(<"$update_key")
  export MYGO_UPDATER_PRIVATE_KEY
fi
go tool mygo build -sign "$identity" -platform darwin/universal
[[ -f $dmg ]] || die "the build made no $dmg"
manifest="build/darwin-universal/update-darwin-universal.json"
[[ -f $manifest ]] || die "the build made no update manifest"
shopt -s nullglob
update_files=(build/darwin-universal/*.tar.gz build/darwin-universal/*.delta)
shopt -u nullglob
((${#update_files[@]})) || die "the build made no update archive"
codesign --verify --deep --strict build/darwin-universal/Kit.app
cp "$dmg" "$asset"

body="$notes

## Install

1. Open \`Kit-$version.dmg\` and drag Kit to Applications.
2. Kit is signed by its author but not notarized by Apple, so macOS blocks it the first time. Run once:
   \`\`\`sh
   xattr -dr com.apple.quarantine /Applications/Kit.app
   \`\`\`
   or open Kit, then System Settings → Privacy & Security → Open Anyway.
3. Allow Accessibility, Screen Recording and Microphone when Kit asks, and add your OpenRouter API key (and a Soniox one,
   for faster speech to text) in Settings → Models.

## 安装

1. 打开 \`Kit-$version.dmg\`，把 Kit 拖到「应用程序」。
2. Kit 由作者签名，但没有经过 Apple 公证，第一次打开会被 macOS 拦截。在终端执行一次：
   \`\`\`sh
   xattr -dr com.apple.quarantine /Applications/Kit.app
   \`\`\`
   或者先打开一次 Kit，再到 系统设置 → 隐私与安全性 → 仍要打开。
3. 按提示允许辅助功能、屏幕录制和麦克风，并在 设置 → 模型 中填入 OpenRouter API Key（填 Soniox 的 Key 转写更快）。"

if ((dry_run)); then
  step "Dry run: not tagged, pushed or released"
  echo "Would tag $tag, push master and it, release $asset, and give the $feed_tag release:"
  printf '  %s\n' "${update_files[@]}" "$manifest"
  echo
  echo "The notes of $tag:"
  echo
  echo "$body"
  ((bumped)) && echo && echo "(The version change of mygo.json and package.json is undone.)"
  exit 0
fi

step "Tagging $tag and pushing"
git tag -a "$tag" -m "Kit $version"
git push origin master "$tag"

step "Creating the release"
# --latest=false: the repository's latest release stays the website's or
# another app's, as Kit's tags have a prefix of their own.
gh release create "$tag" "$asset" --title "Kit $version" --notes "$body" --latest=false

step "Publishing the update"
if ! gh release view "$feed_tag" >/dev/null 2>&1; then
  gh release create "$feed_tag" --prerelease --latest=false --title "Kit updates" \
    --notes "The update feed of Kit, which installed apps check: signed update archives and the manifest of the newest version. To install Kit, download the disk image of the newest kit-v release instead."
fi
# The archives first, then the manifest that points at them.
gh release upload "$feed_tag" "${update_files[@]}" --clobber
gh release upload "$feed_tag" "$manifest" --clobber
echo
echo "Released Kit $version: $(gh release view "$tag" --json url -q .url)"
