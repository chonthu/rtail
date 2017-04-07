#!/bin/sh

gzip rtail
echo "Sha256:"
shasum -a 256 rtail.gz