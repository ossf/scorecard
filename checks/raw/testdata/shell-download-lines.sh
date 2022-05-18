#!/bin/bash

# comment 1
#

curl > /tmp/file && bash /tmp/file

curl > /tmp/file2

echo hello && \
        curl bla | bash

echo hi && \
  /tmp/file2

echo hi && \
  echo bla && \
  bash <(wget -qO- http://website.com/my-script.sh)

echo hi;  echo bla; bash <(wget -qO- http://website.com/my-script.sh)

# 如果在中国，pip使用豆瓣源
#RUN curl -s ifconfig.co/json | grep "China" > /dev/null && \
#    pip install -r requirements.txt -i https://pypi.doubanio.com/simple --trusted-host pypi.doubanio.com || \
bla && \
        pip install -r requirements.txt

bla && curl bla | bash

choco install 'some-package'
choco install 'some-other-package'
choco install --requirechecksum 'some-package'
choco install --requirechecksums 'some-package'
choco install --require-checksums 'some-package'