var global_host = window.location.protocol + "//" + window.location.hostname;
function RequestXhrGetPromise(verb) {
    return new Promise(function(resolve, reject) {
        var req = new XMLHttpRequest();
        req.open("GET", global_host + "/" + verb, true);
        req.withCredentials = true;
        req.onload = function() {
            if (req.status == 200) {
                resolve(req.response);
            } else {
                resolve(null);
            }
        };
        req.send(); 
    })
}

var UI = {
    latestValue: 0.0,
    colorApplyDomId: ["info-value", "info-delta"],
    name: "",
    detail: "",
    ivalue: {},
    chart: null,
    updateIvalue: function(time, val) {
        this.ivalue[time] = val;
    },
    updateValue: function(time, val) {
        document.getElementById("info-value").innerHTML = val;
        var delta = val - this.latestValue;
    
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
        document.getElementById("info-delta").innerHTML += (delta > 0 ? "+" : "-") + (Math.abs(delta / this.latestValue) * 100).toFixed(2) + "%";
        this.latestValue = val;
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
        ws.init();
        ws.conn.onopen = function(evt) {
            ws.send("TOPIC,"+topic);
        }
        var ival = await RequestXhrGetPromise("ival?topic=" + topic + "&term=1m");
        UI.initChart(JSON.parse(ival));
    },
    initChart: function(ivalue) {
        this.chart = Highcharts.chart('container', {
            chart: {
                type: 'spline',
                backgroundColor: '#1C1D21'
            },
            title: {text: this.name},
            xAxis: {
                type: 'datetime',
                title: {
                    text: 'Date'
                }
            },
            plotOptions: {
                line: {
                    dataLabels: {
                        enabled: true
                    },
                    enableMouseTracking: false
                }
            },
            tooltip: {
                headerFormat: '<b>' + this.name + '</b><br>',
                pointFormat: '{point.x:%Y/%B/%e %H:%M:%S}: {point.y:.2f}'
            },
            series: [{
                name: 'Data',
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
                            UI.updateValue(Number(pstr[2]), Number(Number(pstr[3]).toFixed(3)));
                            break;
                    }
            }
        };
    },
    send: function(val) {
        this.conn.send(val);
    }
}