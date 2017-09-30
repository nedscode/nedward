#!/bin/bash

perl -pi -e 's/\bnedscode/yext/g; s/\bNedward/Edward/g; s/\bnedward/edward/g' *.go */*.go
mv nedward edward
