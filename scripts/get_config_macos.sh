#!/bin/bash

# Generates config.json from files on macOS Catalina or older

set -Eeu

token="$(strings /System/Library/PrivateFrameworks/GeoServices.framework/GeoServices | grep 'xyzABC' -A1 | tail -n1)"
rmurl="$(plutil -p ~/Library/Preferences/com.apple.GEO.plist | grep GEOLastResourceManifestURL | awk '{print $3}' | cut -d '"' -f 2)"

if [ -z "$rmurl" ]; then
  tmp="$(mktemp -d)"
  cp ~/Library/Containers/com.apple.geod/Data/Library/Caches/com.apple.geod/GEOConfigStore.db* "$tmp"
  rmurl="$(sqlite3 "$tmp/GEOConfigStore.db" "select value from defaults where key = 'GEOLastResourceManifestURL';")"
  rm -r "$tmp"
fi

echo "{
  \"resourceManifestURL\": \"$rmurl\",
  \"tokenP1\": \"$token\"
}"
