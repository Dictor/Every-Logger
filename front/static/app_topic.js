var Model = {
    TopicId: "",
    TopicName: "",
    TopicDetail: "",
    
    Value: 0.0,
    ValueLastTerm: 0.0,
    ValueDate: 0,
    ValueDateDelta: 0.0,
    ValueDelta: 0.0,
    
    RecievedDateDelta: "",
    RecievedDate: 0,
    
    Term: ((this.Term = (new URL(window.location)).searchParams.get('term')) ? this.Term : "10m"),
    TermList: ["10s", "1m", "10m", "1h", "1d"],
    TermChange: async function(t) {
        await View.ChangeTerm(t);
    },
    HumanTerm: function(t) {
        let post = t[t.length - 1];
        return t.replace(post, ["초", "분", "시간", "일"]["smhd".indexOf(post)]);
    },
    History: {},
    
    ErrorMsg: "",
    Chart: null,
    NowTab: 0,
    
    moment: moment
};

var View = {
    Info: null,
    Tab: null,
    ws: null,
    Init: function(topic) {
        moment.locale('ko');
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
        Model.TopicId = topic;
        this.ws = new WS(this.wsOnOpen, this.wsOnMsg);
        setInterval(function() {
                Model.ValueDateDelta = View.Datef(Model.ValueDate);
                Model.RecievedDateDelta = View.Datef(Model.RecievedDate);
        }, 100);
    },
    ChangeTerm: async function(term) {
        Model.Term = term;
        if (Model.History[term] === undefined) {
            Model.History[term]  = await API.GetValueHistory(Model.TopicId, term);
        }
        this.DrawChart(Model.History[term]);
    },
    Datef: function(timestamp) {
         let diff_sec = (Date.now() - timestamp) / 1000;
         if (diff_sec < 60) {
             return diff_sec.toFixed(1) +"초 전";
         } else {
             return moment.unix(timestamp).fromNow();
         }
    },
    wsOnOpen: async function(evt) {
            View.ws.Send("TOPIC," + Model.TopicId);
            Model.History[Model.Term] = await API.GetValueHistory(Model.TopicId, Model.Term);
            Model.History["1d"] = await API.GetValueHistory(Model.TopicId, "1d");
            let m = Model.History[Model.Term];
            View.DrawChart(m);
            Model.ValueLastTerm = Number(Model.History["1d"][1][1]);
        },
    wsOnMsg: function (evt) {
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
            yAxis: {
                title: {
                    text: ''
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