{ pkgs ? import ./nix/nixpkgs.nix }: with pkgs;
let
  gems = bundlerEnv {
    inherit ruby_2_5;
    name = "scylla-dev-gems";
    gemdir = ./.;
  };
  env = buildEnv {
    name = "scylla-env";
    paths = [
      yarn
      vgo2nix
      cachix
      yarn2nix
      nodejs
      dbmate
      (lowPrio gotools)
      gocode
      goimports
      golangci-lint
      go
      gcc
      nix-prefetch-git
      git
      protobuf3_4
      ejson
      gems.wrappedRuby
      (lowPrio gems)
    ];
  };
in mkShell {
  buildInputs = [ env ];
  PERL5LIB = "${git.outPath}/lib/perl5/site_perl/5.28.0";

  CGO_ENABLED = "1";
  GO111MODULE = "on";
}
