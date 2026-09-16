from pyiceberg.catalog import load_catalog

catalog = load_catalog(
        "r2_catalog",
        **{
            "type": "rest",
            "uri": "https://catalog.cloudflarestorage.com/{ACCOUNT ID}/{CATALOG NAME}",
            "warehouse": "{WAREHOUSE NAME}",
            "token": "{R2 API TOKEN}",
        }
)

table = catalog.load_table('default.daily_sales')

table.add_files(file_paths=["s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE ID}/data/daily_sales-202403.parquet"])

#OK: table.add_files(file_paths=["s3://eval/__r2_data_catalog/{NAMESPACE ID}/{TABLE ID}/daily_sales-202403.parquet"])
#NG: table.add_files(file_paths=["s3://eval/__r2_data_catalog/{NAMESPACE ID}/daily_sales-202403.parquet"])
#NG: table.add_files(file_paths=["s3://eval/__r2_data_catalog/daily_sales-202403.parquet"])
