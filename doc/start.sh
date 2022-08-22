#!/bin/sh
# Copyright (C) 2020 The Takeout Authors.

if [ -z "$GOPATH" ]; then
    PATH=$HOME/go/bin:$PATH
else
    PATH=${GOPATH//://bin:}/bin:$PATH
fi

TAKEOUT_DIR=.
RUN_DIR=$TAKEOUT_DIR
LOG_DIR=$TAKEOUT_DIR

mkdir -p $RUN_DIR
mkdir -p $LOG_DIR

LOG_FILE=$LOG_DIR/takeout.log
PID_FILE=$RUN_DIR/takeout.pid

# kill old instance if running
if [ -f $PID_FILE ]; then
    pid=`head -1 $PID_FILE | egrep '[0-9]'`
    if [ $pid -gt 1 ]; then
	kill $pid
    fi
    rm -f $PID_FILE
fi

exec takeout serve < /dev/null >> $LOG_FILE 2>&1 &
echo $! > $PID_FILE
