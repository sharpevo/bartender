#!/bin/bash
source ~/go/src/automation/password.rc
../transmitter \
    -sourcepath=/tmp/input \
    -watchpath=/tmp/input \
    -transfer=true \
    -remotepath=/home/igenetech/testauto \
    -hostkey="$LAN_KEY" \
    -username=igenetech \
    -watch=true \
    -loglevel=debug \
    -password=$LAN_PWD
    #-namepattern=".\.abc$" \
