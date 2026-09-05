#!/usr/bin/env bash
. "$(dirname "$0")/../../doctor_lib.sh"

# doctor examines an adopted repository; outside one it aborts naming
# the cause and changes nothing (product/rules.md).
mkdir -p "$WORK/bare"
cd "$WORK/bare" || exit 1
git_q init -q

check "a repository with no kit is refused" 1 "not an adopted repository" -- "$WRITRUN" doctor

finish
