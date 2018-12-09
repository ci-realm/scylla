import (
  fetchTarball {
    url = https://github.com/nixos/nixpkgs-channels/archive/a06177e65ac17ca0f043f81f6eeb5223290cbaca.tar.gz;
    sha256 = "1fnr0sm205lj2s1pkzskva5riq8p5fiv58995p1a5gmjrfq0kw2s";
  }
) {
  config = {};
  overlays = [
    (import ./overlay.nix)
  ];
}
