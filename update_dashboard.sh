#!/bin/bash

./stop_webservice.sh

rm -rf *

cp -R /home/golang/git/dashboard/* .
chown http -R *

./start_webservice.sh

