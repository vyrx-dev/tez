{pkgs ? import <nixpkgs> {}}: let
  libraries = with pkgs; [
    alsa-lib
    pulseaudio
    libGL
    wayland
    libxkbcommon
  ];
in
  pkgs.mkShell {
    packages = with pkgs; [
      go
      pkg-config
      goreleaser
    ];

    buildInputs =
      libraries
      ++ [
        pkgs.raylib
        pkgs.libx11
        pkgs.libxrandr
        pkgs.libxi
        pkgs.libxcursor
        pkgs.libxinerama
        pkgs.wayland
        pkgs.wayland-protocols
        pkgs.libxkbcommon
      ];

    shellHook = ''
      export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath libraries}:$LD_LIBRARY_PATH
      echo "tez dev shell ready."
      echo ""
      echo "  go run .                       run the game"
      echo "  go build -o tez . && ./tez     build a binary, then run it"
      echo "  CGO_ENABLED=0 go run .         run without raylib installed, bundles its own libraylib"
      echo "  go test -tags test ./...       run the maths tests (no raylib, no display)"
      echo "  goreleaser release --snapshot --clean   build a full release into dist/"
      echo ""
      echo "Note: run ./tez from inside this shell. Outside it, raylib cannot find"
      echo "libpulse/libasound and falls back to a silent 'Null' audio backend."
    '';
  }
