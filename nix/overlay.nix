self: super:
let
  manveru-nur-packages = fetchTarball {
    url = "https://github.com/manveru/nur-packages/archive/1f492c51eedd1f6abac7b753a7e56dfe46be1860.tar.gz";
    sha256 = "0m5aywc2lnmycj6vvaa7d9p3wj763crpjrwqqylpq6j7bviniq4n";
  };
  manveru-nixpkgs = fetchTarball {
    url = https://github.com/manveru/nixpkgs/archive/07eb9736f0958582411d1c866c34f24f55827801.tar.gz;
    sha256 = "00gjv74ca7hmllq2mf2y3vvfdkdn22gdb1vpxm5ypi4ysg8dvpr7";
  };
in {
  git-info = (self.callPackage "${manveru-nur-packages}/default.nix" {}).lib.git-info;
  go = super.go_1_11;
  gotools = super.gotools.override { go = super.go_1_11; };
}
