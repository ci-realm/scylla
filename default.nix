{ stdenv, lib, buildGoPackage, fetchFromGitHub, makeWrapper, git, nixUnstable }:

let
  # Files that should not trigger a rebuild or be in the nix store.
  ignore = map (path: toString path) [
    ./.envrc
    ./Dockerfile
    ./Gopkg.lock
    ./Gopkg.toml
    ./ci
    ./default.nix
    ./deps.nix
    ./docker
    ./docker-compose.yml
    ./olympus
    ./result
    ./result-bin
    ./shell.nix
    ./vendor
    ./nix
    ./gin-bin
    ./scylla
  ];
in buildGoPackage rec {
  name = "scylla-unstable-${version}";
  version = "2018-07-23";
  rev = "277ad49d97dd0861b889ee7a0d8922f4549affe4";

  goPackagePath = "github.com/manveru/scylla";

  nativeBuildInputs = [ makeWrapper ];

  src = builtins.filterSource (path: type:
    (lib.all (i: i != path) ignore)
  ) ./.;

  goDeps = ./deps.nix;

  preBuild = ''
    go generate ${goPackagePath}
    go test ${goPackagePath}
  '';

  postInstall = ''
    wrapProgram $bin/bin/scylla --run "cd $src"
  '';

  meta = with stdenv.lib; {
    description = "A simple, easy to deploy Nix Continous Integration server";
    homepage = https://github.com/manveru/scylla;
    license = licenses.mit;
    maintainers = [ maintainers.manveru ];
    platforms = platforms.unix;
  };
}
