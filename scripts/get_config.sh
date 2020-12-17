#!/bin/bash

set -Eeuo pipefail
e() { printf "\033[0;31m$1\033[0m\n" >&2; }

# check if tools are available
which curl > /dev/null || which wget > /dev/null || (e 'neither curl nor wget found. please install' && exit 1)
which 7z > /dev/null || (e "7z not found. try installing 'p7zip' or 'p7zip-full'" && exit 1)
echo _ | strings > /dev/null || (e "strings not found. try installing 'binutils' or xcode command line tools" && exit 1)

# print warning
printf '\e[33mThis downloads ~2 GB from Apple and uses ~6 GB temporary disk space.\033[0m\n' >&2
read -p $'\e[33mContinue? (y/n): \033[0m' x && [ "$x" == "y" ] || exit 1

dmg="com.apple.pkg.iPhoneSimulatorSDK10_3-10.3.1.1495751597.dmg"
gs='./Contents/Resources/RuntimeRoot/System/Library/PrivateFrameworks/GeoServices.framework/GeoServices'
dq="?application=geod&application_version=1&country_code=US&hardware=MacBookPro11,2&os=osx&os_build=20B29&os_version=11.0.1"

# create temp directory
tmp="$(mktemp -d)"
echo "Temp dir: $tmp" >&2
cd "$tmp"

# download using either wget or curl
poly(){
    which wget > /dev/null && wget -q --show-progress "$1" -O "$2" && return
    which curl > /dev/null && curl -# --fail "$1" > "$2" && return    
    e "Could not download $1"
    exit 1
}
echo "Downloading SDK" >&2
poly "https://devimages-cdn.apple.com/downloads/xcode/simulators/$dmg" "$dmg"

# pretty print                                      # extract geoservices                       # remove intermediates
echo -ne "\033[2K\r$dmg\n" >&2
echo -ne "└── Extracting" >&2;                      7z x "$dmg" '*/*.pkg' -bsp2 > /dev/null;    rm "$dmg"
                                                    pkg="$(find . -name "*.pkg" -type f)"
echo -ne "\033[2K\r└── $pkg\n" >&2
echo -ne "    └── Extracting" >&2;                  7z x "$pkg" Payload~ -bsp2 > /dev/null;     rm "$pkg"
echo -ne "\033[2K\r    └── Payload\n" >&2
echo -ne "        └── Extracting" >&2;              7z x Payload~ "$gs" -bsp2 > /dev/null;      rm Payload~
echo -ne "\033[2K\r        └── GeoServices\n" >&2

# parse base url and token
base_url="$(strings "$gs" | grep config%{DEVICE_QUERY} | tr '%' '\n' | head -n1)"
token="$(strings "$gs" | grep xyzABC -A1 | tail -n1)"
printf "            ├── Base URL\n" >&2
printf "            └── Token\n" >&2
rmurl="$base_url$dq"

# output
printf "\033[0;32mWriting JSON\033[0m\n" >&2
echo "{
  \"resourceManifestURL\": \"$rmurl\",
  \"tokenP1\": \"$token\"
}"

# remove temp directory
cd ..
rm -rf "$tmp"
