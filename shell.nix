with import <nixpkgs>{};
stdenv.mkDerivation rec {
    name = "LED";
    buildInputs =  [ go ];
    shellHook = ''
        go run *.go
        exit
    '';

    GOPATH="/home/ajanse/.go";
}
