var UI = {
    latestValue: 0.0,
    latestRecieveTime: 0,
    latestTermValue: 0.0,
    colorApplyDomId: ["info-value", "info-delta"],
    name: "",
    detail: "",
    chart: null,
    updateValue: function(time, val) {
        document.getElementById("info-value").innerHTML = val;
        var delta = val - this.latestTermValue;
    
        document.getElementById("info-delta").innerHTML = "";
        for (var id of this.colorApplyDomId) {
            var obj = document.getElementById(id);
            if (delta < 0) {
                obj.classList.remove("color-increase");
                obj.classList.add("color-decrease");
            } else if (delta == 0) {
                obj.classList.remove("color-increase");
                obj.classList.remove("color-decrease");
            } else {
                obj.classList.add("color-increase");
                obj.classList.remove("color-decrease");
            }
        }
        document.getElementById("info-value").classList.remove("color-changed");
        if (this.latestTermValue != 0.0) document.getElementById("info-delta").innerHTML += (delta > 0 ? "+" : "-") + (Math.abs(delta / this.latestTermValue) * 100).toFixed(2) + "%";
        this.latestValue = val;
        if (this.latestRecieveTime == 0) {
            setInterval(function() {
                document.getElementById("info-date").innerHTML = ((Date.now() - UI.latestRecieveTime) / 1000).toFixed(1) + "초 전 수신됨";
            }, 100)
        }
        this.latestRecieveTime = Date.now()
        this.chart.series[0].addPoint([Number(time) * 1000, val]);
    },
    updateInfo: function(name, detail) {
        document.getElementById("info-name").innerHTML = name;
        document.getElementById("info-detail").innerHTML = detail;
        this.name = name;
        this.detail = detail;
    },
    showError: function(msg) {
        document.getElementById("info-error").classList.remove("hidden");
        document.getElementById("info-content").classList.add("hidden");
        document.getElementById("info-error-msg").innerHTML = msg;
        document.getElementById("info").classList.add("info-error");
    },
    init: async function(topic) {
        if (!topic || topic == "") {
            UI.showError("Invalid topic name");
            return;
        }
        ws.init();
        ws.conn.onopen = async function(evt) {
            ws.send("TOPIC,"+topic);
            var ival = await RequestXhrGetPromise("ival?topic=" + topic + "&term=1m");
            ival = JSON.parse(ival);
            UI.latestTermValue = Number(ival[ival.length - 1][1]);
            UI.initChart(ival);
        }
    },
    initChart: function(ivalue) {
        Highcharts.setOptions({
            time: {
                timezone: 'Asia/Seoul'
            }
        });
        this.chart = Highcharts.chart('container', {
            chart: {
                type: 'spline',
                backgroundColor: '#1C1D21'
            },
            title: {
                text: this.name,
                style: {color: '#D9D5C1'}
            },
            xAxis: {
                type: 'datetime',
                title: {
                    text: '시간'
                }
            },
            plotOptions: {
                line: {
                    dataLabels: {
                        enabled: true
                    },
                    enableMouseTracking: false
                },
                series: {
                    color: '#2ECC71'
                }
            },
            tooltip: {
                headerFormat: '<b>' + this.name + '</b><br>',
                pointFormat: '{point.x:%Y/%m/%d %H:%M:%S} : <b>{point.y:.2f}</b>'
            },
            series: [{
                name: this.name,
                data: ivalue
            }]
        });
    }
};

var ws = {
    conn: null,
    init: function() {
        this.conn = new WebSocket("ws://" + document.location.host + "/ws");
        this.conn.onclose = function (evt) {
            UI.showError("Websocket closed, Please refresh this page.");
        };
        this.conn.onmessage = function (evt) {
            //console.log(evt.data)
            var pstr = evt.data.split(",");
            switch (pstr.length) {
                case 2:
                    switch (pstr[0]) {
                        case "ERROR":
                            UI.showError(pstr[1]);
                            break;
                    }
                case 3:
                    switch (pstr[0]) {
                        case "TOPIC":
                            UI.updateInfo(pstr[1], pstr[2]);
                            break;
                    }
                case 4:
                    switch (pstr[0]) {
                        case "VALUE":
                            UI.updateValue(Number(pstr[2]), Number(Number(pstr[3]).toFixed(2)));
                            break;
                    }
            }
        };
    },
    send: function(val) {
        this.conn.send(val);
    }
}