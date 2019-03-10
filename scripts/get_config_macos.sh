#!/bin/bash

# Generates config.json from files on macOS

token="$(strings /System/Library/PrivateFrameworks/GeoServices.framework/GeoServices | grep 'xyzABC' -A1 | tail -n1)"
rmurl="$(plutil -p ~/Library/Preferences/com.apple.GEO.plist | grep GEOLastResourceManifestURL | awk '{print $3}' | cut -d '"' -f 2)"

echo "{
  \"resourceManifestURL\": \"$rmurl\",
  \"tokenP1\": \"$token\"
}"
