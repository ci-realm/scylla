{ pkgs ? import ./nix/nixpkgs.nix }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    cacert
    cachix
    dbmate
    gcc
    git
    go
    gocode
    goimports
    golangci-lint
    go-langserver
    (lowPrio gotools)
    nix-prefetch-git
    nodejs
    pgcli
    vgo2nix
    yarn
    yarn2nix
    elmPackages.elm
    elm2nix
  ];
  PERL5LIB = "${pkgs.git.outPath}/lib/perl5/site_perl/5.28.0";

  CGO_ENABLED = "1";

  shellHook = ''
    unset preHook
    unset GOPATH
  '';
}
