var Model = {
    TopicName: "",
    TopicDetail: "",
    Value: 0.0,
    ValueLastTerm: 0.0,
    ValueDate: 0,
    ValueDelta: 0.0,
    RecievedDateDelta: 0.0,
    RecievedDate: 0,
    Term: ((this.Term = (new URL(window.location)).searchParams.get('term')) ? this.Term : "10m"),
    ErrorMsg: "",
    Chart: null,
    NowTab: 0,
    History: {},
    moment: moment
};

var View = {
    Info: null,
    Tab: null,
    Init: function(topic) {
        this.Info = new Vue({
            el: "#info",
            data: Model
        });
        this.Tab = new Vue({
            el: "#tab",
            data: Model
        });
        if (!topic || topic == "") {
            Model.ErrorMsg = "Invalid topic name.";
            return;
        }
        Ws.Init();
        Ws.conn.onopen = async function(evt) {
            Ws.Send("TOPIC,"+topic);
            Model.History[Model.Term] = await API.GetValueHistory(topic, Model.Term);
            let m = Model.History[Model.Term];
            View.DrawChart(m);
            Model.ValueLastTerm = Number(m[m.length - 1][1]);
        }
        setInterval(function() {
                Model.RecievedDateDelta = (Date.now() - Model.RecievedDate) / 1000;
        }, 100);
    },
    DrawChart: function(ivalue) {
        Highcharts.setOptions({
            time: {
                timezone: 'Asia/Seoul'
            }
        });
        Model.Chart = Highcharts.chart('container', {
            chart: {
                type: 'spline',
                backgroundColor: '#1C1D21'
            },
            title: {
                text: Model.TopicName,
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
                headerFormat: '<b>' + Model.TopicName + '</b><br>',
                pointFormat: '{point.x:%Y/%m/%d %H:%M:%S} : <b>{point.y:.2f}</b>'
            },
            series: [{
                name: Model.TopicName,
                data: ivalue
            }]
        });
    }
};

var Ws = {
    conn: null,
    Init: function() {
        this.conn = new WebSocket("ws://" + document.location.host + "/ws");
        this.conn.onclose = function (evt) {
            Model.ErrorMsg = "Websocket closed, Please refresh this page.";
        };
        this.conn.onmessage = function (evt) {
            var pstr = evt.data.split(",");
            switch (pstr.length) {
                case 2:
                    switch (pstr[0]) {
                        case "ERROR":
                            Model.ErrorMsg = pstr[1];
                            break;
                    }
                case 3:
                    switch (pstr[0]) {
                        case "TOPIC":
                            Model.TopicName = pstr[1];
                            Model.TopicDetail = pstr[2];
                            break;
                    }
                case 4:
                    switch (pstr[0]) {
                        case "VALUE":
                            Model.Value = Number(pstr[3]);
                            Model.ValueDate = Number(pstr[2]);
                            Model.ValueDelta = (Model.Value - Model.ValueLastTerm) / Model.ValueLastTerm * 100;
                            Model.RecievedDate = Date.now();
                            break;
                    }
            }
        };
    },
    Send: function(val) {
        this.conn.send(val);
    }
}

