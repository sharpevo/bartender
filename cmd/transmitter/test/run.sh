#!/bin/bash
source ~/go/src/automation/password.rc
../transmitter \
    -sourcepath=input \
    -transfer=true \
    -remotepath=/root/wuy/testauto \
    -hostkey="$ECS_KEY" \
    -username=root \
    -watch=true \
    -loglevel=debug \
    -password=$ECS_PWD
    #-namepattern=".\.abc$" \
