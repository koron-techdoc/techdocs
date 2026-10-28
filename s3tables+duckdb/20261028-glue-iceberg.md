# Glue Icebergの実験

## 実験ログ

DuckDBに `~/.aws/credentials` を参照させる。 
`credential_chain` を指定することでそれらが参照される。
特定のプロファイルを使う場合は `PROFILE` を追加指定する。

```
CREATE SECRET s3_secret(TYPE s3, PROVIDER credential_chain);
```

カタログに接続する。
`SECRET` に指定した `s3_secret` は前ステップで与えた名前。

```
ATTACH '{カタログ名(デフォルトはアカウントID)}' AS glue_catalog(TYPE iceberg, ENDPOINT_TYPE 'glue', SECRET s3_secret);
```

必要に応じて `USE {カタログ名}.{データベース名};` すると楽

DuckDBからテーブルを作れた。
メタデータを保存するS3バケットの指定 `WITH (location = 's3://.../')` が必要。
指定したパスには自動的に `/metadata/` が追加される。

```sql
CREATE TABLE daily_sales(
    sale_date date, 
    product_category string, 
    sales_amount double)
    WITH (location = 's3://.../.../.../');
```

DuckDBからはテーブルを消せない。 
エラーメッセージは `PurgeRequested` オプションに対応していないとのこと。

コンソールから消す必要あり。S3上のメタデータも忘れずに。

DuckDBからパーティションを指定してテーブルを作れた。
コンソールからは確認できないが、
Icebergカタログのメタデータには記録されている。

```sql
CREATE TABLE daily_sales(
    sale_date date, 
    product_category string, 
    sales_amount double)
    PARTITIONED BY (month(sale_date))
    WITH (location = 's3://.../.../.../');
```

言い換えるとコンソールから作ったテーブルにはパーティションが設定できず、
最適なアクセスにはならない。
面倒でもDuckDBなどからやる必要がある。


PyIcebergを使って、S3上の既存parquetファイルを追加する。

```console
$ cat glue_add_files3.py
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

$ time python ./glue_add_files3.py

real    0m28.245s
user    0m0.511s
sys     0m0.119s
```

parquetファイルと、テーブルのパーティション設定が食い違っていると、
正しいデータが取得できなくなるので注意。

実はデータベース(スキーマ)自体をDuckDBから作れるぞ。

```sql
CREATE SCHEMA {カタログ名}.{データベース=スキーマ名};
```

コンソールからデータベースを作った際の Location の指定は、
DuckDBからINSERTを実行した差異のデータの格納先になる。
これは `CREATE SCHEMA` からは指定できないので、
通常はデータベースはコンソールから作った方が良い。

つまりカタログを用意して、それにアクセス可能なcredentialを渡しておけば、
ほぼほぼフルコントロール可能なデータベースインスタンスってこと。
