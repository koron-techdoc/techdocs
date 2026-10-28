# Glue Iceberg カタログとは

## 概要

Glue Iceberg カタログとは、
カタログを用意して、
それとデータ保存用のS3バケツにアクセス可能なIAMロールのクレデンシャルを渡しておけば、
DuckDBなどからほぼフルコンロール可能なデータベースインスタンス。

現在のIcebergには大きく分けて3世代のカタログがある。
第1世代はHadoop/Hiveベースのカタログ。
第2世代はクラウド各種ベンダーが参入した際のカタログ。
最新の第3世代RESTカタログ。
Glueのカタログは第2世代のもの。
Icebergクライアント実装は、
これら3つの世代のカタログにアクセスできると期待できる。

Glue Iceberg カタログを使う際は、
まずカタログを作り、
次にデータベースを作り、
最後にテーブルを作る必要がある。
ここまでの操作はAWSコンソールから行える。

しかし適切なIAMロールを設定していれば、データベースとテーブルの作成はDuckDBなどのクライアントからでも行える。
パーティション設定の観点で言うと、テーブルの作成はDuckDBから行った方が良い。
コンソールからではパーティションが設定できない。
逆にデータベースの作成は、
レコードを追加した際の保存場所(`Location`)を指定するのであれば、
コンソールから行う方が良い。

以下では、具体的な操作を紹介する。

## DuckDB からの操作

### モジュールロード

DuckDBから Iceberg や AWS にアクセスするには以下のようにプラグインのインストール&ロードが必要とされている。

```
LOAD iceberg;
LOAD aws;
```

しかし DuckDB 1.5.0 や 2.0.0 (Windows版) で試した限りでは、
特に明示しなくても自動的にインストール&ロードされた。
このあたりの違いがどこから来るのかは不明。

### 接続

DuckDBに `~/.aws/credentials` を参照させる。 
`credential_chain` を指定することでそれらが参照される。
特定のプロファイルを使う場合は `PROFILE` を追加指定する。

```sql
CREATE SECRET s3_secret(TYPE s3, PROVIDER credential_chain);
```

名称には `s3_secret` を指定してる。
省略すればデフォルトシークレットの指定となる。
複数のシークレットが設定可能なのは、アクセス対象となるリソース毎に異なるシークレットを指定できるという、DuckDBの機能に基づく。

以下でカタログに接続する。

```
ATTACH '{カタログ名(デフォルトはアカウントID)}' AS glue_catalog(TYPE iceberg, ENDPOINT_TYPE 'glue', SECRET s3_secret);
```


`SECRET` に指定した `s3_secret` は前ステップで与えた名前であり、
省略した場合はデフォルトシークレットを使用する。

必要に応じて `USE {カタログ名}.{データベース名};` すると楽

### テーブルの操作

メタデータを保存するS3バケットの指定 `WITH (location = 's3://.../')` が必要。
指定したパスには自動的に `/metadata/` が追加される。

```sql
CREATE TABLE daily_sales(
    sale_date date, 
    product_category string, 
    sales_amount double)
    PARTITIONED BY (month(sale_date))
    WITH (location = 's3://.../.../.../');
```

DuckDBからはテーブルを消せない。 
エラーメッセージは `PurgeRequested` オプションに対応していないとのこと。
コンソールから消す必要がある。S3上のメタデータの削除も忘れずに。

パーティションを指定してテーブルを作成できる。
コンソールからは確認できないが、
Icebergカタログのメタデータには記録されている。

言い換えるとコンソールから作ったテーブルにはパーティションが設定できず、
最適なアクセスにはならないことに注意が必要。

### データベースの作成

データベース(スキーマ)自体もDuckDBから作れる。

```sql
CREATE SCHEMA {カタログ名}.{データベース=スキーマ名};
```

コンソールからデータベースを作った際の Location の指定は、
DuckDBからINSERTを実行した差異のデータの格納先になる。
これは `CREATE SCHEMA` からは指定できないので、
通常はデータベースはコンソールから作った方が良い。

## PyIceberg でファイル追加

PyIcebergを使って、S3上の既存parquetファイルを追加できる。

サンプルスクリプト:
AWS用の credentials は設置してあるが、PyIceberg 用の設定は特にはしていない。

```console
$ cat glue_add_files.py
from pyiceberg.catalog import load_catalog

catalog = load_catalog(
        "glue_catalog",
        **{
            "type": "glue",
            "s3.region": "ap-northeast-1",
            "glue.region": "ap-northeast-1",
        }
)

table = catalog.load_table('{データベース}.{テーブル}')

table.add_files(file_paths=[
    "s3://foo/bar/01a084a0-9551-75c0-b9f4-02c7d10fdf4e.parquet",
    "s3://foo/bar/01a084a0-9552-7780-ade6-603d6e3238ae.parquet",
])

$ time python ./glue_add_files.py

real    0m28.245s
user    0m0.511s
sys     0m0.119s
```

結構時間がかかる。

parquetファイルと、テーブルのパーティション設定が食い違っていると、
正しいデータが取得できなくなるので注意。

テーブル作成時に指定した `location` とは別のS3バケットのファイルも登録できる。
あちらはあくまでもメタデータの置き場。

1つのファイルを複数の異なるテーブルに登録することもできる。
