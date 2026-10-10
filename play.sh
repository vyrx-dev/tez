#!/usr/bin/env bash
# Runs the prebuilt ./tez binary.
#
# On NixOS the libraries raylib needs (libpulse, libX11, libGL, libwayland)
# only exist inside /nix/store, with hash-based paths that a binary cannot
# find on its own. Without them raylib silently falls back to a "Null" audio
# backend and the game runs with no sound.
#
# The nix shell knows the right paths, so we borrow its environment rather than
# hardcoding store hashes that will change.
set -euo pipefail

cd "$(dirname "$0")"

if [ ! -x ./tez ]; then
  echo "tez: no binary yet, building one first"
  nix-shell --run "go build -o tez ."
fi

exec nix-shell --run "./tez"
