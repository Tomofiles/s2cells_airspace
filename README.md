S2 Geometryを用いた空間データの表示制御
====

Overview

## Description
このリポジトリーは、地図のズームイン・アウト時に表示する空間データの絞り込みを、S2 Geometryを用いて構築したサンプルです。

[S2 Geometry](https://s2geometry.io/)  
[S2 geometry library in Go](https://github.com/golang/geo)

地図上に表示する空間データは、`DID`と`飛行場周辺`のデータとなります。  
この2種類の空間データは、ドローンの飛行許可申請が必要な空域であり、このアプリケーションは、そのまま`ドローン用空域確認マップ`として、利用していただけます。

使用技術
- バックエンド：Go, echo, S2 Geometry
- フロントエンド：React, Leaflet (React-Leaflet)
- データベース：CockroachDB
- プラットフォーム：Docker, Docker Compose

## Demo
![](https://user-images.githubusercontent.com/27773127/84903803-83017200-b0e9-11ea-9f01-622d5601b47e.gif)

## Usage

## Install

### 1. 以下のコマンドを実行して、アプリケーションを起動します。
```sh
$ cd build_app/dev/
$ docker-compose -f docker-compose_s2cells-airspace.yaml -p s2cells-airspace_sandbox up
```
以下のURLにて、アクセスできます。

http://localhost:8080/

### 3. DIDの空間データを投入します。
以下の国土数値情報のダウンロードサイトにアクセスし、全国の「A16-15_GML.zip」をダウンロードします。  
https://nlftp.mlit.go.jp/ksj/gml/datalist/KsjTmplt-A16-v2_3.html

zipを解凍し、「A16-15_00_DID.geojson」があることを確認します。

curlコマンド等で、アプリケーションに投入します。  
```sh
$ curl -i -X POST http://localhost:8080/upload/did_areas -H "application/json" -d "@./A16-15_00_DID.geojson"
```

### 4. 飛行場周辺の空間データを投入します。
以下のサイトから、世界の飛行場データを取得します。（このサイトは、有志が航空データを収集して提供しているようです）  
https://ourairports.com/data/airports.csv

curlコマンド等で、アプリケーションに投入します。  
```sh
$ curl -i -X POST http://localhost:8080/upload/airport_areas -F "file=@./airports.csv"
```

## Licence
MIT License

## Author
[Tomofiles](https://github.com/Tomofiles)