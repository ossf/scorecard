#!/bin/bash

SOURCE="${BASH_SOURCE[0]}"
input=$(dirname $SOURCE)/projects.txt
output=$(date +"%m-%d-%Y").csv
touch $output
while read -r line
do
    echo $line
    go run . --repo=$line --format=csv 2>/dev/null >> $output
done < "$input"
