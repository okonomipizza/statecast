{
  description = "statecast - AI agent state sharing daemon";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    go-overlay = {
      url = "github:purpleclay/go-overlay";
      inputs = {
        nixpkgs.follows = "nixpkgs";
        flake-utils.follows = "flake-utils";
      };
    };
  };

  outputs = { nixpkgs, flake-utils, go-overlay, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ go-overlay.overlays.default ];
        };
        go = pkgs.go-bin.fromGoMod ./go.mod;
        statecast = pkgs.callPackage ./nix/package.nix {
          inherit go;
          src = ./.;
        };
      in
      {
        packages.default = statecast;
        packages.statecast = statecast;

        apps.default = flake-utils.lib.mkApp {
          drv = statecast;
        };

        devShells.default = import ./nix/devShell.nix {
          inherit pkgs go;
          govendor = go-overlay.packages.${system}.govendor;
        };
      }
    );
}
