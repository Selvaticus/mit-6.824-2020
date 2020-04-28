#!/bin/sh

#
# basic map-reduce test
#

RACE=

# uncomment this to run the tests with the Go race detector.
RACE=-race

DEBUG=true

# run the test in a fresh sub-directory.
rm -rf mr-tmp
mkdir mr-tmp || exit 1
cd mr-tmp || exit 1
rm -f mr-*

# make sure software is freshly built.
(cd ../../mrapps && go build $RACE -buildmode=plugin crash.go) || exit 1
(cd ../../mrapps && go build $RACE -buildmode=plugin nocrash.go) || exit 1
(cd .. && go build $RACE mrmaster.go) || exit 1
(cd .. && go build $RACE mrworker.go) || exit 1
(cd .. && go build $RACE mrsequential.go) || exit 1

failed_any=0

echo '***' Generating correct output of crash test

# generate the correct output
../mrsequential ../../mrapps/nocrash.so ../pg*txt || exit 1
sort mr-out-0 > mr-correct-crash.txt
rm -f mr-out*

echo '***' Starting crash test.

(DEBUG=$DEBUG timeout -k 2s 180s ../mrmaster ../pg*txt ; touch mr-done ) 2>&1 | awk '{ print "\033[31m"$0"\033[0m" }' &
sleep 1

# start multiple workers
DEBUG=$DEBUG timeout -k 2s 180s ../mrworker ../../mrapps/crash.so 2>&1 | awk '{ print "\033[32m"$0"\033[0m" }' &

# mimic rpc.go's masterSock()
SOCKNAME=/var/tmp/824-mr-`id -u`

( while [ -e $SOCKNAME -a ! -f mr-done ]
  do
    DEBUG=$DEBUG timeout -k 2s 180s ../mrworker ../../mrapps/crash.so
    sleep 1
  done ) 2>&1 | awk '{ print "\033[33m"$0"\033[0m" }' &

( while [ -e $SOCKNAME -a ! -f mr-done ]
  do
    DEBUG=$DEBUG timeout -k 2s 180s ../mrworker ../../mrapps/crash.so
    sleep 1
  done ) 2>&1 | awk '{ print "\033[34m"$0"\033[0m" }' &

while [ -e $SOCKNAME -a ! -f mr-done ]
do
  DEBUG=$DEBUG timeout -k 2s 180s ../mrworker ../../mrapps/crash.so 2>&1 | awk '{ print "\033[35m"$0"\033[0m" }'
  sleep 1
done

wait
wait
wait

rm $SOCKNAME
sort mr-out* | grep . > mr-crash-all
if cmp mr-crash-all mr-correct-crash.txt
then
  echo '---' crash test: PASS
else
  echo '---' crash output is not the same as mr-correct-crash.txt
  echo '---' crash test: FAIL
  failed_any=1
fi

if [ $failed_any -eq 0 ]; then
    echo '***' PASSED ALL TESTS
else
    echo '***' FAILED SOME TESTS
    exit 1
fi
