# s3tablesinfo

S3 Tablesを構成するデータファイル(.parquet)を
S3オブジェクトとして取得するサンプルプログラム

## Usage

以下のようにS3 Tables バケットのARNを引数 `-arn` に指定して実行すると、
全名前空間の全テーブルについて、カレントスナップショットを構成する全データファイルの
S3オブジェクトのヘッドオブジェクトを取得する。

```console
$ go run main.go -arn arn:aws:s3tables:ap-northeast-1:000000000000:bucket/my-s3-tables-tutorial

namespaces=[[t1]]

[t1 daily_sales]
#0 FilePath=s3://bd13fd45-e8b0-4330-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx--table-s3/data/lI6hLQ/sale_date_month=2024-01/20260624_031951_00160_s5zfz-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy.parquet
    bucket=bd13fd45-e8b0-4330-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx--table-s3
    key=data/lI6hLQ/sale_date_month=2024-01/20260624_031951_00160_s5zfz-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy.parquet
&s3.HeadObjectOutput{
  AcceptRanges:         &"bytes",
  ContentLength:        &647,
  ContentType:          &"application/octet-stream",
  ETag:                 &"\"52214d2b2d75c8409552a46dc8122870\"",
  LastModified:         &2026-06-24 03:19:53 UTC,
  ServerSideEncryption: "AES256",
  VersionId:            &"jySvqMm0TOpdRUA59QKNHelB7.fIeAjE",
}
#1 FilePath=s3://bd13fd45-e8b0-4330-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx--table-s3/data/QaNS4Q/sale_date_month=2024-02/20260624_031951_00160_s5zfz-zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz.parquet
    bucket=bd13fd45-e8b0-4330-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx--table-s3
    key=data/QaNS4Q/sale_date_month=2024-02/20260624_031951_00160_s5zfz-zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz.parquet
&s3.HeadObjectOutput{
  AcceptRanges:         &"bytes",
  ContentLength:        &686,
  ContentType:          &"application/octet-stream",
  ETag:                 &"\"6f9e2f478b5891a4896e4a2888492652\"",
  LastModified:         &2026-06-24 03:19:53 UTC,
  ServerSideEncryption: "AES256",
  VersionId:            &"2jFl.w6okDWv43A7R7JfHzC83C_DfGYe",
}
```

## 結論

AWS S3 Tablesには Iceberg REST Catalog 互換のエンドポイントがある。

```
https://s3tables.<region>.amazonaws.com/iceberg
```

参考: <https://docs.aws.amazon.com/ja_jp/AmazonS3/latest/userguide/s3-tables-integrating-open-source.html>

このエンドポイントを介して、Icebergの実装(ライブラリ)がカタログ情報→メタデータ→マニフェストリスト→マニフェストファイルを取得でき、
データファイルに相当するS3オブジェクトのURLが確定する。

そのURLに対しては、内部的にS3の HeadObject や GetObject などの通常と同じAPI操作でアクセスできる。
ただし、IAM権限には s3tables:GetTableData を指定し、
リクエストのAWS署名（SigV4）対象サービス名として s3tables を正しく指定（設定）しておく必要がある。

なお [S3 TablesのAPI](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Operations_Amazon_S3_Tables.html) は、S3 Tablesのリソース管理(コントロールプレーン)のためのAPI。
SQLで言うところのDDL (Data Definition Language)
