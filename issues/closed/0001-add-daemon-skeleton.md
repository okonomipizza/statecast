# daemon と CLI の骨格を追加する

- Priority: High
- Created: 2026-06-06
- Completed: 2026-06-06
- Model: Composer 2.5
- Branch: feature/add-daemon-skeleton
- Polished: 2026-06-06

## 目的

README で定義した statecast の中核機能（daemon による状態管理と CLI による読み書き）を実装する土台を作る。以降の機能追加はこの骨格の上に積み上げる。

## 優先度根拠

High — プロジェクトの最初の実装 issue であり、他の機能はすべてこの daemon / CLI 基盤に依存する。

## 現状（着手時）

- README に CLI 仕様（`run` / `list` / `update` / `get`）が定義されていた
- Nix 開発環境（go-overlay）と `go.mod` のみ存在し、Go ソースコードは未実装だった

## 設計方針

- **言語**: Go（`go.mod` のバージョンに従う）
- **プロセス構成**: CLI 1 バイナリ + バックグラウンド daemon 1 プロセス
- **通信**: Unix ドメインソケット上の HTTP API で CLI と daemon を接続する
- **状態管理**: daemon 内でエージェント名単位の JSON 状態をインメモリ保持する
- **エージェント識別**: `register --name` で登録し、`update` / `get` では `--name` で指定する（name は一意）
- **CLI フレームワーク**: Cobra
- **データ構造**: 状態は `json.RawMessage` で保持し、スキーマは daemon 側では固定しない

## 完了条件

- [x] `statecast start` で daemon が起動し、バックグラウンドでリクエストを受け付ける
- [x] `statecast stop` で daemon を停止できる
- [x] `statecast register --name` でエージェントを登録できる
- [x] `statecast list` で登録済みエージェント名の一覧が表示される
- [x] `statecast update --name` で登録済みエージェントの状態が更新される
- [x] `statecast get --name` で登録済みエージェントの状態が取得できる
- [x] daemon 未起動時に CLI を実行すると、わかりやすいエラーメッセージ（英語）が表示される
- [x] 上記コマンドの動作を確認するテストが存在する（テストメッセージは日本語、モック / スタブは使わない）

## 解決方法

- Cobra で `start` / `stop` / `register` / `list` / `update` / `get` サブコマンドを実装
- daemon は Unix ドメインソケット上の HTTP API（`/health`, `/v1/agents`）で CLI と通信
- 状態は `internal/store` のインメモリ Store でエージェント名単位に保持（`json.RawMessage`）
- エージェントは `register --name` で明示登録。同名の再登録は拒否する
- `statecast start` はバックグラウンドプロセスとして daemon を起動（`--foreground` は内部用）
- `statecast stop` は PID ファイルから SIGTERM を送り、graceful shutdown する
- ランタイムファイルは `~/.local/statecast/`（`STATECAST_RUNTIME_DIR` で上書き可能）
- 統合テストは実 daemon + 実クライアントで検証（モック / スタブ不使用）
- エージェント名と JSON 状態に fuzzing / PBT テストを追加（`internal/store`）
- リクエストボディは 1 MiB 上限（`readLimitedBody`）
- PID ファイルは `0o600`（所有者のみ読み書き）
- daemon 起動待ちは `IsRunning` ポーリング（最大 5 秒）
