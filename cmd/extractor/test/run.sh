#!/bin/bash
source ~/go/src/automation/password.rc
../extractor \
    -inputpath=input \
    -sheet=1 \
    -rowstart=2 \
    -rowend=-1 \
    -columns=1,2,3,4,5,9,11 \
    -outputtype=txt \
    -outputpath=output/ \
    -remotepath=/root/wuy/testauto \
    -transfer=true \
    -hostkey="$ECS_KEY" \
    -username=root \
    -password=$ECS_PWD \
    -watch=false \
    -loglevel=debug
#-namepattern="^.*\\.(xlsx|xlsm|xls$"
