# flagrun

monitoring-forgeのMackerel pluginで広く利用している [go-flags](https://github.com/jessevdk/go-flags) の共通処理をまとめたライブラリです。

## 概要

`flagrun` は、Mackerel plugin のエントリーポイントで繰り返し書くような以下の処理を1つの `Go` 関数にまとめたものです。

- ヘルプ表示（`--help` / `-h`）
- バージョン表示（`--version` / `-v`）
- 引数が必要かどうかの判定
- エラー時の終了コード返却（UNKNOWN）

また、用途に応じて以下の3種類のインターフェースを提供します。

- `Runner[T]` — 汎用的な `(メッセージ, 終了コード)` を返す形式
- `Checker` — [mackerelio/checkers](https://github.com/mackerelio/checkers) の `*checkers.Checker` を返す形式
- `Shipper` — 何も返さず、副作用でメトリクスなどを送信する形式

## インストール

```bash
go get github.com/monitoring-forge/flagrun
```

## 使い方

### `Runner[T]` — 汎用的な実行

`Runner[T]` インターフェースを実装した構造体を `flagrun.Go` に渡します。

`Run` メソッドの戻り値は `(メッセージ, 終了コード)` です。終了コードが `OK` の場合、メッセージは標準出力へ出力されます。`OK` 以外の場合は標準エラー出力へ出力されます。終了コードは `os.Exit` に渡されます。

```go
package main

import (
    "github.com/monitoring-forge/flagrun"
)

type Opt struct {
    Host string `short:"H" long:"host" default:"localhost" description:"Target host"`
    Port int    `short:"p" long:"port" default:"8080" description:"Target port"`
    Version bool `short:"v" long:"version" description:"Show version"`
}

func (p *Opt) Run(args []string) (string, int) {
    // Mackerel plugin のメイン処理を実装
    return "ok\t1", flagrun.OK
}

func main() {
    opt := &Opt{}
    os.Exit(flagrun.Go(
        opt,
        flagrun.Version(version),
    ))
}
```

### `Checker` — mackerelio/checkers を使う

`Checker` インターフェースを実装した構造体を `flagrun.Check` に渡します。

`Run` メソッドの戻り値は `*checkers.Checker` です。`Checker.String()` の結果を標準出力へ出力し、`Checker.Status` を終了コードとして返します。

```go
package main

import (
    "github.com/mackerelio/checkers"
    "github.com/monitoring-forge/flagrun"
)

type Opt struct {
    Host string `short:"H" long:"host" default:"localhost" description:"Target host"`
    Version bool `short:"v" long:"version" description:"Show version"`
}

func (p *Opt) Run(args []string) *checkers.Checker {
    return checkers.Ok("service is reachable")
}

func main() {
    opt := &Opt{}
    os.Exit(flagrun.Check(
        opt,
        flagrun.Version(version),
    ))
}
```

### `Shipper` — 副作用だけで実行

`Shipper` インターフェースを実装した構造体を `flagrun.Ship` に渡します。

`Run` メソッドは戻り値を持ちません。メトリクスの送信など、副作用だけを行いたい場合に使います。終了コードは常に `OK` を返します。

```go
package main

import (
    "github.com/monitoring-forge/flagrun"
)

type Opt struct {
    Host string `short:"H" long:"host" default:"localhost" description:"Target host"`
    Version bool `short:"v" long:"version" description:"Show version"`
}

func (p *Opt) Run(args []string) {
    // 副作用でメトリクスを送信
}

func main() {
    opt := &Opt{}
    os.Exit(flagrun.Ship(
        opt,
        flagrun.Version(version),
    ))
}
```

## オプション

| `flagrun.Go` / `flagrun.Check` / `flagrun.Ship` では、以下の関数を使って動作をカスタマイズできます。

| 関数 | 説明 |
|------|------|
| `flagrun.Version(version string)` | バージョン表示に使用する文字列を指定します。 |
| `flagrun.Commit(commit string)` | コミットハッシュなどを指定します（デフォルト: `dev`）。 |
| `flagrun.ArgsRequired()` | コマンドライン引数を必須にします。引数がない場合は UNKNOWN で終了します。 |
| `flagrun.AlwaysStdout()` | `Run` の戻り値を、終了コードに関係なく標準出力へ出力します。`flagrun.Check` では常に標準出力へ出力されるため、このオプションは不要です。 |

## 終了コード

| 定数 | 値 | 説明 |
|------|---|------|
| `flagrun.OK` | `0` | 正常終了 |
| `flagrun.WARNING` | `1` | 警告 |
| `flagrun.CRITICAL` | `2` | 致命的エラー |
| `flagrun.UNKNOWN` | `3` | 不明なエラー（パースエラー、引数不足など） |

## ライセンス

[LICENSE](LICENSE) を参照してください。
