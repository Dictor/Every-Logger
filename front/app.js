var UI = {
    latestValue: 0.0,
    colorApplyDomId: ["info-value", "info-delta"],
    name: "",
    detail: "",
    chart: null,
    updateValue: function(val) {
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
        this.chart.series[0].addPoint(val);
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
    init: function(topic) {
        this.initChart();
        ws.init();
        ws.conn.onopen = function(evt) {
            ws.send("TOPIC,"+topic);
        }
    },
    initChart: function() {
        this.chart = Highcharts.chart('container', {
            chart: {
                type: 'line',
                backgroundColor: '#1C1D21'
            },
            title: {text: this.name},
            subtitle: {text: this.detail},
            xAxis: {
                categories: this.timeHistory
            },
            plotOptions: {
                line: {
                    dataLabels: {
                        enabled: true
                    },
                    enableMouseTracking: false
                }
            },
            series: [{
                name: 'Data',
                data: this.dataHistory
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
            var pstr = evt.data.split(",");
            switch (pstr.length) {
                case 2:
                    switch (pstr[0]) {
                        case "ERROR":
                            alert(pstr[1]);
                            break;
                    }
                case 3:
                    switch (pstr[0]) {
                        case "VALUE":
                            UI.updateValue(Number(Number(pstr[2]).toFixed(2)));
                            break;
                        case "TOPIC":
                            UI.updateInfo(pstr[1], pstr[2]);
                            break;
                    }
            }
        };
    },
    send: function(val) {
        this.conn.send(val);
    }
}