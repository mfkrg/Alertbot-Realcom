#!/bin/sh

touch /var/log/cron.log

echo "Starting cron..."

crond

tail -f /var/log/cron.log
