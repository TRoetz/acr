#!/bin/bash
set -e

echo 'Webitel ACR '$VERSION

if [ "$ACR_COUNT" ]; then
	COUNTER=10030
	ACR_COUNT_STR=10030
	let ACR_COUNT_MAX=$COUNTER+$ACR_COUNT-1
	while [  $COUNTER -lt $ACR_COUNT_MAX ]; do
		let COUNTER=COUNTER+1
		ACR_COUNT_STR=$ACR_COUNT_STR', '$COUNTER
	done
	sed -i 's/ACR_COUNT/'$ACR_COUNT_STR'/g' /acr/conf/config.json
	echo 'Starting on ports '$ACR_COUNT_STR
else
	sed -i 's/ACR_COUNT/10030/g' /acr/conf/config.json
	echo 'Starting on the port 10030'
fi

if [ "$LOGLEVEL" ]; then
	sed -i 's/LOGLEVEL/'$LOGLEVEL'/g' /acr/conf/config.json
	echo 'Loglevel set to '$LOGLEVEL
else
	sed -i 's/LOGLEVEL/warn/g' /acr/conf/config.json
	echo 'Loglevel set to warn'
fi

if [ "$LOGSTASH_ENABLE" ]; then
	sed -i 's/LOGSTASH_ENABLE/'$LOGSTASH_ENABLE'/g' /acr/conf/config.json
else
	sed -i 's/LOGSTASH_ENABLE/false/g' /acr/conf/config.json
fi

if [ "$LOGSTASH_HOST" ]; then
	sed -i 's/LOGSTASH_HOST/'$LOGSTASH_HOST'/g' /acr/conf/config.json
	echo 'LOGSTASH_HOST set to '$LOGSTASH_HOST
fi

if [ "$MONGODB_HOST" ]; then
	sed -i 's/MONGODB_HOST/'$MONGODB_HOST'/g' /acr/conf/config.json
	echo 'MONGODB_HOST set to '$MONGODB_HOST
fi

exec node app.js