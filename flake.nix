{
  description = "Sergio TD dev environment (Godot client + Go server)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        runServer = pkgs.writeShellApplication {
          name = "sergio-td-server";
          runtimeInputs = [ pkgs.go ];
          text = ''
            cd ${self}/server
            exec go run ./cmd/server
          '';
        };

        runClient = pkgs.writeShellApplication {
          name = "sergio-td-client";
          runtimeInputs = [ pkgs.godot_4 ];
          text = ''
            exec godot4 --path ${self}/client
          '';
        };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            golangci-lint
            gotools
            godot_4
            websocat
            jq
          ];

          shellHook = ''
            echo "[sergio-td] Ambiente carregado."
            echo "- Run server: nix run .#server"
            echo "- Run client: nix run .#client"
          '';
        };

        apps = {
          default = {
            type = "app";
            program = "${runServer}/bin/sergio-td-server";
          };

          server = {
            type = "app";
            program = "${runServer}/bin/sergio-td-server";
          };

          client = {
            type = "app";
            program = "${runClient}/bin/sergio-td-client";
          };
        };
      });
}
