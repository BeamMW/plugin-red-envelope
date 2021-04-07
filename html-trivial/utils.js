export default class Utils {
    static reload () {
        window.location.reload()
    }

    //
    // API Exposed by the wallet itself
    //
    static BEAM = null

    static onLoad(cback) {
        window.addEventListener('load', () => new QWebChannel(qt.webChannelTransport, (channel) => {
            Utils.BEAM = channel.objects.BEAM
            cback(Utils.BEAM)
        }))
    }

    static byId(id) {
        return document.querySelector(['#', id].join(''))
    }

    static callApi(obj) {
        Utils.BEAM.api.callWalletApi(JSON.stringify(obj))
    }
}
