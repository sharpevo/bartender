#!/bin/bash
cd /opt/automation/extractor
./extractor \
    -inputpath=/public/home/link/chart \
    -sheet=1 \
    -rowstart=2 \
    -rowend=-1 \
    -columns=1,2,3,4,5,9,11 \
    -outputtype=txt \
    -outputpath=output \
    -remotepath=/root/upload \
    -transfer=true \
    -hostkey="***REMOVED***" \
    -username=root \
    -password=***REMOVED*** \
    -watch=true \
    -interval=5s
