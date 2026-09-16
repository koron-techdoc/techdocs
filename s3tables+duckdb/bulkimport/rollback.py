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

history = table.history()

if len(history) >= 2:
    previous_snapshot_id = history[-2].snapshot_id
    print(f"current snapshot id: {history[-1].snapshot_id}")
    print(f"rollback to snapshot id: {previous_snapshot_id}")
    with table.manage_snapshots() as ms:
        ms.rollback_to_snapshot(previous_snapshot_id)
        ms.commit()
    print("rollbacked successfully")
else:
    print("can't rollback, less history")

