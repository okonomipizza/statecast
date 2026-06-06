# CHANGES

## develop

- [ADD] daemon と CLI の骨格を実装する（`start` / `stop` / `register` / `list` / `update` / `get`）
- [ADD] エージェント名と JSON 状態の fuzzing / PBT テストを追加する
- [ADD] `Makefile` に build / test / lint ターゲットを追加する
- [ADD] GitHub Releases 向けの release ワークフローと CI を追加する
- [ADD] バイナリインストール用スクリプト `scripts/install.sh` を追加する

### misc

- [UPDATE] README にインストール手順（バイナリ / Nix）を追記する
- [UPDATE] README に CLI 仕様と開発手順を追記する
- [UPDATE] issue 0001 を完了し issues/closed へ移動する
