{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
        {
          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              postgresql
            ];

            shellHook = ''
              export PG=$PWD/.dev_postgres/
              export PGDATA=$PG/data
              export PGPORT=5432
              export PGHOST=localhost
              export PGUSER=$USER
              export PGPASSWORD=postgres
              export PGDATABASE=pinguino
              export DB_URL=postgres://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE?sslmode=disable
              echo "Run . hook.sh to start the database"
            '';
          };
        }
    );
}
