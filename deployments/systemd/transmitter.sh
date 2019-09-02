#!/bin/bash

INPUT="/opt/automation/transmitter/input"
cd /opt/automation/transmitter
inotifywait -m -r -q $INPUT -e close_write --format '%w%f'|while read newfile
do
    echo "==> $newfile"
./transmitter \
    -sourcepath="$newfile" \
    -watchpath="$INPUT" \
    -transfer=true \
    -remotepath=/public/home/link/ecs \
    -hostkey="***REMOVED***" \
    -username=igenetech \
    -watch=false \
    -password=***REMOVED***
#-loglevel=debug \
#-namepattern=".\.abc$" \
done
