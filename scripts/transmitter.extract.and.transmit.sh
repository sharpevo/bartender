#!/bin/bash
cd /opt/automation/transmitter
./transmitter \
    -sourcepath=$1 \
    -namepattern=".\.abc$" \
    -transfer=true \
    -remotepath=/opt/automation/transmitter/test \
    -hostkey="***REMOVED***" \
    -username=igenetech \
    -watch=false \
    -loglevel=debug \
    -password=***REMOVED***
