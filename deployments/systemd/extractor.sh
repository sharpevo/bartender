#!/bin/bash
BASE="/opt/automation/extractor"
source $BASE/password.rc
INPUT="/public/home/link/chart"
cd $BASE
./extractor \
    -inputpath="$INPUT" \
    -namepattern="^.*\\.(xlsx|xlsm|xls|txt)$" \
    -extractpattern="^.*\\.(xlsx|xlsm|xls)$" \
    -sheet=1 \
    -rowstart=2 \
    -rowend=-1 \
    -columns=1,2,3,4,5,9,11 \
    -outputtype=txt \
    -outputpath=output \
    -remotepath=/root/upload \
    -transfer=true \
    -hostkey="$ECS_KEY" \
    -username=root \
    -password=$ECS_PWD \
    -watch=true \
    -interval=10s
done
