#!/bin/bash
INPUT="/public/home/link/chart"
cd /opt/automation/extractor
inotifywait -m -r -q $INPUT -e close_write --format '%w%f'|while read newfile
do
    echo "==> $newfile"
./extractor \
    -inputpath="$newfile" \
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
    -watch=false \
    -namepattern="^.*\\.(xlsx|xlsm|xls)$" \
    -interval=5s
done
