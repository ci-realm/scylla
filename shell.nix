{ pkgs ? import ./nix/nixpkgs.nix }: with pkgs;
let
  goDeps = callPackage ./. {};
  gems = bundlerEnv {
    inherit ruby_2_5;
    name = "scylla-dev-gems";
    gemdir = ./.;
  };
in mkShell {
  buildInputs = [
    nix
    dbmate
    dep2nix
    gotools
    gocode
    goimports
    golangci-lint
    goDeps.go
    nix-prefetch-git
    protobuf3_4
    remarshal
    ejson
    gems.wrappedRuby
    (lowPrio gems)
  ];
}
