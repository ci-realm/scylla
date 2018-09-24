{ pkgs ? import ./nix/nixpkgs.nix }: with pkgs;
let goDeps = callPackage ./. {};
in mkShell {
  buildInputs = [
    nix
    dep2nix
    gotools
    goDeps.go
    nix-prefetch-git
    protobuf3_4
  ];
}
