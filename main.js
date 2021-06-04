
async function GoEvent(id, sender, para) {
    var durl = "/api/goevent/";

    if (id) {
        durl += id.toString() + "/"
    } else {
        durl += "noid/"
    }

    if (sender) {
        durl += sender.toString() + "/"
    } else {
        durl += "nosender/"
    }

    if (para) {
        durl += encodeURI(para.toString())
    } else {
        durl += "null"
    }

    var resp = await fetch(durl)
        .then(response => response.json());

    console.log("Event raised : " + durl + "  :  Response " + resp);

    if (resp.Reload == true) {
        console.log("Reload page")
        location.reload();
    } else if (resp.Update == true) {
        console.log("Update " + resp.ID)
        InjectElement(resp.ID, resp.Content)
        // location.reload();
    }
}


function InjectElement(ID, content) {
    document.getElementById(ID).innerHTML = content
}