#!/bin/bash
# Use as many TCP port as possible
# This test contains 10 sessions, each session has 100 connections, and each connection sends 3 requests

# httperf \

# --hog \

# --server 130.104.229.36 \

# --port 31112 \

# --uri /function/tngo \

# --wsesslog 2,100,session.txt \

# --period e0.1 \

# --timeout 5

httperf --hog --print-reply --print-request --server 130.104.229.12 --port 31112 --uri /function/tngo --add-header='Content-Type:application/json\n' --wsesslog 200,100,session.txt --period e0.5