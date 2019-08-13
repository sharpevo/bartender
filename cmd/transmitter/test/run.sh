#!/bin/bash
../transmitter \
    -sourcepath=input \
    -namepattern=".\.abc$" \
    -transfer=true \
    -remotepath=/root/testauto \
    -hostkey="***REMOVED***" \
    -username=root \
    -watch=true \
    -loglevel=debug \
    -password=***REMOVED***
