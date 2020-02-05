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

class API {
    static async GetValueHistory(topic_name, data_term) {
        var data = await RequestXhrGetPromise("ival?topic=" + topic_name + "&term=" + data_term);
        return JSON.parse(data);
    }
}