#!/usr/bin/env sh
set -eu

ENV_FILE="../.env"
TMP_FILE="$(mktemp)"

ACCESS_SECRET="$(openssl rand -hex 64)"
REFRESH_SECRET="$(openssl rand -hex 64)"

awk -v acc="$ACCESS_SECRET" -v ref="$REFRESH_SECRET" '
    /^# TOKEN MANAGER #$/ { in_block=1; print; next }
    {
        if (in_block) {
            if ($0 ~ /^ACCESS_SECRET=/) { print "ACCESS_SECRET=" acc; next }
            if ($0 ~ /^REFRESH_SECRET=/) { print "REFRESH_SECRET=" ref; next }
            if ($0 ~ /^$/ || $0 ~ /^#/) { in_block=0 }
        }
        print
    }
' "$ENV_FILE" > "$TMP_FILE"

mv "$TMP_FILE" "$ENV_FILE"
