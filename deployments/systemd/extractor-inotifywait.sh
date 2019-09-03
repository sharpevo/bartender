#!/bin/bash
source ~/go/src/automation/password.rc
INPUT="/public/home/link/chart"
cd /opt/automation/extractor
inotifywait -m -r -q $INPUT -e close_write --format '%w%f'|while read newfile
do
    echo "==> $newfile"
./extractor \
    -inputpath=$newfile \
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
    -watch=false \
    -interval=5s
done
