資料:

- <https://docs.aws.amazon.com/ja_jp/AmazonS3/latest/userguide/s3-tables-getting-started.html>
- <https://duckdb.org/docs/current/core_extensions/iceberg/amazon_s3_tables>


Amazon Athena コンソールで以下のクエリを実行

1.  テーブル作成
2.  データ投入
3.  クエリー実行

```
--Use the following statement to create a table in your S3 Table bucket.
CREATE TABLE `t1`.daily_sales (
sale_date date, 
product_category string, 
sales_amount double)
PARTITIONED BY (month(sale_date))
TBLPROPERTIES ('table_type' = 'iceberg');

-- Next steps 1) Use the following SQL statement to insert data to your table.
INSERT INTO daily_sales
VALUES
(DATE '2024-01-15', 'Laptop', 900.00),
(DATE '2024-01-15', 'Monitor', 250.00),
(DATE '2024-01-16', 'Laptop', 1350.00),
(DATE '2024-02-01', 'Monitor', 300.00),
(DATE '2024-02-01', 'Keyboard', 60.00),
(DATE '2024-02-02', 'Mouse', 25.00),
(DATE '2024-02-02', 'Laptop', 1050.00),
(DATE '2024-02-03', 'Laptop', 1200.00),
(DATE '2024-02-03', 'Monitor', 375.00);

-- 2) Use the following SQL statement to run a sample analytics query.
SELECT 
product_category,
COUNT(*) as units_sold,
SUM(sales_amount) as total_revenue,
AVG(sales_amount) as average_price
FROM daily_sales
WHERE sale_date BETWEEN DATE '2024-02-01' and DATE '2024-02-29'
GROUP BY product_category
ORDER BY total_revenue DESC;
```

DuckDB (1.5.4) で以下を実行

1.  AWS credential chain (~/.config/aws/credentials) を利用する設定
2.  s3tables に ARN を指定して接続
3.  接続したテーブルを表示
4.  テーブルに対してクエリを実行(Athenaで実行したものと同じ

```
create secret ( type s3, provider credential_chain ); 

attach 'arn:aws:s3tables:ap-northeast-1:961554615704:bucket/kaoriya-s3tables-tutorial' as s3table ( type iceberg, endpoint_type s3_tables);

show all tables;

SELECT
         product_category,
         COUNT(*) as units_sold,
         SUM(sales_amount) as total_revenue,
         AVG(sales_amount) as average_price
         FROM s3table.t1.daily_sales
         WHERE sale_date BETWEEN DATE '2024-02-01' and DATE '2024-02-29'
         GROUP BY product_category
         ORDER BY total_revenue DESC;
```

普通に DuckDB でやるので良いのでは?
