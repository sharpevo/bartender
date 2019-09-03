#!/bin/bash
source ~/go/src/automation/password.rc
../transmitter \
    -sourcepath=input \
    -namepattern=".\.abc$" \
    -transfer=true \
    -remotepath=/root/testauto \
    -hostkey="$ECS_KEY" \
    -username=root \
    -watch=true \
    -loglevel=debug \
    -password=$ECS_PWD
