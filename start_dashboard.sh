#!/bin/bash

# run as root
su http --command="PORT=4400 ./dashboard 1>>/var/log/martini/go_dashboard.log 2>&1 &" -s /bin/sh
