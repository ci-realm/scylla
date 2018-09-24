{ stdenv, lib, buildGoPackage, fetchFromGitHub, makeWrapper, git, nixUnstable }:
buildGoPackage rec {
  name = "scylla-unstable-${version}";
  version = "2018-07-23";
  rev = "277ad49d97dd0861b889ee7a0d8922f4549affe4";

  goPackagePath = "github.com/manveru/scylla";

  nativeBuildInputs = [ makeWrapper ];
  src = fetchGit ./.;

  goDeps = ./deps.nix;

  preBuild = ''
    go generate ${goPackagePath}
    go test ${goPackagePath}
  '';

  postInstall = ''
    wrapProgram $bin/bin/scylla \
      --prefix PATH : ${lib.makeBinPath [ git nixUnstable ]} \
      --run "cd $src"
  '';

  meta = {
    description = "A simple, easy to deploy Nix Continous Integration server";
    homepage = https://github.com/manveru/scylla;
    license = stdenv.lib.licenses.mit;
    maintainers = stdenv.lib.maintainers.manveru;
    platforms = stdenv.lib.platforms.unix;
  };
}
