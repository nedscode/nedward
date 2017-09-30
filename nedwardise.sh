#!/bin/bash

perl -pi -e 's/\byext/nedscode/g; s/\bEdward/Nedward/g; s/\bedward/nedward/g' *.go */*.go
mv edward nedward
