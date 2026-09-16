# IcebergへのBulkインポートの考察

ローカルにある複数のParquetファイルを、
Apache Icebergのテーブルにまとめてインポートする方法を検討する。

考えられる方法は以下の2つ

-   クエリエンジンを使って一括挿入
-   メタデータを直接書き換える

## 検討

### クエリエンジンを使って一括挿入

コード例:

```sql
INSERT INTO iceberg_catalog.db.my_table 
  SELECT * FROM parquet.`s3://my-bucket/path/to/files/`;
```

DuckDBやAWS Glueで通常のRDBMSのように挿入ができる。

おおよその目安として、
1回にインポートするファイル総量が1TBに届きそうまたは越えそうならばGlueを、
それ以下ならばDuckDBを選択するのが、コスト的に妥当なライン。

AWSを使用する場合、DuckDBはEC2, ECS, もしくはLambdaで動かし、
S3にParquetファイルが既に置いてある前程(その他細則有り)であれば、
ネットワーク転送量はかからない。
Glueも同様だが、
ストレージAPIリクエストはGlueのほうが少し多くなる可能性がある。

### メタデータを直接書き換える

PyIcebergを使えば、ストレージにアップロード済みのParquetを追加するように、
Icebergカタログのメタデータを直接書き換えることができる。

R2にアップロード済みのParquetファイルを、
R2 Data Caltalogのメタデータに追加するコード例:

```python
from pyiceberg.catalog import load_catalog

# Load the catalog
catalog = load_catalog(
    "r2_catalog",
    **{
        "type": "rest",
        "uri": "https://catalog.cloudflarestorage.com/{アカウントID}/{カタログ名}",
        "warehouse": "{ウェアハウス名}"
        "token": "{R2 APIトークン}"
    }
)

# Load the table to add files
table = catalog.load_table("my_database.my_table")

# Add files to the table
table.add_files(file_paths=["r2://my-bucket/path/to/new_file.parquet"])
```

この時、追加するファイルのスキーマやパーティション設定は、
対象としているテーブルと完全に一致している必要がある。
スキーマが違っていた場合はエラーとなり追加できない一方で、
カタログとは異なるパーティションのファイルを登録すると、
データに参照できないエラーになるなどの、リスクがある。

また、この方法はR2では利用できるが、S3では利用できない。
S3 Tablesはデータファイルのストレージが通常のS3ではなく、
S3 Tablesの専用のストレージ(S3と同等のAPIで読める)に保存されるため。
この専用ストレージへ直接ファイルを配置する方法が無い。

Apache SparkにもPyIceberg同様に、
メタデータにファイルを追加するための
[`add_files` プロシージャ][addfiles] がある。

[addfiles]:https://iceberg.apache.org/docs/1.5.1/spark-procedures/#add_files

これらの方法は、一括インポートとしてはコストが最小になるが、
一方でカタログのメンテナンスコスト
(自前システムのバグ対応や、Iceberg自身の仕様変更への追従)
が別途かかることに留意が必要。

## 実験

[./bulkimport](./bulkimport/README.md) 参照

## まとめ

-   一括インポートは基本、クエリエンジン(DuckDBなど)を使う
    -   データ総量が1TBを越えるかそれに近い時は AWS Glue を使う
-   メタデータの直接書き換えには PyIceberg を使う
    -   S3 Tables ではできない
    -   実際にやってみたら…できた!
