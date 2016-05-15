/**
 * Created by i.navrotskyj on 26.01.2015.
 */

var MongoDb = require("mongodb"),
    MongoClient = MongoDb.MongoClient,
    format = require('util').format,
    config = require('../conf'),
    log = require('../lib/log')(module)
    ;


class Drv {
    _initDB (db) {
        this.db = db;
        return this.db;
    }

    getCollection (name) {
        try {
            return this.db.collection(name)
        } catch (e) {
            log.error(`mongodb error: ${e.message}`);
        }
    }
}

let drv = new Drv(),
    timerId = null;

function connect () {
    if (timerId)
        clearTimeout(timerId);

    let mongodbClient = new MongoClient(),
        option = {
            server: {
                auto_reconnect: true
            }
        }
    ;

    mongodbClient.connect(config.get('mongodb:uri'), option, function(err, db) {
        if (err) {
            log.error('Connect db error: %s', err.message);
            return timerId = setTimeout(connect, 1000);
        };
        drv._initDB(db);

        log.info('Connected db %s ', config.get('mongodb:uri'));
        db.on('close', function () {
            log.error('close mongo');
        });

        db.on('error', function (err) {
            log.error(err);
        });
    });
}

connect();

module.exports = drv;