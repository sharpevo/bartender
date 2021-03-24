#!/bin/bash
source ~/go/src/github.com/sharpevo/bartender/password.rc
../extractor \
    -inputpath=/tmp/input \
    -namepattern="^.*\\.(xlsx|xlsm|xls|txt)$" \
    -extractpattern="^.*\\.(xlsx|xlsm|xls)$" \
    -sheet=1 \
    -rowstart=2 \
    -rowend=-1 \
    -columns=1,2,3,4,5,6 \
    -outputtype=txt \
    -outputpath=output/ \
    -remotepath=/home/igenetech/testauto \
    -transfer=true \
    -hostkey="$LAN_KEY" \
    -username=igenetech \
    -password=$LAN_PWD \
    -watch=true \
    -loglevel=debug
#-namepattern="^.*\\.(xlsx|xlsm|xls$"
    #-columns=1,2,3,4,5,9,11 \
