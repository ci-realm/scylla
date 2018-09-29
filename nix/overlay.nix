self: super:
let
  manveru-nur-packages = fetchTarball {
    url = "https://github.com/manveru/nur-packages/archive/1f492c51eedd1f6abac7b753a7e56dfe46be1860.tar.gz";
    sha256 = "0m5aywc2lnmycj6vvaa7d9p3wj763crpjrwqqylpq6j7bviniq4n";
  };
in {
  git-info = (self.callPackage "${manveru-nur-packages}/default.nix" {}).lib.git-info;
  go = super.go_1_11;
  decrypt-ejson = path:
    (builtins.fromJSON (builtins.readFile
      (super.runCommand "ejson" { nativeBuildInputs = with super; [ ruby ejson jq ]; } ''
        export EJSON_KEYDIR='${/opt/ejson/keys}'
        ejson decrypt '${path}' | ruby ${./ejson.rb} > $out
      '')));
  makeDeploymentID = sha:
    super.lib.removeSuffix "\n" (builtins.readFile (
      super.runCommand "deployment_id" {sha = sha;} ''
        ${super.ruby}/bin/ruby -r securerandom -e 'print ENV["sha"][0...8] + "-#{SecureRandom.hex(4)}"' > $out
      ''
    ));
  tagFromGit = self.git-info "git rev-parse --verify HEAD";
  dbmate = self.callPackage ./pkgs/dbmate {};
}
