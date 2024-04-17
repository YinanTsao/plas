#!bin/bash

wget https://github.com/rtCamp/httperf/archive/master.zip && \
unzip master.zip && \
cd httperf-master && \
autoreconf -i && \
mkdir build && \
cd build && \
../configure && \
make && \
make install