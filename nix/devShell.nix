{ pkgs, go }:

pkgs.mkShell {
  packages = [
    go.withDefaultTools
  ];

  shellHook = ''
    mkdir -p .go/pkg/mod .go/cache
    export GOMODCACHE="$PWD/.go/pkg/mod"
    export GOCACHE="$PWD/.go/cache"
    echo "Go $(go version | cut -d' ' -f3)"
  '';
}
