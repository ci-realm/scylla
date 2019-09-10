
ElmScroll = { app: undefined }

ElmScroll.scrollIntoView = function (id) {
    var element = document.getElementById(id)
    console.debug({ id, element })
    element.scrollIntoView({ behavior: "smooth", block: "end", inline: "nearest" })
}

ElmScroll.init = function (app) {
    ElmScroll.app = app;

    app.ports.scrollIntoView.subscribe(ElmScroll.scrollIntoView);
}
