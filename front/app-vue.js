var Model = {
    TopicName: "",
    TopicDetail: "",
    Value: 0.0,
    ValueLastTerm: 0.0,
    ValueDate: 0,
    ValueDelta: 0.0,
    RecievedDateDelta: 0.0,
    RecievedDate: 0,
    Chart: null
};

//
var View = {
    InfoContent: null,
    Init: function(topic) {
        this.InfoContent = new Vue({
            el: "#info-content",
            data: Model
        });
        
        Ws.Init();
        Ws.conn.onopen = function(evt) {
            Ws.Send("TOPIC,"+topic);
        }
        
        setInterval(function() {
                Model.RecievedDateDelta = (Date.now() - Model.RecievedDate) / 1000;
            }, 100);
    }
};

var Ws = {
    conn: null,
    Init: function() {
        this.conn = new WebSocket("ws://" + document.location.host + "/ws");
        this.conn.onclose = function (evt) {
            //UI.showError("Websocket closed, Please refresh this page.");
        };
        this.conn.onmessage = function (evt) {
            var pstr = evt.data.split(",");
            switch (pstr.length) {
                case 2:
                    switch (pstr[0]) {
                        case "ERROR":
                            //UI.showError(pstr[1]);
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
                            Model.ValueDelta = (Model.ValueLastTerm - Model.Value) / Model.ValueLastTerm * 100;
                            Model.ValueLastTerm = Model.Value;
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

