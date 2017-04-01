/**
 * Created by igor on 27.03.17.
 */

"use strict";

const CallTreeInterator = require('./iterator'),
    log = require(__appRoot + '/lib/log')(module),
    moment = require('moment-timezone')
    ;

class Call {
    constructor (conn, shema, acr) {
        this._routeLog = [];
        this._id = conn._id;
        this._uuid = conn.channelData.getHeader('variable_uuid');

        if (!this._uuid) {
            this.log(`Not found uuid in ${this._id}`, true);
            this._uuid = this._id;
        }

        this.domain = shema.domain;
        this.timezone = shema.fs_timezone;
        this.callFlowIter = new CallTreeInterator(shema.callflow, acr);
        
        // this.bigData = new Array(1e6).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');


        this.execApp = (appName, data, options = {}, cb) => {
            if (!appName)
                return cb(new Error('Application name is required.'));

            if (options.async) {
                this.log(`Execute async app: ${appName}, with data: ${data}`);
                conn.setEventLock(false);
            } else {
                this.log(`Execute sync app: ${appName}, with data: ${data}`);
                conn.setEventLock(true);
            }

            conn.execute(appName, data || '', cb);
        };

        const end = () => {
            console.dir(this.logToJson(), {depth: 10, colors: true});
            // this.execApp('hangup', '');
            // return;
        };

        const exec = (err, res) => {
            if (err)
                this.log(err, true);

            let app = this.callFlowIter.next() || this.callFlowIter.getParent();
            if (!app) {
                return end();
            }
            app.execute(this, (err, res) => {
                if (app.break === true) {
                    this.log(`Break call flow`);
                    return end();
                }

                return exec(err, res);
            });

        };

        exec();

    }

    getDate () {
        if (!this.timezone) {
            return moment();
        }

        return moment().tz(this.timezone);
    }

    log(data, e) {
        this._routeLog.push({
            time: Date.now(),
            log: data
        });

        if (e)
            log.error(data);
        else log.trace(`[${this._uuid}]: ${data}`); //TODO to uuid
    }

    logToJson () {
        return JSON.stringify(this._routeLog);
    }
}

module.exports = Call;

