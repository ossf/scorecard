#!/bin/env sh -e
# Copyright 2021 Security Scorecard Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Find input files
MY_INPUT_FILE="${TEST_SRCDIR}/some/path/myinputfile.dat"
readonly MY_INPUT_FILE
MY_OUTPUT_FILE="${TEST_TMPDIR}/myoutput.txt"
readonly MY_OUTPUT_FILE

# Do something
echo hello || die "Failed in bar()"

# Check something
check_eq "${A}" "${B}"

echo "PASS"

if [ $1 -gt 100 ]
then
    echo Hey that\'s a large number.
    pwd
    echo hi && curl -s blabla | bash
fi

curl bla > myfile
./myfile

sh -c "curl bla | sh"
curl bla > file2
bash -c "file2"

sh -c "curl bla > file1"
sh -c "./file1"

bash <(wget -qO- http://website.com/my-script.sh)

wget http://file-with-sudo -O /tmp/file3
bash /tmp/file3

date