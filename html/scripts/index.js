import Utils from "./utils.js";

const GROTHS_IN_BEAM = 100000000;
const STAKE_FEE = 100000;
const TIMEOUT_VALUE = 1000;
const WS_PATH = "ws://3.136.182.25/ws";
const ADDR_COMMENT = "BEAM Envelope Withdraw";
const DEPOSIT_COMMENT = "BEAM Red Envelope Deposit";

class RedEnvelope {
    constructor() {
        this.errTimeout = null;
        this.connectionTimeout = null;
        this.socket = null;
        this.envelopeData = {
            deposit: 0,
            env_address: null,
            wallet_address: null,
            remaining: 0,
            incoming: 0,
            paid_reward: 0,
            outgoing_reward: 0,
            available_reward: 0,
            taken_amount: null,
            is_withdraw_active: true,
            is_catch_active: true,
            is_deposit_active: true,
            wallet_status_available: 0,
            is_deposit_in_progress: false,
            is_deposit_finished: null
        };
    }

    initEnvelopeData = (params) => {
        this.envelopeData.env_address = params.envelope_addr;
        this.envelopeData.remaining = this.convertGrothsToBeam(params.envelope_remaining);
        this.envelopeData.incoming = this.convertGrothsToBeam(params.envelope_incoming);
        this.envelopeData.paid_reward = this.convertGrothsToBeam(params.paid_reward);
        this.envelopeData.outgoing_reward = this.convertGrothsToBeam(params.outgoing_reward);
        this.envelopeData.available_reward = this.convertGrothsToBeam(params.available_reward);

        if (params.taken_amount) {
            this.envelopeData.taken_amount = this.convertGrothsToBeam(params.taken_amount);
        } else {
            this.envelopeData.taken_amount = null;
        }

        this.updateEnvelopeView();
    }

    hideViews = () => {
        Utils.hide('envelope-catched-main');
        Utils.hide('first-deposit-main');
        Utils.hide('deposited-main');
        Utils.hide('withdraw-main');
        Utils.hide('second-deposit-main');
        Utils.hide('deposit-in-progress-main');
    }

    updateEnvelopeView = () => {
        this.hideViews();

        Utils.setText('in-envelope', this.envelopeData.remaining);
        Utils.setText('reward', this.envelopeData.available_reward);
        Utils.setText('incoming', this.envelopeData.incoming);
        Utils.setText('deposited', this.convertGrothsToBeam(this.envelopeData.deposit));


        if (this.envelopeData.is_deposit_in_progress && !this.envelopeData.is_deposit_finished) {
            Utils.show('deposit-in-progress-main');
        } else if (!this.envelopeData.is_deposit_in_progress && this.envelopeData.is_deposit_finished) {
            Utils.show('deposited-main');
        } else if (!this.envelopeData.is_deposit_in_progress && this.envelopeData.is_deposit_finished == null) {
            if (this.envelopeData.available_reward > 0) {
                this.envelopeData.is_withdraw_active = true;
                Utils.show('envelope-catched-main');
                this.notEnoughAlertUpdate('catched-not-enough-alert', 'catched-deposit-button');
                Utils.setText('catched-value', `${this.envelopeData.available_reward} BEAM`);
            } else {
                Utils.removeClassById('withdraw-button-popup', 'disabled');
                if (this.envelopeData.outgoing_reward > 0) {
                    this.notEnoughAlertUpdate('with-not-enough-alert', 'with-deposit-button');
                    Utils.hide('withdraw-popup');
                    Utils.show('withdraw-main');
                } else {
                    if (this.envelopeData.remaining === 0) {
                        Utils.show('first-deposit-main');
                        this.notEnoughAlertUpdate('first-not-enough-alert', 'first-deposit-button');
                    } else {
                        if (!this.envelopeData.taken_amount || this.envelopeData.taken_amount === 0) {
                            Utils.hide('catch-more-after');
                            this.envelopeData.is_catch_active = true;
                            Utils.removeClassById('welcome-catch-button', 'disabled');
                            Utils.removeClassById('dep-catch-button', 'disabled');
                        } else {
                            Utils.show('catch-more-after');
                            this.envelopeData.is_catch_active = false;
                            Utils.addClassById('welcome-catch-button', 'disabled');
                            Utils.addClassById('dep-catch-button', 'disabled');
                        }
                        Utils.show('second-deposit-main');
                        this.notEnoughAlertUpdate('welcome-not-enough-alert', 'welcome-deposit-button');
                    }
                }
            }
        }
    }

    notEnoughAlertUpdate = (id, buttonId) => {
        if (this.envelopeData.wallet_status_available > 0) {
            Utils.hide(id);
            this.envelopeData.is_deposit_active = true;
            Utils.removeClassById(buttonId, 'disabled');
        } else {
            Utils.show(id);
            this.envelopeData.is_deposit_active = false;
            Utils.addClassById(buttonId, 'disabled');
        }
    }

    setError = (text) => {
        if (text) {
            if (text[text.length-1] !== '.') text += '.'
            text += " Restarting..."
        }

        Utils.setText('error', text)
        if (this.errTimeout) {
            clearTimeout(this.errTimeout)
        }
        this.errTimeout = setTimeout(() => this.setError(""), 3000)
    }
    
    convertGrothsToBeam = (value) => {
        const bigValue = new Big(value);
        const result = bigValue.div(GROTHS_IN_BEAM);
        return result.toFixed();
    };

    restart = (now) => {
        if (this.socket) {
            this.socket.close();
        }

        if (this.connectionTimeout) {
            clearTimeout(this.connectionTimeout)
            this.connectionTimeout = null;
        }

        Utils.hide('envelope');

        if (this.socket) {
            this.socket.close()
            this.socket = null
        }

        this.connectionTimeout = setTimeout(this.start, now ? 0 : TIMEOUT_VALUE)   
    }

    start = () => {
        Utils.callApi({
            "jsonrpc": "2.0",
            "id":      "addr_list",
            "method":  "addr_list",
            "params":  {
                "own": true
            }
        })
    }
    
    reconnect = (now) => {
        if (this.connectionTimeout) {
            clearTimeout(this.connectionTimeout);
            this.connectionTimeout = null;
        }
        Utils.hide('envelope');

        if (this.socket) {
            this.socket.close()
            this.socket = null
        }

        this.connectionTimeout = setTimeout(this.connect, now ? 0 : TIMEOUT_VALUE);
    }

    connect = () => {
        this.socket = new WebSocket(WS_PATH)
        this.socket.onopen = (evt) => {
            this.socket.send(JSON.stringify({
                jsonrpc: "2.0",
                id:      "login",
                method:  "login",
                params: {
                    user_addr: this.envelopeData.wallet_address
                }
            }));
        }

        this.socket.onerror = (event) => {
            this.socket.close();
        };

        this.socket.onclose = (evt) => {
            if (evt.code == 1000)  {
                this.setError('Connection closed');
                this.reconnect(false);
            } else {
                this.setError('Connection error');
                this.reconnect();
            }
        }

        this.socket.onmessage = (evt) => {
            let msg = JSON.parse(evt.data);
            
            if (msg.error) {
                this.setError(["Server error: ", msg.error.code, ", ", msg.error.message].join(''))
                this.reconnect()
                return
            }

            if (msg.method == "game-status") {
                this.getWalletStatus();
                this.getTxStatus();
                this.initEnvelopeData(msg.params);

                const currentDate = new Date();
                var options = { year: 'numeric', month: 'long', day: 'numeric' };
                Utils.setText('reloaded-at', 
                    `(last updated on ${currentDate.toLocaleDateString("en-US", options)} at 
                    ${currentDate.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })})`);

                this.setError("")
                Utils.show('envelope')
            }
        }
    }

    getWalletStatus = () => {
        Utils.callApi({
            "jsonrpc":"2.0", 
            "id": "wallet_status",
            "method":"wallet_status"
        })
    }

    getTxStatus = () => {
        Utils.callApi({
            "jsonrpc":"2.0", 
            "id": "tx_list",
            "method":"tx_list"
        })
    }

    applyStylesFromApi = (beamAPI) => {
        const topColor = [beamAPI.style.appsGradientOffset || -130, "px,"].join('');
        const mainColor = [beamAPI.style.appsGradientTop || 100, "px,"].join('');
        document.body.style.backgroundImage = `linear-gradient(to bottom, 
            ${beamAPI.style.background_main_top} ${topColor}
            ${beamAPI.style.background_main} ${mainColor}
            ${beamAPI.style.background_main}`;
        document.body.style.color = beamAPI.style.content_main;
        document.querySelectorAll('.popup').forEach(item => {
            item.style.backgroundImage = `linear-gradient(to bottom, 
                ${Utils.hex2rgba(beamAPI.style.background_main_top, 0.6)} ${topColor}
                ${Utils.hex2rgba(beamAPI.style.background_main, 0.6)} ${mainColor}
                ${Utils.hex2rgba(beamAPI.style.background_main, 0.6)}`;
        });
        document.querySelectorAll('.popup__content').forEach(item => {
            item.style.backgroundColor = Utils.hex2rgba(beamAPI.style.background_popup, 1);
        });
    }
}

Utils.getById('refresh').addEventListener('click', () => {
    Utils.reload();
    return false
})

Utils.onLoad(async (beamAPI) => {
    let redEnvelope = new RedEnvelope();
    redEnvelope.applyStylesFromApi(beamAPI);

    Utils.getById('deposit-button-popup').addEventListener('click', (ev) => {
        ev.preventDefault();
        const bigValue = new Big(Utils.getById('deposit-input').value);
        const value = bigValue.times(GROTHS_IN_BEAM);
        redEnvelope.envelopeData.is_deposit_in_progress = true;
        redEnvelope.envelopeData.is_deposit_finished = false;
        redEnvelope.envelopeData.deposit += parseInt(value);
        Utils.callApi({
            "jsonrpc": "2.0",
            "id":      "tx_send",
            "method":  "tx_send",
            "params":  {
                "value": parseInt(value),
                "fee": STAKE_FEE,
                "from": redEnvelope.envelopeData.wallet_address,
                "address": redEnvelope.envelopeData.env_address,
                "comment": DEPOSIT_COMMENT,
            }
        })

        Utils.hide('deposit-popup');
    })

    document.querySelectorAll('.container__main__controls__catch').forEach(item => {
        item.addEventListener('click', event => {
            event.preventDefault();
            if (redEnvelope.envelopeData.is_catch_active) {
                redEnvelope.envelopeData.is_deposit_finished = null;
                redEnvelope.envelopeData.is_catch_active = false;
                Utils.addClassById('welcome-catch-button', 'disabled');
                Utils.addClassById('dep-catch-button', 'disabled');
                
                redEnvelope.socket.send(JSON.stringify({
                    jsonrpc: "2.0",
                    id:      "take",
                    method:  "take",
                    params: {
                        user_addr: redEnvelope.envelopeData.wallet_address
                    }
                }));
            }
        })
    });

    document.querySelectorAll('.container__main__controls__withdraw').forEach(item => {
        item.addEventListener('click', event => {
            Utils.show('withdraw-popup');
            Utils.setText('catched-value-popup', `${redEnvelope.envelopeData.available_reward} BEAM`);
        })
    });

    document.querySelectorAll('.container__main__controls__deposit').forEach(item => {
        item.addEventListener('click', event => {
            if (redEnvelope.envelopeData.is_deposit_active) {
                Utils.show('deposit-popup');
            }
        })
    });

    Utils.getById('cancel-button-popup-with').addEventListener('click', (ev) => {
        Utils.hide('withdraw-popup');
    })

    Utils.getById('cancel-button-popup-dep').addEventListener('click', (ev) => {
        Utils.hide('deposit-popup');
    })

    Utils.getById('withdraw-button-popup').addEventListener('click', (ev) => {
        if (redEnvelope.envelopeData.is_withdraw_active) {
            ev.preventDefault();
            redEnvelope.envelopeData.is_withdraw_active = false;
            Utils.addClassById('withdraw-button-popup', 'disabled');
            redEnvelope.socket.send(JSON.stringify({
                jsonrpc: "2.0",
                id:      "withdraw",
                method:  "withdraw",
                params: {
                    user_addr: redEnvelope.envelopeData.wallet_address
                }
            }));
            Utils.hide('withdraw-popup');
        }
    })

    Utils.getById('deposit-input').addEventListener('keydown', (event) => {
        const specialKeys = [
            'Backspace', 'Tab', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'ArrowUp',
            'Control', 'Delete', 'F5'
          ];

        if (specialKeys.indexOf(event.key) !== -1) {
            return;
        }

        const current = Utils.getById('deposit-input').value;
        const next = current.concat(event.key);
      
        if (!Utils.handleString(next)) {
            event.preventDefault();
        }
    })

    Utils.getById('deposit-input').addEventListener('paste', (event) => {
        const text = event.clipboardData.getData('text');
        if (!Utils.handleString(text)) {
            event.preventDefault();
        }
    })

    // Go!
    beamAPI.api.callWalletApiResult.connect((json) => {
        let res = undefined;
        let err = undefined;

        try {
            res = JSON.parse(json);
            err = JSON.stringify(res.error);
        } catch (e) {
            err = e.toString();
        } 

        if (err) {
            redEnvelope.restart();
            return
        }

        if (res.id === "addr_list") {
            for (let idx = 0; idx < res.result.length; ++idx) {
                let addr = res.result[idx];
                if (addr.comment == ADDR_COMMENT) {
                    redEnvelope.envelopeData.wallet_address = addr.address;
                    break;
                }
            }

            if (!redEnvelope.envelopeData.wallet_address) {
                Utils.callApi({
                    "jsonrpc": "2.0",
                    "id":      "create_address",
                    "method":  "create_address",
                    "params":  {
                        "expiration": "never",
                        "comment": ADDR_COMMENT
                    }
                })
            } else {
                redEnvelope.connect();
            }
        }

        if (res.id === "create_address") {
            redEnvelope.envelopeData.wallet_address = res.result;
            redEnvelope.restart();
        }

        if (res.id === "wallet_status") {
            redEnvelope.envelopeData.wallet_status_available = res.result.available;
        }

        if (res.id === "tx_list") {
            const transacions = res.result;

            const depositTrasaction = transacions.find((item) => {
                return item.comment === DEPOSIT_COMMENT && (item.status === 0 || item.status === 1 || item.status === 5);
            })

            if (depositTrasaction !== undefined) {
                redEnvelope.envelopeData.is_deposit_in_progress = true;
                redEnvelope.envelopeData.is_deposit_finished = false;
            } else {
                if (redEnvelope.envelopeData.is_deposit_in_progress) {
                    redEnvelope.envelopeData.is_deposit_in_progress = false;
                    redEnvelope.envelopeData.is_deposit_finished = true;
                }
            }
        }
    });

    redEnvelope.getTxStatus();
    redEnvelope.getWalletStatus();
    redEnvelope.start();
})