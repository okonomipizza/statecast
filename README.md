# statecast

AI エージェント間で状態を共有するためのローカル daemon。

各エージェントは CLI 経由で自分の状態を読み書きし、他のエージェントの状態を参照できる。daemon プロセスは 1 つで、管理対象の状態はエージェント単位で分離して保持する。

将来的なクラウド連携も視野に入れているが、データ構造の詳細は実装を進めながら決める。

## 背景

複数の AI エージェントが並行して動くとき、互いの作業状況や判断結果を共有する手段が必要になる。statecast はその共有レイヤをローカル環境向けに提供する。

## アーキテクチャ

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Agent A    │     │  Agent B    │     │  Agent C    │
│  (CLI)      │     │  (CLI)      │     │  (CLI)      │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                    ┌──────▼──────┐
                    │  statecast  │
                    │  (daemon)   │
                    │             │
                    │  Agent A ──►│ state A
                    │  Agent B ──►│ state B
                    │  Agent C ──►│ state C
                    └─────────────┘
```

- **daemon**: 1 プロセスがインメモリで全エージェントの状態を管理する
- **CLI**: エージェントが daemon に対して読み取り・更新を行う
- **スコープ**: 現時点はローカル完結

## データ形式

状態は JSON 形式の文字列として扱う。エージェントが CLI から直接渡す。

```bash
statecast update --name "Agent A" '{"key": "value"}'
```

ファイルからの読み込みは将来対応の候補だが、現時点では CLI 引数による文字列指定のみを想定する。

## CLI

### daemon の起動・停止

```bash
statecast start
statecast stop
```

`start` はバックグラウンドで daemon を起動し、エージェントからのリクエストを受け付ける。`stop` は起動中の daemon に SIGTERM を送り、停止する。

### エージェントの登録

```bash
statecast register --name "Agent A"
```

- `--name`: エージェント名（必須、一意）

登録後、name が表示される。以降の `update` / `get` では `--name` でこの名前を指定する。

### 状態の一覧

```bash
statecast list
```

登録済みエージェント名の一覧を表示する。

```
# name
Agent A
Agent B
```

### 状態の更新

```bash
statecast update --name "Agent A" '{"key": "value"}'
```

登録済みエージェントの状態を、指定した JSON 文字列で更新する。

### 状態の読み取り

```bash
statecast get --name "Agent A"
```

登録済みエージェントの状態を取得する。

## エージェントの識別

エージェントは `register` で name を登録する。name は一意。`update` / `get` では `--name` フラグで対象エージェントを指定する。

## 想定する利用フロー

1. `statecast start` で daemon を起動する
2. Agent A が `statecast register --name "Agent A"` で登録する
3. Agent A が `statecast update --name "Agent A"` で自身の状態を書き込む
4. Agent B が `statecast get --name "Agent A"` で Agent A の状態を読み取る
5. Agent B が `statecast register --name "Agent B"` で自身を登録し、`statecast update --name "Agent B"` で自身の状態を更新する

## スコープ外（現時点）

- ファイルからの状態読み込み
- クラウド連携
- 状態の永続化（再起動で失われるインメモリ管理を前提）

## インストール

最新版を `~/.local/bin` にインストールする:

```bash
curl -fsSL https://raw.githubusercontent.com/okonomipizza/statecast/master/scripts/install.sh | bash
```

`PATH` に `~/.local/bin` を追加する:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

特定バージョンを入れる:

```bash
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/okonomipizza/statecast/master/scripts/install.sh | bash
```

インストール先を変える（`/usr/local/bin` など）:

```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/okonomipizza/statecast/master/scripts/install.sh | sudo bash
```

### 手動ダウンロード

[Releases](https://github.com/okonomipizza/statecast/releases) から `statecast_<version>_<os>_<arch>.tar.gz` を取得し、展開したバイナリを PATH の通ったディレクトリへ置く。`checksums.txt` で SHA256 を確認できる。

### Nix

flake からバイナリを入れる:

```bash
nix profile install github:okonomipizza/statecast
```

## 開発

Go で実装する。開発環境は Nix flake と [go-overlay](https://github.com/purpleclay/go-overlay) で提供する。

```bash
nix develop
```

### ビルド

```bash
make build
```

`./statecast` バイナリが生成される。

Nix でバイナリをビルドする場合:

```bash
nix build
./result/bin/statecast --help
```

`go.mod` / `go.sum` を更新したら、Nix ビルド用の依存マニフェストも同期する:

```bash
nix develop -c govendor
```

生成された `govendor.toml` をコミットする。CI では `govendor --check` でドリフトを検出できる。

### テスト

```bash
make test
make lint
```

fuzz テスト（任意）:

```bash
go test ./internal/store -fuzz=FuzzRegisterName -fuzztime=30s
go test ./internal/store -fuzz=FuzzPutStateValidJSON -fuzztime=30s
```

### 開発環境

利用する Go のバージョンは `go.mod` の `go` ディレクティブで指定する（現在: 1.25）。go-overlay が対応するツールチェーンと開発ツールを提供する。

dev shell には `go.withDefaultTools` が含まれる（`go`、`gopls`、`golangci-lint` など）。

Go のモジュールキャッシュはプロジェクト内の `.go/` に置く。
