#!/usr/bin/env bash

cd gossiper
../clean_test.sh

#cd ../part1
#../clean_test.sh

cd ../..
cp -r peerster peerster_backup

cd peerster
rm .DS_Store
rm gossiper/.DS_Store
rm -rf .git
rm -rf .idea

cd ..

tar -cvzf hw3.tar.gz peerster
rm -rf peerster
mv peerster_backup peerster