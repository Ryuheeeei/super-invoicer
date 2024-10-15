# super-invoicer

See https://github.com/upsidr/coding-test/blob/main/web-api-language-agnostic/README.ja.md

## Help

```console
$ go run main.go -h
App for creating and getting invoices.

Usage:
   [flags]

Flags:
      --basic-auth.enable            Enable basic authentication or not
      --basic-auth.password string   Password for basic authentication
      --basic-auth.username string   Username for basic authentication
  -h, --help                         help for this command
```

## Environment Variables

| Name                  | Description                                                                       |
| --------------------- | --------------------------------------------------------------------------------- |
| `MYSQL_USERNAME`      | Username for mysql cluster                                                        |
| `MYSQL_PASSWORD`      | Password for mysql cluster                                                        |
| `MYSQL_ROOT_PASSWORD` | Root password for mysql cluster (NOTE: necessary only when run in docker compose) |

## How to run in docker compose environment

```console
$ MYSQL_USERNAME=root
$ MYSQL_PASSWORD=mysql
$ MYSQL_ROOT_PASSWORD=mysql
$ make up
```

MySQL の起動時に`data/init.sql`の内容が実行されます。

## API Spec. and Behavior Check

### `GET /api/invoices`

現在の日付から`due_date`の日付までに支払い必要のある`company_id`の請求書一覧を返却します。

```txt
HTTP Method: GET
Query:
- company_id: string
- due_date: YYYY-MM-DD
```

レスポンス例

200 ok

```console
$ curl -i -u "foo:bar" "localhost:8080/api/invoices?company_id=1&due_date=2026-02-02"
HTTP/1.1 200 OK
Date: Tue, 15 Oct 2024 22:21:29 GMT
Content-Length: 428
Content-Type: text/plain; charset=utf-8

{"invoices":[{"invoice_id":"1","company_id":"1","issue_date":"2024-11-01T00:00:00Z","amount":10000,"fee":400,"fee_rate":0.04,"tax":40,"tax_rate":0.1,"total":10440,"due_date":"2024-12-01T00:00:00Z","status":"unprocessed"},{"invoice_id":"2","company_id":"1","issue_date":"2024-10-01T00:00:00Z","amount":5000,"fee":200,"fee_rate":0.04,"tax":20,"tax_rate":0.1,"total":5220,"due_date":"2024-11-01T00:00:00Z","status":"processing"}]}
```

400 bad request

-   Query Parameter が両方指定されていない
-   due_date が日付(YYYY-MM-DD)として不適切

```console
$ curl -i -u "foo:bar" "localhost:8080/api/invoices?company_id=&due_date=2026-02-02"
HTTP/1.1 400 Bad Request
Date: Tue, 15 Oct 2024 22:25:27 GMT
Content-Length: 43
Content-Type: text/plain; charset=utf-8

{"message":"'company_id' mustn't be empty"}

$ curl -i -u "foo:bar" "localhost:8080/api/invoices?company_id=1&due_date="
HTTP/1.1 400 Bad Request
Date: Tue, 15 Oct 2024 22:26:59 GMT
Content-Length: 53
Content-Type: text/plain; charset=utf-8

{"message":"Can't convert duedate parameter to date"}

$ curl -i -u "foo:bar" "localhost:8080/api/invoices?company_id=1&due_date=INVALID"
HTTP/1.1 400 Bad Request
Date: Tue, 15 Oct 2024 22:27:37 GMT
Content-Length: 53
Content-Type: text/plain; charset=utf-8

{"message":"Can't convert duedate parameter to date"}
```

401 Unauthorized

-   Basic 認証有効時にヘッダーが指定されていない
-   Basic 認証有効時に誤った認証情報を送信している

NOTE: `compose.yaml`を以下のように修正すると、Basic 認証なしのアプリケーションで起動ができます

<details><summary> compose.yaml </summary>

```yaml
services:
    super-invoicer:
        build: .
        environment:
            MYSQL_USERNAME: ${MYSQL_USERNAME}
            MYSQL_PASSWORD: ${MYSQL_PASSWORD}
        ports:
            - "8080:8080"
    db:
        image: mysql:8.4.2
        environment:
            MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
        volumes:
            - ./data/:/docker-entrypoint-initdb.d
        ports:
            - "3306:3306"
```

</details>

```console
$ curl -i "localhost:8080/api/invoices?company_id=1&due_date=2026-02-02"
HTTP/1.1 401 Unauthorized
Date: Tue, 15 Oct 2024 22:24:48 GMT
Content-Length: 48
Content-Type: text/plain; charset=utf-8

{"message":"Authorization Header doesn't exist"}

$ curl -i -u "foo:INCORRECT" "localhost:8080/api/invoices?company_id=1&due_date=2026-02-02"
HTTP/1.1 401 Unauthorized
Date: Tue, 15 Oct 2024 22:30:58 GMT
Content-Length: 26
Content-Type: text/plain; charset=utf-8

{"message":"Unauthorized"}
```

500 Internal Server Error

-   DB との接続に失敗した場合など

### `POST /api/invoices`

リクエストボディに記載された内容で請求書データを作成します。
レスポンスボディとして新規作成された請求書データを返却します。

```txt
HTTP Method: POST
Request Body:
- company_id: string
- amount: int
- issue_date: YYYY-MM-DD
- due_date: YYYY-MM-DD
- status: ["unprocessed", "processing", "paid", "error"]
```

200 ok

```console
# curlの場合Basic認証は以下のように書くことも可能です
$ curl -XPOST -d '{"company_id": "1", "amount": 10000, "issue_date": "2020-01-01", "due_date": "2026-01-21", "status": "paid"}' -H "Authorization:Basic $(echo -n foo:bar | openssl base64)" "localhost:8080/api/invoices"
{"invoice_id":"5","company_id":"1","issue_date":"2020-01-01T00:00:00Z","amount":10000,"fee":400,"fee_rate":0.04,"tax":40,"tax_rate":0.1,"total":10440,"due_date":"2026-01-21T00:00:00Z","status":"paid"}
```

<details><summary>実行後のテーブル</summary>

invoice_id=5 のレコードが新規に追加されている

```sql
> SELECT * FROM invoice;
+------------+------------+------------+--------+-----+----------+-----+----------+-------+------------+-------------+
| invoice_id | company_id | issue_date | amount | fee | fee_rate | tax | tax_rate | total | due_date   | status      |
+------------+------------+------------+--------+-----+----------+-----+----------+-------+------------+-------------+
|          1 |          1 | 2024-11-01 |  10000 | 400 |     0.04 |  40 |     0.10 | 10440 | 2024-12-01 | unprocessed |
|          2 |          1 | 2024-10-01 |   5000 | 200 |     0.04 |  20 |     0.10 |  5220 | 2024-11-01 | processing  |
|          3 |          1 | 2024-07-01 |  20000 | 800 |     0.04 |  80 |     0.10 | 20880 | 2024-08-01 | paid        |
|          4 |          2 | 2024-04-01 |   5000 | 200 |     0.04 |  20 |     0.10 |  5220 | 2024-11-01 | error       |
|          5 |          1 | 2020-01-01 |  10000 | 400 |     0.04 |  40 |     0.10 | 10440 | 2026-01-21 | paid        |
+------------+------------+------------+--------+-----+----------+-----+----------+-------+------------+-------------+
5 rows in set (0.01 sec)
```

</details>

400 Bad Reqeust

-   company_id が指定されていない場合
-   issue_date, due_date が日付として不適切な場合
-   status が [unprocessed, processing, paid, error] のいずれでもない

```console
$ curl -i -XPOST -d '{"company_id": "1", "amount": 10000, "issue_date": "2020-01-01", "due_date": "2026-01-21", "status": "UNKNOWN"}' -H "Authorization:Basic $(echo -n foo:bar | openssl base64)" "localhost:8080/api/invoices"
HTTP/1.1 400 Bad Request
Date: Tue, 15 Oct 2024 22:38:09 GMT
Content-Length: 93
Content-Type: text/plain; charset=utf-8

{"message":"'status' must be one of [unprocessed, processing, paid, error], but got UNKNOWN"}
```

401 Unauthorized

-   `GET /api/invoices`の時と同様

500 Internal Server Error

-   DB との接続に失敗した場合など
