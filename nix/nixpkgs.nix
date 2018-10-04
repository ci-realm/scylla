import (
  fetchTarball {
    url = https://github.com/manveru/nixpkgs/archive/92d46857d5b8de67cf8ba62ba2e1d13c92db9ca7.tar.gz;
    sha256 = "1anmif94caphi2539n57fs0dyp46q3xqx8i8wrc7cg70729jvhra";
  }
) {
  config = {};
  overlays = [
    (import ./overlay.nix)
  ];
}
