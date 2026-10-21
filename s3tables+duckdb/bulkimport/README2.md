# Glue への `add_files`

[前回](./README.md) の Glue への焼き直し。

## PyIceberg のセットアップ

PyIceberg を venv 環境にセットアップ

```console
$ python -m venv venv

$ python -m pip install --upgrade pip

$ pip install "pyiceberg[glue,s3fs,hive,pyarrow]"
```

## PyIcebergを用いたファイルの追加

テーブルをIceberg Catalogで作ってから、PyIceberg の `add_files` でデータ追加

```console
$ time python glue_add_files.py

real    0m23.540s
user    0m0.509s
sys     0m0.077s
```

ちょっと時間がかかる。

`glue_add_files.py`

```python
from pyiceberg.catalog import load_catalog

catalog = load_catalog(
        "glue_catalog",
        **{
            "type": "glue",
            "warehouse": "s3://kaoriya/daily_sales/",
            "s3.region": "ap-northeast-1",
            "glue.region": "ap-northeast-1",
        }
)

table = catalog.load_table('gluetest.daily_sales2')

table.add_files(file_paths=[
    "s3://kaoriya/daily_sales/01a084a0-9551-75c0-b9f4-02c7d10fdf4e.parquet",
    "s3://kaoriya/daily_sales/01a084a0-9552-7780-ade6-603d6e3238ae.parquet",
])
```

実行には ~/.aws/credentials に適切な設定が必要

`s3://kaoriya/daily_sales/metadata` 下に必要なメタデータが作られる

## DuckdBによる読み取り

```
create secret s3_secret( type s3, provider credential_chain);

attach '961554615704' as glue_catalog(type iceberg, endpoint_type 'glue', SECRET s3_secret);

select * from glue_catalog.gluetest.daily_sales2 LIMIT 5;
┌────────────┬──────────────────┬──────────────┐
│ sale_date  │ product_category │ sales_amount │
│    date    │     varchar      │    double    │
├────────────┼──────────────────┼──────────────┤
│ 2024-01-15 │ Laptop           │        900.0 │
│ 2024-01-15 │ Monitor          │        250.0 │
│ 2024-01-16 │ Laptop           │       1350.0 │
│ 2024-02-01 │ Monitor          │        300.0 │
│ 2024-02-01 │ Keyboard         │         60.0 │
└────────────┴──────────────────┴──────────────┘
```
