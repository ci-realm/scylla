import (
  fetchTarball {
    url = https://github.com/nixos/nixpkgs-channels/archive/f753852e11d72c05cb74d1058ea8b7f6d5dd4748.tar.gz;
    sha256 = "0xvjrsi3j4hzq9cdzqpccxnl9gqc8f5y59lkgqs2s2dkng35zv74";
  }
) {
  config = {};
  overlays = [
    (import ./overlay.nix)
  ];
}
