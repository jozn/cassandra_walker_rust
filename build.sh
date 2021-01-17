#!/usr/bin/env bash
# Run me with "bash build.sh"
echo "First running go-bindata."ba
bash bind-templates.sh
echo "======== Building Go Binary ========"
#echo "pwd: $(pwd)"
go fmt .
go build
echo "======== End of Building Go Binary ========"


# temp
#./cassandra_walker  twitter -d "/home/hamid/life/_active/backbone/src/"
#./cassandra_walker  system -d "/home/hamid/life/_active/backbone/micro/logic/src/"
#./cassandra_walker  twitter -c 37.152.182.28 -d "/home/hamid/life/_active/backbone/micro/logic/src/" > debug_gen_sample.txt
#./cassandra_walker  twitter -c 185.239.107.163 -d "/home/hamid/life/_active/backbone/micro/play/src/"

#dir="/home/hamid/life/_active/backbone/lib/shared"
#echo $dir
#dir2="$dir/src"
#./cassandra_walker  msgs -c 185.239.107.163 -d $dir2
#echo $dir2
#cd $dir2
#cargo fmt .
./cassandra_walker  msgs -c 185.239.107.163 -d "/home/hamid/life/_active/backbone/lib/shared/src"
