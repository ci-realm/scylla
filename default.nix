{ stdenv, buildGoPackage, fetchFromGitHub, runTests ? true }:

buildGoPackage rec {
  name = "scylla-unstable-${version}";
  version = "2018-07-21";
  rev = "1d6a7ec1c5753cfa4bf1c158770437815a7f9241";

  goPackagePath = "github.com/manveru/scylla";

  src = stdenv.lib.cleanSource ./.;
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
