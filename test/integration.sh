#!/usr/bin/env bash

. test_lib.sh

cd ..
go build -race
cd client
go build
cd ../test

UIPort=12345
gossipPort=5000
name='A'

for i in `seq 1 5`;
do
	outFileName="$name.out"
	peerPort=$(($gossipPort+1))
	if [[ $gossipPort == 5004 ]] ; then
		peerPort=5000
	fi
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	../gossiper -UIPort=${UIPort} -gossipAddr=${gossipAddr} -name=${name} -peers=${peer} > ${outFileName} &
	outputFiles+=("$outFileName")
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

# network warm-up
sleep 1

../client/client -UIPort=12349 -msg=${message_c1_1}
../client/client -UIPort=12346 -msg=${message_c2_1}
../client/client -UIPort=12347 -msg=${message_c3_1}
sleep 1

../client/client -UIPort=12349 -msg=${message_c1_2}
../client/client -UIPort=12346 -msg=${message_c2_2}
../client/client -UIPort=12347 -msg=${message_c3_2}

sleep 1
../client/client -UIPort=12345 -msg="HOLA" -Dest="C"

wait_grep "E.out" "CLIENT $message_c1_1 E"
wait_grep "E.out" "CLIENT $message_c1_2 E"
wait_grep "B.out" "CLIENT $message_c2_1 B"
wait_grep "B.out" "CLIENT $message_c2_2 B"
wait_grep "C.out" "CLIENT $message_c3_1"
wait_grep "C.out" "CLIENT $message_c3_2"

gossipPort=5000
for i in `seq 0 4`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
		relayPort=5004
	fi
	nextPort=$(($gossipPort+1))
	if [[ "$nextPort" == 5005 ]] ; then
		nextPort=5000
	fi
	gossipPort=$(($gossipPort+1))
	wait_grep "${outputFiles[$i]}" ${message_c1_1}
	wait_grep "${outputFiles[$i]}" ${message_c1_2}
	wait_grep "${outputFiles[$i]}" ${message_c2_1}
	wait_grep "${outputFiles[$i]}" ${message_c2_2}
	wait_grep "${outputFiles[$i]}" ${message_c3_1}
	wait_grep "${outputFiles[$i]}" ${message_c3_2}
	#wait_grep "${outputFiles[$i]}" "IN SYNC WITH 127.0.0.1:$relayPort"
	#wait_grep "${outputFiles[$i]}" "IN SYNC WITH 127.0.0.1:$nextPort"
	#wait_grep "${outputFiles[$i]}" "MONGERING with 127.0.0.1:$relayPort"
	#wait_grep "${outputFiles[$i]}" "MONGERING with 127.0.0.1:$nextPort"
done

wait_grep "C.out" "HOLA"

pkill -f gossiper &
cleantest
