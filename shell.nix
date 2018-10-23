{ pkgs ? import ./nix/nixpkgs.nix }: with pkgs;
let
  default = callPackage ./. {};
  gems = bundlerEnv {
    inherit ruby_2_5;
    name = "scylla-dev-gems";
    gemdir = ./.;
  };
  env = buildEnv {
    name = "scylla-env";
    paths = [
      yarn
      dbmate
      dep2nix
      (lowPrio gotools)
      gocode
      goimports
      golangci-lint
      go
      nix-prefetch-git
      git
      protobuf3_4
      remarshal
      ejson
      gems.wrappedRuby
      (lowPrio gems)
    ];
  };
in mkShell {
  buildInputs = [ env ];
  PERL5LIB = "${git.outPath}/lib/perl5/site_perl/5.28.0";

  CGO_ENABLED = "1";
  GOPATH="/home/manveru/go";
  GOROOT="${go}/share/go";

  shellHook = ''
    export GOPATH="$HOME/go";
    export GOROOT="${go}/share/go"
    if [[ -e shell.nix ]]; then
      set -x
      rm -rf vendor
      ln -s ${default.depTree}/vendor $PWD/vendor
      set +x
    fi
  '';
}
