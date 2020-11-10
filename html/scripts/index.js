import Utils from "./utils.js";

const GROTHS_IN_BEAM = 100000000;
const STAKE_FEE = 100;
const TIMEOUT_VALUE = 3000;
const WS_PATH = "ws://3.136.182.25/ws";

class RedEnvelope {
    constructor() {
        this.errTimeout = null;
        this.connectionTimeout = null;
        this.socket =  new WebSocket(WS_PATH);
        this.envelopeData = {
            deposit_clicked: false,
            deposit_in_progress: false,
            deposit: 0,
            env_address: null,
            wallet_address: null,
            remaining: 0,
            incoming: 0,
            paid_reward: 0,
            outgoing_reward: 0,
            available_reward: 0,
            taken_amount: null
        };
    }

    initEnvelopeData = (params) => {
        this.envelopeData.env_address = params.envelope_addr;
        this.envelopeData.remaining = this.convertGrothsToBeam(params.envelope_remaining);
        this.envelopeData.incoming = this.convertGrothsToBeam(params.envelope_incoming);
        this.envelopeData.paid_reward = this.convertGrothsToBeam(params.paid_reward);
        this.envelopeData.outgoing_reward = this.convertGrothsToBeam(params.outgoing_reward);
        this.envelopeData.available_reward = this.convertGrothsToBeam(params.available_reward);

        if (this.envelopeData.incoming === 0 && this.envelopeData.deposit_clicked) {
            this.envelopeData.deposit_in_progress = false;
            this.envelopeData.deposit_clicked = false;
        }

        if (this.envelopeData.incoming > 0 && !this.envelopeData.deposit_clicked) {
            this.envelopeData.deposit_in_progress = true;
        }

        if (params.taken_amount) {
            this.envelopeData.taken_amount = this.convertGrothsToBeam(params.taken_amount);
        }

        this.updateEnvelopeView();
    }

    hideViews = () => {
        Utils.hide('envelope-catched-main');
        Utils.hide('first-deposit-main');
        Utils.hide('deposited-main');
        Utils.hide('withdraw-main');
        Utils.hide('second-deposit-main');
    }

    updateEnvelopeView = () => {
        this.hideViews();

        Utils.setText('in-envelope', this.envelopeData.remaining);
        Utils.setText('reward', this.envelopeData.available_reward);
        Utils.setText('incoming', this.envelopeData.incoming);
        Utils.setText('deposited', this.convertGrothsToBeam(this.envelopeData.deposit));

        if (this.envelopeData.taken_amount !== null && this.envelopeData.taken_amount > 0) {
            if (this.envelopeData.available_reward === 0) {
                Utils.show('withdraw-main');
            } else {
                Utils.show('envelope-catched-main');
                Utils.setText('catched-value', `${this.envelopeData.taken_amount} BEAM`);
            }
        } else {
            if (this.envelopeData.paid_reward === 0) {
                Utils.show('first-deposit-main');
            } else {
                if (this.envelopeData.deposit > 0 &&
                    !this.envelopeData.deposit_in_progress) {
                    Utils.show('deposited-main');
                } else {
                    Utils.show('second-deposit-main');
                }
            }
        }
    }
    
    convertGrothsToBeam = (value) => {
        const bigValue = new Big(value);
        const result = bigValue.div(GROTHS_IN_BEAM);
        return parseFloat(result);
    };
    
    reconnect = (now) => {
        if (this.connectionTimeout) {
            clearTimeout(this.connectionTimeout);
            this.connectionTimeout = null;
        }
        Utils.hide('envelope');

        this.socket = null;
        this.connectionTimeout = setTimeout(this.connect, now ? 0 : TIMEOUT_VALUE);
    }

    connect = () => {
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

        this.socket.onclose = (evt) => {
            if (evt.code == 1000)  {
                this.reconnect()
            } else {
                this.reconnect()
            }
        }

        this.socket.onmessage = (evt) => {
            let msg = JSON.parse(evt.data);
            
            if (msg.error) {
                this.reconnect();
                return
            }

            if (msg.method == "game-status") {
                this.initEnvelopeData(msg.params);

                const currentDate = new Date();
                var options = { year: 'numeric', month: 'long', day: 'numeric' };
                Utils.setText('reloaded-at', 
                    `(last updated on ${currentDate.toLocaleDateString("en-US", options)} at 
                    ${currentDate.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })})`);
                Utils.show('envelope')
            }
        }
    }

    applyStylesFromApi = (beamAPI) => {
        document.body.style.backgroundImage = `linear-gradient(to bottom, 
            ${beamAPI.style.background_main_top} -130px, 
            ${beamAPI.style.background_main} 100px, 
            ${beamAPI.style.background_main}`;
        document.body.style.color = beamAPI.style.content_main;
        document.querySelectorAll('.popup').forEach(item => {
            item.style.backgroundColor = Utils.hex2rgba(beamAPI.style.background_main, 0.6);
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
        redEnvelope.envelopeData.deposit_in_progress = true;
        redEnvelope.envelopeData.deposit_clicked = true;
        redEnvelope.envelopeData.deposit += parseInt(value);
        beamAPI.sendBEAM("BEAM Envelope", redEnvelope.envelopeData.env_address, parseInt(value), STAKE_FEE);
        Utils.hide('deposit-popup');
    })

    document.querySelectorAll('.container__main__controls__catch').forEach(item => {
        item.addEventListener('click', event => {
            event.preventDefault();
            redEnvelope.socket.send(JSON.stringify({
                jsonrpc: "2.0",
                id:      "take",
                method:  "take",
                params: {
                    user_addr: redEnvelope.envelopeData.wallet_address
                }
            }))
        })
    });

    document.querySelectorAll('.container__main__controls__withdraw').forEach(item => {
        item.addEventListener('click', event => {
            Utils.show('withdraw-popup');
        })
    });

    document.querySelectorAll('.container__main__controls__deposit').forEach(item => {
        item.addEventListener('click', event => {
            Utils.show('deposit-popup');
        })
    });

    Utils.getById('cancel-button-popup-with').addEventListener('click', (ev) => {
        Utils.hide('withdraw-popup');
    })

    Utils.getById('cancel-button-popup-dep').addEventListener('click', (ev) => {
        Utils.hide('deposit-popup');
    })

    Utils.getById('withdraw-button-popup').addEventListener('click', (ev) => {
        ev.preventDefault();
        redEnvelope.socket.send(JSON.stringify({
            jsonrpc: "2.0",
            id:      "withdraw",
            method:  "withdraw",
            params: {
                user_addr: redEnvelope.envelopeData.wallet_address
            }
        }));
        Utils.hide('withdraw-popup');
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

    beamAPI.permanentAddressGenerated.connect((address) => {
        redEnvelope.envelopeData.wallet_address = address;
        redEnvelope.connect();
    })
    
    // Go!
    beamAPI.generatePermanentAddress("BEAM Envelope Withdraw");
})