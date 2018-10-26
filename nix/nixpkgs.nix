import (
  fetchTarball {
    url = https://github.com/NixOS/nixpkgs/archive/5f33fbbc7beef255c411fcbc3c4dea30fb260d6d.tar.gz;
    sha256 = "04scz08pfwic1byg30zrdhw6wmqjklxcblkxh12gk6kgj6zbnqbw";
  }
) {
  config = {};
  overlays = [
    (import ./overlay.nix)
  ];
}
