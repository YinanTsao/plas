#!/bin/bash

wget https://github.com/rtCamp/httperf/archive/master.zip && \
unzip master.zip && \
cd httperf-master && \
apt install -y autoconf && \
apt install -y libtool automake autoconf nasm pkgconf && \
autoreconf -i && \
mkdir build && \
cd build && \
../configure && \
make && \
make install && \
httperf -v | grep open