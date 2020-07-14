#!/bin/sh

cd ..
version_string=$(< VERSION)
IFS='=' read -r -a latest_version <<< "$version_string"
if grep -q "LANTERN_USER_AGENT" ".env"; then
    lineNum="$(grep -n "LANTERN_USER_AGENT" .env | head -n 1 | cut -d: -f1)"
    current_version=`sed "${lineNum}q;d" .env`
    version_number="$(grep -o "[0-9]\+.[0-9]\+.[0-9]\+" <<< ${current_version})"
    if [ "${latest_version[1]}" != "$version_number" ]; then
        sed -i -e "${lineNum}d" .env
    else
        exit
    fi
fi

echo "\nLANTERN_USER_AGENT="LANTERN/${latest_version[1]}"" >> .env


