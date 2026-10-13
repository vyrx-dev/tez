{
  description = "A typing game with a stretchy blob";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "tez";
          version = "1.0.0";
          src = ./.;

          vendorHash = "sha256-6SU4S4l+qWCDhCC7lezL+D2Tz6/ALBbgccsVfA10WyA=";

          # raylib's C bindings, so the game links the real thing instead of
          # the purego path that needs libffi and friends at runtime.
          env.CGO_ENABLED = 1;

          nativeBuildInputs = [ pkgs.pkg-config pkgs.makeWrapper ];
          buildInputs = with pkgs; [
            raylib
            libGL
            libx11
            libxrandr
            libxi
            libxcursor
            libxinerama
            libxkbcommon
            wayland
            alsa-lib
            pulseaudio
          ];

          postInstall = ''
            wrapProgram $out/bin/tez \
              --prefix LD_LIBRARY_PATH : ${pkgs.lib.makeLibraryPath (with pkgs; [ alsa-lib pulseaudio libGL libx11 wayland libxkbcommon ])}
          '';

          ldflags = [ "-s" "-w" ];

          meta = {
            description = "A typing game with a stretchy blob";
            homepage = "https://github.com/vyrx-dev/tez";
            license = pkgs.lib.licenses.mit;
            mainProgram = "tez";
            platforms = pkgs.lib.platforms.unix;
          };
        };

        # Same shell as shell.nix, so `nix develop` works for people who
        # prefer flakes to plain nix-shell.
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            pkg-config
            goreleaser
          ];

          buildInputs = with pkgs; [
            raylib
            libGL
            libx11
            libxrandr
            libxi
            libxcursor
            libxinerama
            libxkbcommon
            wayland
            alsa-lib
            pulseaudio
          ];
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/tez";
        };
      }
    );
}
