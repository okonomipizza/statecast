# daemon と CLI の骨格を追加する

- Priority: High
- Created: 2026-06-06
- Model: Composer 2.5
- Branch: feature/add-daemon-skeleton
- Polished: 2026-06-06

## 目的

README で定義した statecast の中核機能（daemon による状態管理と CLI による読み書き）を実装する土台を作る。以降の機能追加はこの骨格の上に積み上げる。

## 優先度根拠

High — プロジェクトの最初の実装 issue であり、他の機能はすべてこの daemon / CLI 基盤に依存する。

## 現状

- README に CLI 仕様（`run` / `list` / `update` / `get`）が定義されている
- Nix 開発環境（go-overlay）と `go.mod` のみ存在し、Go ソースコードは未実装

## 設計方針

- **言語**: Go（`go.mod` のバージョンに従う）
- **プロセス構成**: CLI 1 バイナリ + バックグラウンド daemon 1 プロセス
- **通信**: ローカル IPC（Unix ドメインソケット等）で CLI と daemon を接続する
- **状態管理**: daemon 内でエージェント ID 単位の JSON 状態をインメモリ保持する
- **エージェント識別**: 環境変数 `STATECAST_AGENT_ID` または `--agent-id` フラグ（実装時にどちらか、または両方を確定）
- **CLI フレームワーク**: サブコマンド構造を持つライブラリ（例: cobra）を検討する
- **データ構造**: エージェント状態の内部表現は最小限から始め、README の「実装を進めながら決める」方針に従う

## 完了条件

- `statecast run` で daemon が起動し、バックグラウンドでリクエストを受け付ける
- `statecast list` で登録済みエージェントの状態一覧が表示される
- `statecast update '{"key": "value"}'` で呼び出し元エージェントの状態が更新される
- `statecast get` で自身の状態が取得できる
- `statecast get --agent <agent-id>` で他エージェントの状態が取得できる
- daemon 未起動時に CLI を実行すると、わかりやすいエラーメッセージ（英語）が表示される
- 上記コマンドの動作を確認するテストが存在する（テストメッセージは日本語、モック / スタブは使わない）

## 解決方法

（未着手 — 実装完了時に記載する）
