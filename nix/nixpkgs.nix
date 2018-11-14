import (
  # fetchTarball {
  #   url = https://github.com/NixOS/nixpkgs/archive/e8df5045cac40f9919e7b14d356308ace53a947e.tar.gz;
  #   sha256 = "09k3azns0cwhndgklz5yhkka7a509s5nyvnfz30pvrl1gj4qqg2w";
  # }
  fetchTarball {
    url = https://github.com/nixos/nixpkgs-channels/archive/c9e13806267f7fd3351d52a19cc6db5fa2985ca9.tar.gz;
    sha256 = "0qsa3j4i2ndiw4yxla3y4i5f8r12waj34h2z84xjig4l54cx184q";
  }
) {
  config = {};
  overlays = [
    (import ./overlay.nix)
  ];
}
