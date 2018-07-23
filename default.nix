{ stdenv, buildGoPackage, fetchFromGitHub, runTests ? true }:

buildGoPackage rec {
  name = "scylla-unstable-${version}";
  version = "2018-07-23";
  rev = "277ad49d97dd0861b889ee7a0d8922f4549affe4";

  goPackagePath = "github.com/manveru/scylla";

  src = fetchGit ./.;
  # src = fetchFromGitHub {
  #   inherit rev;
  #   owner = "manveru";
  #   repo = "scylla";
  #   sha256 = "1skjli6zpb3vm17glg3w65j4bizza0nvmmbh1gwj9gsjd89cqfri";
  # };

  goDeps = ./deps.nix;

  preBuild = ''
    sleep 30
    go generate ${goPackagePath}
  '' + (stdenv.lib.optionalString runTests "go test ${goPackagePath}");

  meta = {
    description = "A simple, easy to deploy Nix Continous Integration server";
    homepage = https://github.com/manveru/scylla;
    license = stdenv.lib.licenses.mit;
    maintainers = stdenv.lib.maintainers.manveru;
    platforms = stdenv.lib.platforms.unix;
  };
}
