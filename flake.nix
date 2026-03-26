{
  description = "Sergio TD dev environment (Godot client + Go server)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
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
        pkgs = import nixpkgs {
          inherit system;
          config = {
            allowUnfree = true;
          };
        };

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
            exec godot4 --path ${self}/client -- "$@"
          '';
        };
        runDev = pkgs.writeShellApplication {
          name = "sergio-td-dev";
          runtimeInputs = [
            pkgs.go
            pkgs.godot_4
            pkgs.coreutils
            pkgs.stdenv.cc
          ];
          text = ''
              set -e

              echo "[sergio-td] Starting server..."
              (
               cd ${self}/server
               go run ./cmd/server
              ) &
              SERVER_PID=$!
              until nc -z localhost 8080; do
                sleep 0.2
              done

              echo "[sergio-td] Starting client..."
              godot4 --path ${self}/client &
              CLIENT_PID=$!

              cleanup() {
                echo "Stopping processes..."
                  kill $SERVER_PID $CLIENT_PID 2>/dev/null || true
              }

            trap cleanup EXIT

              wait
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
            claude-code
            opencode
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

          dev = {
            type = "app";
            program = "${runDev}/bin/sergio-td-dev";
          };
        };
      }
    );
}
