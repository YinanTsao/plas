#!/bin/bash
timestamp=$(date +"%Y%m%d_%H%M%S")
path="$HOME/httperf_results"
server=130.104.229.12
port=31112
uri="/function/tngo"


# --print-reply --print-request

httperf --hog --server $server --port $port --uri $uri --add-header='Content-Type:application/json\n' --wsesslog 500,0,session1.txt --period e0.2 > $path/results_$timestamp.txt