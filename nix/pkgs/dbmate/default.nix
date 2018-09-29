{ stdenv, lib, buildGoPackage, fetchFromGitHub }:

buildGoPackage rec {
  name = "dbmate-${version}";
  version = "1.4.1";

  goPackagePath = "github.com/amacneil/dbmate";

  src = fetchFromGitHub {
    owner = "amacneil";
    repo = "dbmate";
    rev = "v${version}";
    sha256 = "0s3l51kmpsaikixq1yxryrgglzk4kfrjagcpf1i2bkq4wc5gyv5d";
  };

  goDeps = ./deps.nix;

  meta = {
    description = "Database migration tool";
    homepage = https://dbmate.readthedocs.io;
    license = stdenv.lib.licenses.mit;
    maintainers = stdenv.lib.maintainers.manveru;
    platforms = stdenv.lib.platforms.unix;
  };
}
