package main

import (
	ws "github.com/dictor/wswrapper"
	"github.com/kennygrant/sanitize"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func SetRouting(e *echo.Echo, h *ws.WebsocketHub) {
	/* Websocket*/
	e.GET("/ws", func(c echo.Context) error {
		h.AddClient(c.Response().Writer, c.Request())
		return nil
	})

	/* API */
	e.GET("/api/history/:topic/:term", rHistory)
	e.GET("/api/topic", rTopic)

	/* Frontend serving */
	e.Static("/static", "front/static")
	e.GET("/*", rFront)
}

func rTopic(c echo.Context) error {
	log.Println(makeEchoPrefix(c, "rTopic"))
	return c.JSON(http.StatusOK, topicDetail)
}

func rFront(c echo.Context) error {
	tparent, err := ioutil.ReadFile("front/index.html")
	if err != nil {
		log.Printf("%s %s", makeEchoPrefix(c, "rFront"), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	tchild, err := ioutil.ReadFile("front/view/" + sanitize.BaseName(c.Param("*")) + ".html")
	if err != nil {
		log.Printf("%s %s", makeEchoPrefix(c, "rFront"), err)
		return c.NoContent(http.StatusNotFound)
	}

	log.Printf("%s %d", makeEchoPrefix(c, "rFront"), http.StatusOK)
	return c.HTML(http.StatusOK, strings.Replace(string(tparent), "{view}", string(tchild), -1))
}

func rHistory(c echo.Context) error {
	log.Println(makeEchoPrefix(c, "rHistory"))
	topic_name := c.Param("topic")
	if _, topic_vaild := topicDetail[topic_name]; !topic_vaild {
		return c.JSON(http.StatusOK, map[string]interface{}{"result": false, "msg": "Invalid topic name"})
	}

	term := c.Param("term")
	if !isValidTerm(term) {
		return c.JSON(http.StatusOK, map[string]interface{}{"result": false, "msg": "Invalid term"})
	}

	smax_count := c.QueryParam("max")
	max_count := 100
	if val, err := strconv.Atoi(smax_count); err == nil {
		max_count = val
	}

	result := []interface{}{}
	if res, err := GetValue(topic_name, term, max_count); err == nil {
		for _, val := range res {
			now_value := []interface{}{val.Time * 1000, val.Value} //js uses timestamp with millisec while golang uses sec
			result = append(result, now_value)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"result": true, "value": result})
	} else {
		return c.JSON(http.StatusOK, map[string]interface{}{"result": false, "msg": "Internal server error"})
	}
}

func isValidTerm(term string) bool {
	if len(term) < 2 {
		return false
	}
	if !strings.Contains("dhms", string(term[len(term)-1])) {
		return false
	}
	if _, err := strconv.Atoi(term[0 : len(term)-1]); err != nil {
		return false
	}
	return true
}
