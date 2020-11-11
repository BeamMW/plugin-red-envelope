
const MIN_AMOUNT = 0.00000001;
const MAX_AMOUNT = 254000000;

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

    static getById = (id)  => {
        return document.getElementById(id);
    }

    static addClassById = (id, className) => {
        const elem = this.getById(id);
        elem.classList.add(className);
    }

    static removeClassById = (id, className) => {
        const elem = this.getById(id);
        elem.classList.remove(className);
    }

    static setText = (id, text) => {
        this.getById(id).textContent = text;
    }

    static show(id) {
        let obj = this.getById(id)
        obj.style.display="flex";
    }
    
    static hide(id) {
        let obj = this.getById(id)
        obj.style.display="none";
    }

    static hex2rgba = (hex, alpha = 1) => {
        const [r, g, b] = hex.match(/\w\w/g).map(x => parseInt(x, 16));
        return `rgba(${r},${g},${b},${alpha})`;
    };

    static callApi = (obj) => {
        Utils.BEAM.callWalletApi(JSON.stringify(obj));
    }

    static handleString(next) {
        let result = true;
        const regex = new RegExp(/^-?\d+(\.\d*)?$/g);
        const floatValue = parseFloat(next);
        const afterDot = next.indexOf('.') > 0 ? next.substring(next.indexOf('.') + 1) : '0';
        if ((next && !String(next).match(regex)) ||
            (String(next).length > 1 && String(next)[0] === '0' && next.indexOf('.') < 0) ||
            (parseInt(afterDot, 10) === 0 && afterDot.length > 7) ||
            (afterDot.length > 8) ||
            (floatValue === 0 && next.length > 1 && next[1] !== '.') ||
            (floatValue < 1 && next.length > 10) ||
            (floatValue > 0 && (floatValue < MIN_AMOUNT || floatValue > MAX_AMOUNT))) {
          result = false;
        }
        return result;
      }
}
