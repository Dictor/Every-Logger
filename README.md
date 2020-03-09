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
 

