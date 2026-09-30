# コストを抑制したログ管理と、マルチクラウド構成の検証

## TL;DR

S3上の既存parquetをIceberg REST Catalogに最小コストで統合することは、現時点では現実的な選択肢とは言えない。
通常のクエリエンジンによるコピーを伴う挿入か、Glue Data Catalogを使うか、Iceberg v4 のマルチクラウド構成を待つべき。

## 出発点: S3上のparquetのIceberg REST Catalog化

もともとの発想は、S3に記録したログ≒parquetを、最小のコストでIceberg REST Catalogに統合することだった。これは理論上はメタデータとマニフェストだけを更新すれば達成できるとの考えに基づく。

## PyIcebergによる検証

この考えは正しくて、具体的には PyIceberg (`add_files`関数) でできることがわかった。
この時 PyIceberg で起こっているのは、
追加するデータ(parquet)のメタデータの取得、
テーブル毎のクレデンシャルの取得、
マニフェストやメタデータの修正・作成と書き込み、だった。

テーブル毎のクレデンシャルの取得は読み込み操作においても行われ、
parquetの取得にも使われる。
これは R2 Data Catalogにおいて、
parquetを置く場所がテーブル用のディレクトリに制限されている原因でもあった。

## DuckDBの振舞い

ここまでの過程において、 R2に置いたデータファイルをAWS S3にあるものだと、DuckDBが誤解して取得を試みるケースに遭遇した。
これはDuckDBの機能としてクレデンシャルのフォールバックが起こることにより発生していた。
それを見て我々はR2 Data CatalogとAWS S3を混在させる、
一種のマルチクラウド戦略が可能であると考えた。

## マルチクラウド構成の実現性

しかし検証の結果、これは実用的ではないとの結論に至った。

その主たる理由はデータ追加時にある。
PyIcebergはマニフェスト・メタデータの計算のために、
対象のparquetファイルの一部を取得する必要がある。
それはR2のカタログ側でなく、手元のクライアント側がR2バケットとAWS S3の双方のURLを正しく識別し、
それぞれの認証情報を動的に切り替えてフッターを取得する必要がある、ということである
素の状態ではR2ストレージのURLとして解釈・取得していることから、
S3として解釈するように設定・改造できるかもしれないが、
開発・運用・メンテナンスのコストから考え現実的な手段ではない。

同じことはDuckDB側においても言える。
S3などのストレージにおけるクレデンシャルの自動選択というUX目的の機能と、
Icebergのより厳密なクレデンシャル運用とがミスマッチしている。
その仕様を転用してマルチクラウドを実現するのは、できたとしてもコスト的に現実的とは言えない。

S3とR2の二重アップロードをするなど、実現手段は考えられる。
しかしそれは当初の「最小のコストで…」に反する。

## Iceberg の認証機構/ストレージアクセス委譲

一方で AWS S3 Tables においては、テーブル毎のクレデンシャルではなく、
リモート署名を起点にした認証を行っている。
これは鍵を直接やり取りしないという点においてセキュアであり歓迎されている。
またIcebergにおいては比較的新しい認証方法である。

Iceberg自体は元々がNetflixの内製システムであったため、
比較的セキュアなネットワークを前提にした認証≒クレデンシャルを用いていた。
そのため近年のゼロトラスト哲学と相容れず、リモート署名が導入された。

R2 Data CatalogとAWS S3を混在運用するということは、この2つの認証方法を混在させるということでもある。
これは仮に可能であったとしても、思わぬセキュリティホールを生じさせる可能性がある。
そのメリットに対してリスクとコストが大きすぎると、容易に判断できる。

参考: <https://iceberg.apache.org/docs/nightly/rest-protocol/#storage-access-delegation>

## 次期 Iceberg v4 におけるマルチクラウド対応

またIcebergの次期版v4において、まさに当初の課題であったマルチクラウド構成の導入が検討されている。
であるならばなおのこと、
現時点でDuckDBの仕様を転用してIcebergの仕様にはないマルチクラウドを実現することの優先度は下がる。

参考: <https://www.linkedin.com/pulse/apache-iceberg-v4-what-means-your-ai-data-stack-andrew-madson-22kac/>

## 補足: Iceberg REST Catalog の自前ホスト

Iceberg REST Catalogを自前でホストするコストが許容可能なら、S3上の既存データを比較的低コストで統合できる。
その際の選択肢としては以下のプロダクトが考えられる。

- [Apache Polaris][polaris] - Java
- [Lakekeeper][lakekeeper] - Rust
- [Project Nessie][nessie] - Java (ブランチやマージのようなことがきる)

[polaris]:https://polaris.apache.org/
[lakekeeper]:https://github.com/lakekeeper/lakekeeper
[nessie]:https://projectnessie.org/

## 補足: AWS Glue Data Catalog

Glue Data Catalog を使えば、既存のS3上のparquetにIceberg互換のAPIを提供できる。
もちろん別途メンテナンスに伴うコスト(DPU料金を含む)が生じる。
バッチ的なデータ追加が主であればGlueが向いている。
