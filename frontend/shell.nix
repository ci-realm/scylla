{ pkgs ? import ./nix/nixpkgs.nix }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    cacert
    yarn
    yarn2nix
    elmPackages.elm
    elm2nix
  ];

  shellHook = ''
    unset preHook
  '';
}
