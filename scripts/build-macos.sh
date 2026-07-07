#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_NAME="CatScope"
VERSION="${1:-v0.6.3-preview}"
PRODUCT_VERSION="${VERSION#v}"
DIST_DIR="$ROOT/dist"
STAGING_DIR="$DIST_DIR/macos-dmg-staging"
APP_PATH="$ROOT/build/bin/$APP_NAME.app"
DMG_BASENAME="$APP_NAME-$VERSION-macos-universal"
DMG_PATH="$DIST_DIR/$DMG_BASENAME.dmg"
SHA_PATH="$DMG_PATH.sha256"

run() {
  printf '\n==> %s\n' "$*"
  "$@"
}

if [[ "$VERSION" != v* ]]; then
  echo "Version must start with v, for example v0.6.3-preview." >&2
  exit 1
fi

cd "$ROOT"

run go test ./...
run npm run build --prefix frontend
run wails build -platform darwin/universal -clean -skipbindings

if [[ ! -d "$APP_PATH" ]]; then
  echo "Expected app bundle was not found: $APP_PATH" >&2
  exit 1
fi

ARCHS="$(lipo -archs "$APP_PATH/Contents/MacOS/$APP_NAME")"
if [[ "$ARCHS" != *"x86_64"* || "$ARCHS" != *"arm64"* ]]; then
  echo "Expected universal app with x86_64 and arm64, got: $ARCHS" >&2
  exit 1
fi

mkdir -p "$DIST_DIR"
rm -rf "$STAGING_DIR"
rm -f "$DMG_PATH" "$SHA_PATH"
mkdir -p "$STAGING_DIR"

cp -R "$APP_PATH" "$STAGING_DIR/"
ln -s /Applications "$STAGING_DIR/Applications"

run hdiutil create \
  -volname "$APP_NAME $PRODUCT_VERSION" \
  -srcfolder "$STAGING_DIR" \
  -format UDZO \
  -ov "$DMG_PATH"

run hdiutil verify "$DMG_PATH"

(
  cd "$DIST_DIR"
  shasum -a 256 "$DMG_BASENAME.dmg" > "$DMG_BASENAME.dmg.sha256"
)

rm -rf "$STAGING_DIR"

printf '\nmacOS universal DMG created:\n'
printf -- '- %s\n' "$DMG_PATH"
printf -- '- %s\n' "$SHA_PATH"
printf 'Architectures: %s\n' "$ARCHS"
