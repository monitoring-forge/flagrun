# flagrun

monitoring-forgeのMackerel pluginで広く利用している [go-flags](https://github.com/jessevdk/go-flags) の共通処理をまとめたライブラリです。

## 概要

`flagrun` は、Mackerel plugin のエントリーポイントで繰り返し書くような以下の処理を1つの `Go` 関数にまとめたものです。

- ヘルプ表示（`--help` / `-h`）
- バージョン表示（`--version` / `-v`）
- 引数が必要かどうかの判定
- エラー時の終了コード返却（UNKNOWN）

## インストール

```bash
go get github.com/monigoring-forge/flagrun
```

## 使い方

`Runner` インターフェースを実装した構造体を `flagrun.Go` に渡します。

`Run` メソッドの戻り値は `(メッセージ, 終了コード)` です。終了コードが `OK` の場合、メッセージは標準出力へ出力されます。`OK` 以外の場合は標準エラー出力へ出力されます。終了コードは `os.Exit` に渡されます。

```go
package main

import (
    _ "github.com/jessevdk/go-flags"
    "github.com/monigoring-forge/flagrun"
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
        flagrun.Commit(commit),
    ))
}
```

## オプション

`flagrun.Go` では、以下の関数を使って動作をカスタマイズできます。

| 関数 | 説明 |
|------|------|
| `flagrun.Version(version string)` | バージョン表示に使用する文字列を指定します。 |
| `flagrun.Commit(commit string)` | コミットハッシュなどを指定します（デフォルト: `dev`）。 |
| `flagrun.ArgsRequired()` | コマンドライン引数を必須にします。引数がない場合は UNKNOWN で終了します。 |

## 終了コード

| 定数 | 値 | 説明 |
|------|---|------|
| `flagrun.OK` | `0` | 正常終了 |
| `flagrun.WARNING` | `1` | 警告 |
| `flagrun.CRITICAL` | `2` | 致命的エラー |
| `flagrun.UNKNOWN` | `3` | 不明なエラー（パースエラー、引数不足など） |

## ライセンス

[LICENSE](LICENSE) を参照してください。
