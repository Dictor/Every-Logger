var Model = {
    TopicDetail: null,
    Menu: ["전체 토픽", "급상승 토픽"],
    NowMenu: "전체 토픽",
    goTopic: function(u) {
        let w = window.open(global_host + "/topic?id=" + u, "");
        w.focus();
    }
}

var View = {
    vNavbar: null,
    vContent: null,
    Init: async function() {
        this.vNavbar = new Vue({
            el: "#menu",
            data: Model
        });
        this.vContent = new Vue({
            el: "#content",
            data: Model
        });
        Model.TopicDetail = await API.GetTopicDetail();
    }
}