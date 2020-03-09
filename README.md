# Every-Logger

_This project is currently **discontinued**. Without there is nothing special, developing doesn't be continued. 
Some problems like [#2](https://github.com/Dictor/Every-Logger/issues/2), [#3](https://github.com/Dictor/Every-Logger/issues/3) aren't solved._

A web service which log every data across web. Crawl (=fetch) values of topic and store and display it to user.

## Structure
- **front/** : Static front-end implemented with vuejs. Live values are fetched through websocket.
- **db.go** : Thin model layer of [badger kv-store](https://github.com/dgraph-io/badger). Store and load fetched values.
- **fetch.go** : Fetching functions which supporting ways like from _Raw html on web, Json on web, Local file, Random created value_. 
Support selecting specific html element with [goquery](https://github.com/PuerkitoBio/goquery).
- **fetch_chrome.go** : Fetching function with [chromedp chrome driver](https://github.com/chromedp/chromedp).
- **fetch_util.go** : Utility functions using in fetching functions.
- **main.go** : Initialing several routine including [echo web framework](https://echo.labstack.com).
- **route.go** : Handler functions used in echo framework.
- **topic.go** : Topic managing functions and Used topics are defined.
- **websocket.go** : Thin websocket wrapper of [wswrapper](https://github.com/Dictor/wswrapper).
 
## Config
- In _./config.json_
```json
{
    "ws_origin": ["<allowed websocket origins>"]
}
```
- In _./db/topic_detail.json_

Current defined topic.go using below detail config.
```json
{
    "test": {"name": "테스트 데이터", "detail": "랜덤으로 생성되는 테스트용 데이터입니다."},
    "btcusd": {"name": "비트코인-미국달러", "detail": "1비트코인 당 미국 달러 환율입니다."},
    "co19-cn-cur": {"name": "코로나19 중국 현재 감염자", "detail": "중국의 현재 감염자 수입니다. (완치 및 사망자 제외)"},
    "co19-kr-all": {"name": "코로나19 한국 누적 감염자", "detail": "한국의 누적 감염자 수입니다."},
    "test-file": {"name": "테스트 파일 데이터", "detail": "파일 데이터 가져오기 테스트용 데이터입니다."}
}
```
