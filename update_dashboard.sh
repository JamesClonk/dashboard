#!/bin/bash

./stop_dashboard.sh

rm -rf *

cp -R /home/golang/git/dashboard/* .
chown http -R *

./start_dashboard.sh

