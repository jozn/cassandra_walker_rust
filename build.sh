#!/usr/bin/env bash
# Run me with "bash build.sh"
echo "First running go-bindata."ba
bash bind-templates.sh
echo "======== Building Go Binary ========"
#echo "pwd: $(pwd)"
go build
echo "======== End of Building Go Binary ========"


# temp
./cassandra_walker  twitter -d "/home/hamid/life/_active/backbone/src/"
