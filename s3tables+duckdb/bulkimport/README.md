# Iceberg (R2)へのBulk import の実験

## 実験手順

[過去の手順](../20260624-s3tables-trial.md) に従い、予めデータを作っておく

### データ作成

追加する parquet ファイルを作成。2024年3月分のデータをでっちあげる。

```sql
CREATE TABLE daily_sales(
    sale_date date, 
    product_category string, 
    sales_amount double);

INSERT INTO daily_sales VALUES
    (DATE '2024-03-11', 'Laptop', 1400.0),
    (DATE '2024-03-12', 'Monitor', 400.0),
    (DATE '2024-03-12', 'Keyboard', 200.0),
    (DATE '2024-03-12', 'Mouse', 100.0),
    (DATE '2024-03-13', 'Desktop', 2000.0),
    (DATE '2024-03-13', 'Keyboard', 75.0),
    (DATE '2024-03-13', 'Mouse', 20.0),
    (DATE '2024-03-14', 'Laptop', 1750.0),
    (DATE '2024-03-14', 'Mouse', 150.0);

COPY daily_sales TO 'daiy_sales-202403.parquet' (FORMAT 'parquet');
```

内容確認

```console
$ duckdb daily_sales-202403.parquet -c "show all tables; pragma table_info('file')"
┌────────────────────┬─────────┬────────────────────┬─────────────────────────────────────────────┬─────────────────────────┬───────────┐
│      database      │ schema  │        name        │                column_names                 │      column_types       │ temporary │
│      varchar       │ varchar │      varchar       │                  varchar[]                  │        varchar[]        │  boolean  │
├────────────────────┼─────────┼────────────────────┼─────────────────────────────────────────────┼─────────────────────────┼───────────┤
│ daily_sales-202403 │ main    │ daily_sales-202403 │ [sale_date, product_category, sales_amount] │ [DATE, VARCHAR, DOUBLE] │ false     │
│ daily_sales-202403 │ main    │ file               │ [sale_date, product_category, sales_amount] │ [DATE, VARCHAR, DOUBLE] │ false     │
└────────────────────┴─────────┴────────────────────┴─────────────────────────────────────────────┴─────────────────────────┴───────────┘
┌───────┬──────────────────┬─────────┬─────────┬────────────┬─────────┐
│  cid  │       name       │  type   │ notnull │ dflt_value │   pk    │
│ int32 │     varchar      │ varchar │ boolean │  varchar   │ boolean │
├───────┼──────────────────┼─────────┼─────────┼────────────┼─────────┤
│     0 │ sale_date        │ DATE    │ false   │ NULL       │ false   │
│     1 │ product_category │ VARCHAR │ false   │ NULL       │ false   │
│     2 │ sales_amount     │ DOUBLE  │ false   │ NULL       │ false   │
└───────┴──────────────────┴─────────┴─────────┴────────────┴─────────┘
```

そうしてできたデータは [daily_sales-202603.parquet](./daily_sales-202603.parquet)

### R2へのアップロード

`daily_sales-202403.parquet` を、利用しているR2ストレージの
`__r2_data_catalog/{ネームスペースID}/{テーブルUUID}` (=テーブルのためのディレクトリ)以下へ、ダッシュボードからアップロード。


### PyIceberg のセットアップ

PyIceberg を venv 環境にセットアップ

```console
$ python -m venv venv

$ pip install --upgrade pip

$ pip install "pyiceberg[s3fs,hive,pyarrow]"
```

### Import前の確認

DuckDB でR2に接続して、既存データとこれから追加する領域の確認。

```
CREATE SECRET r2_secret(TYPE ICEBERG, TOKEN '{R2 APIトークン}');

ATTACH '{ウェアハウス名}' as r2 ( TYPE ICEBERG, ENDPOINT 'https://catalog.cloudflarestorage.com/{アカウント名}/{カタログ名}');

USE r2.default;

SELECT product_category, COUNT(*) as units_sold, SUM(sales_amount) as total_revenue, AVG(sales_amount) as average_price
    FROM daily_sales
    WHERE sale_date BETWEEN DATE '2024-03-01' and DATE '2024-03-31'
    GROUP BY product_category ORDER BY total_revenue DESC;
┌──────────────────┬────────────┬───────────────┬───────────────┐
│ product_category │ units_sold │ total_revenue │ average_price │
│     varchar      │   int64    │    double     │    double     │
├──────────────────┼────────────┼───────────────┼───────────────┤
│ Laptop           │          2 │        2250.0 │        1125.0 │
│ Monitor          │          2 │         675.0 │         337.5 │
│ Keyboard         │          1 │          60.0 │          60.0 │
│ Mouse            │          1 │          25.0 │          25.0 │
└──────────────────┴────────────┴───────────────┴───────────────┘

SELECT product_category, COUNT(*) as units_sold, SUM(sales_amount) as total_revenue, AVG(sales_amount) as average_price
    FROM daily_sales
    WHERE sale_date BETWEEN DATE '2024-03-01' and DATE '2024-03-31'
    GROUP BY product_category ORDER BY total_revenue DESC;
┌──────────────────┬────────────┬───────────────┬───────────────┐
│ product_category │ units_sold │ total_revenue │ average_price │
│     varchar      │   int64    │    double     │    double     │
└──────────────────┴────────────┴───────────────┴───────────────┘
```

### Import実行

[`python add_files.py`](./add_files.py) を実行する。

期待しない結果になったら [`python rollback.py`](./rollback.py) を実行して巻き戻せる。

### Import後の確認

DuckDB で確認。
前のステップでR2との接続を確立したDuckDBセッションをそのまま流用できる。

```
SELECT product_category, COUNT(*) as units_sold, SUM(sales_amount) as total_revenue, AVG(sales_amount) as average_price
    FROM daily_sales
    WHERE sale_date BETWEEN DATE '2024-03-01' and DATE '2024-03-31'
    GROUP BY product_category ORDER BY total_revenue DESC;
┌──────────────────┬────────────┬───────────────┬───────────────┐
│ product_category │ units_sold │ total_revenue │ average_price │
│     varchar      │   int64    │    double     │    double     │
├──────────────────┼────────────┼───────────────┼───────────────┤
│ Laptop           │          2 │        3150.0 │        1575.0 │
│ Desktop          │          1 │        2000.0 │        2000.0 │
│ Monitor          │          1 │         400.0 │         400.0 │
│ Keyboard         │          2 │         275.0 │         137.5 │
│ Mouse            │          3 │         270.0 │          90.0 │
└──────────────────┴────────────┴───────────────┴───────────────┘
```

### Import前後の inspect 結果

-   [Import前](./r01-before.txt)
-   [Import後](./r02-after-add_files.txt)

<details>
<summary>差分</summary>

```diff
$ diff -u bulkimport/r01-before.txt bulkimport/r02-after-add_files.txt
--- bulkimport/r01-before.txt   2026-09-16 15:04:05.829230400 +0900
+++ bulkimport/r02-after-add_files.txt  2026-09-16 15:10:08.684626300 +0900
@@ -66,18 +66,18 @@
       │     └─ Schema ID: 0
       └─ daily_sales
          ├─ Identifier: default.daily_sales
-         ├─ Metadata: (location: s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE UUID}/metadata/00005-01a0a8ce-94cb-73d1-812a-4be0ff22120c.gz.metadata.json)
+         ├─ Metadata: (location: s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE UUID}/metadata/00008-01a0a8d5-d25a-7cd2-9548-a7dacb59fb1b.gz.metadata.json)
          │  ├─ Version: V2
          │  ├─ Table UUID: {TABLE UUID}
          │  ├─ Location: s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE UUID}
-         │  ├─ Last Updated Millis: 1789538505931 (2026-09-16T15:01:45+09:00)
+         │  ├─ Last Updated Millis: 1789538978743 (2026-09-16T15:09:38+09:00)
          │  ├─ Last Column ID: 3
          │  ├─ Current Schema ID: 0
          │  ├─ Default Partition Spec: 0
-         │  ├─ Current Snapshot ID: 3382049451819024693
+         │  ├─ Current Snapshot ID: 6230909222005612047
          │  ├─ Properties (1):
          │  │  └─ schema.name-mapping.default: [{"names":["sale_date"],"field-id":1},{"names":["product_category"],"field-id":2},{"names":["sales_amount"],"field-id":3}]
-         │  └─ Last Sequence Number: 3
+         │  └─ Last Sequence Number: 5
          ├─ Current Schema: (ID: 0)
          │  ├─ [1] sale_date: date (optional)
          │  ├─ [2] product_category: string (optional)
@@ -85,11 +85,42 @@
          ├─ Partition Spec: (ID: 0)
          │  └─ month_sale_date_1 (Field ID:1000) : month(sale_date)
          └─ Current Snapshot:
-            ├─ Snapshot ID: 3382049451819024693
-            ├─ Sequence Number: 1
-            ├─ Timestamp MS: 1788931511998 (2026-09-09T14:25:11+09:00)
-            ├─ Manifest List: s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE UUID}/metadata/snap-3382049451819024693-a0d8c37b-a52c-4487-8500-7e1153f610a7.avro
-            │  └─ [0] Manifest
+            ├─ Snapshot ID: 6230909222005612047
+            ├─ Parent Snapshot ID: 64342069251976
+            ├─ Sequence Number: 5
+            ├─ Timestamp MS: 1789538978743 (2026-09-16T15:09:38+09:00)
+            ├─ Manifest List: s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE UUID}/metadata/snap-6230909222005612047-0-1d8aa9c6-bf55-483d-b64c-b8089c1ad163.avro
+            │  ├─ [0] Manifest
+            │  │  ├─ Version: 2
+            │  │  ├─ File Path: s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE UUID}/metadata/1d8aa9c6-bf55-483d-b64c-b8089c1ad163-m0.avro
+            │  │  ├─ Length: 4592
+            │  │  ├─ Partition Spec ID: 0
+            │  │  ├─ Snapshot ID: 6230909222005612047
+            │  │  ├─ Added Data Files: 1
+            │  │  ├─ Existing Data Files: 0
+            │  │  ├─ Added Rows: 9
+            │  │  ├─ Existing Rows: 0
+            │  │  ├─ Sequence Number: 5
+            │  │  ├─ Min Sequence Num: 5
+            │  │  └─ Manifest Entries:
+            │  │     └─ [0] Manifest Entry
+            │  │        ├─ Status: 1:ADDED
+            │  │        ├─ Snapshot ID: 6230909222005612047
+            │  │        ├─ Sequence Num: 5
+            │  │        ├─ File SequenceNum: 5
+            │  │        └─ Data File
+            │  │           ├─ Content Type: Data
+            │  │           ├─ File Path: s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE UUID}/data/daily_sales-202403.parquet
+            │  │           ├─ File Format: PARQUET
+            │  │           ├─ Partition: map[1000:650]
+            │  │           ├─ Count: 9
+            │  │           ├─ File Size Bytes: 612
+            │  │           ├─ Column Sizes: map[1:47 2:95 3:76]
+            │  │           ├─ Value Counts: map[1:9 2:9 3:9]
+            │  │           ├─ Null Value Counts: map[1:0 2:0 3:0]
+            │  │           ├─ Lower Bound Values: map[1:[81 77 0 0] 2:[68 101 115 107 116 111 112] 3:[0 0 0 0 0 0 52 64]]
+            │  │           └─ Upper Bound Values: map[1:[84 77 0 0] 2:[77 111 117 115 101] 3:[0 0 0 0 0 64 159 64]]
+            │  └─ [1] Manifest
             │     ├─ Version: 2
             │     ├─ File Path: s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE UUID}/metadata/48151aa4-aa78-4814-8dd2-81ffe32b6124-m0.avro
             │     ├─ Length: 2765
@@ -135,12 +166,12 @@
             ├─ Summary:
             │  ├─ Operation: append
             │  └─ Properties (8):
-            │     ├─ added-data-files: 2
+            │     ├─ added-data-files: 1
+            │     ├─ added-files-size: 612
             │     ├─ added-records: 9
-            │     ├─ deleted-data-files: 0
-            │     ├─ deleted-records: 0
-            │     ├─ total-data-files: 2
+            │     ├─ changed-partition-count: 1
+            │     ├─ total-data-files: 3
             │     ├─ total-delete-files: 0
             │     ├─ total-position-deletes: 0
-            │     └─ total-records: 9
+            │     └─ total-records: 18
             └─ Schema ID: 0
```

</details>

Manifest Listが変更になり、新規のManifest (Entry 1個) と最初のManifest (Entry 2個)が格納されている。

その他はスナップショットIDなど、Metadata の当然変わるよねというところが変わっている。

## まとめ

-   実験の最中に何度かロールバックしたので、meta/ に多くのメタデータができてしまった
-   メタデータ群がオブジェクトストレージ上に上手く格納されている
-   変更が、最小限のオブジェクトの変更で表現されている
-   R2ではデータファイルの置き場所に、テーブル用のディレクトリ下に置く、という強い制限がある
