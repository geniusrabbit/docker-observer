#!/usr/bin/env bash

function _md5() {
  if [ -z "$(which md5)" ]; then
    md5sum "$1" | cut -d ' ' -f 1
  else
    md5 -q $1
  fi
}

function doreload() {
  echo "Nginx reload config: $1"
  nginx -g 'daemon on; master_process on;' -s reload
}

sum=`_md5 $1`

if [ -f "$1.checksum" ]; then
  prevSum=`cat $1.checksum`

  if [ $sum != $prevSum ]; then
    doreload $1
  else
    echo "No any changes: $1"
  fi
else
  echo $sum >> "$1.checksum"
fi
