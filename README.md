# DcardBackendIntern2024

## Api
### POST /api/v1/ad
```bash
curl -X POST -H "Content-Type: application/json" 127.0.0.1:8000/api/v1/ad --data '{
  "title": "AD 55",
  "startAt": "2024-04-07T00:05:00.000Z",
  "endAt": "2024-04-21T03:00:00.000Z",
  "conditions": [
    {
      "ageStart": 20,
      "ageEnd": 30,
      "platform": [
        "android",
        "ios"
      ],
      "gender": [
        "M"
      ]
    }
  ]
}'
```

### GET /api/v1/ad
```bash
curl "127.0.0.1:8000/api/v1/ad" | jq
curl "127.0.0.1:8000/api/v1/ad?offset=15&limit=65&age=35&gender=F&country=SO&platform=android" | jq
```

## Run
### Use postgres
```bash
docker compose up -d
```

### Use mariadb
```bash
docker compose up -f docker-compose-mysql.yml -d
```

## Build
```bash
make clean && make
```

## Unit test
```bash
make test
```

## Install
```bash
make install
```

## Uninstall
```bash
make uninstall
```

## Design
程式的部分使用 gorm 與資料庫溝通，資料庫建立兩個 table 分別為 ads 與 conditions。conditions 的部分再 insert 時會將所有屬性作 sha256 以後作為 ID 並且加上編號以防 Hash Collision，這樣當兩個 ad 有同個 condition 時可以減少 create 的次數。ads 會與 conditions 做關聯。

Create 的流程大致如下：
1. 計算所有 condition 的 Hash
2. 搜尋具有同 Hash（UUID） 值的 condition
3. 如果有同 Hash 值但不同內容，將 UUID 上編號
4. 如果該 condition 不存在 database 對 condition 做 create
5. Create AD
6. 清除 cache

Filter 的流程大致如下：
1. 將 ads 與 conditions Join
2. 檢查 filter 是否存在於 cache 中，有則值接回傳
3. 沒有在 cache 中的話，使用 filter 的內容對 database 做 select
4. 計算 cache 的 expiration 並使用 filter 作為 key 存入 cache server（redis）

Unit test 有
1. AD.Create 測試是否正常新增資料
2. Filter.SQL 測試是否正常對 database 做搜尋
3. Filter.Find 測試 cache 是否正常運作
4. Filter.Stress_Gen 與 Filter.Stress 壓力測試



