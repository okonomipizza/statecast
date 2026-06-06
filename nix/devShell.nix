{ pkgs, go }:

pkgs.mkShell {
  packages = [
    go.withDefaultTools
  ];

  env = {
    # プロジェクトローカルに Go モジュールキャッシュを置く
    GOMODCACHE = ".go/pkg/mod";
    GOCACHE = ".go/cache";
  };

  shellHook = ''
    mkdir -p .go/pkg/mod .go/cache
    echo "Go $(go version | cut -d' ' -f3)"
  '';
}
