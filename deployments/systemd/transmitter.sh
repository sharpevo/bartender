#!/bin/bash

INPUT="/opt/automation/transmitter/input"
cd /opt/automation/transmitter
inotifywait -m -r -q $INPUT -e close_write --format '%w%f'|while read newfile
do
./transmitter \
    -sourcepath=$newfile \
    -namepattern=".\.abc$" \
    -transfer=true \
    -remotepath=/opt/automation/transmitter/test \
    -hostkey="***REMOVED***" \
    -username=igenetech \
    -watch=false \
    -password=***REMOVED***
#-loglevel=debug \
done
