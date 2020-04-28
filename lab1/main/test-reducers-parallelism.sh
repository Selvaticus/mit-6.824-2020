#!/bin/sh

#
# Test reducer parallelism
#

RACE=

# uncomment this to run the tests with the Go race detector.
RACE=-race

# uncomment this to run with debug log on
DEBUG=true

# run the test in a fresh sub-directory.
rm -rf mr-tmp
mkdir mr-tmp || exit 1
cd mr-tmp || exit 1
rm -f mr-*

# make sure software is freshly built.

echo '***' Building reduce parallelism test...

(cd ../../mrapps && go build $RACE -buildmode=plugin rtiming.go) || exit 1
(cd .. && go build $RACE mrmaster.go) || exit 1
(cd .. && go build $RACE mrworker.go) || exit 1

failed_any=0

echo Done.

echo '***' Starting reduce parallelism test.

rm -f mr-out* mr-worker*

DEBUG=$DEBUG timeout -k 2s 180s ../mrmaster ../pg*txt 2>&1 | awk '{ print "\033[31m"$0"\033[0m" }' &
sleep 1

DEBUG=$DEBUG timeout -k 2s 180s ../mrworker ../../mrapps/rtiming.so 2>&1 | awk '{ print "\033[34m"$0"\033[0m" }'&
DEBUG=$DEBUG timeout -k 2s 180s ../mrworker ../../mrapps/rtiming.so 2>&1 | awk '{ print "\033[33m"$0"\033[0m" }'

NT=`cat mr-out* | grep '^[a-z] 2' | wc -l | sed 's/ //g'`
if [ "$NT" -lt "2" ]
then
  echo '---' too few parallel reduces.
  echo '---' reduce parallelism test: FAIL
  failed_any=1
else
  echo '---' reduce parallelism test: PASS
fi

wait ; wait