#!/bin/bash

siege -c5 -r10 --content-type "application/json" 'http://130.104.229.12:31112/function/tngo < payload.txt'