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
rm part1/.DS_Store
rm -rf .git
rm -rf .idea

rm _Downloads/*.pdf
rm _Downloads/*.txt

cd ..

tar -cvzf hw3.tar.gz peerster
rm -rf peerster
mv peerster_backup peerster