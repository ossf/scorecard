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

pip install --no-deps --editable .
pip install --no-deps -e .
pip install --no-deps -e hg+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package
pip install --no-deps -e svn+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package
pip install --no-deps -e bzr+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package
pip install --no-deps -e git+https://github.com/username/repo.git
pip install --no-deps -e git+https://github.com/username/repo.git#egg=package
pip install --no-deps -e git+https://github.com/username/repo.git@v1.0
pip install --no-deps -e git+https://github.com/username/repo.git@v1.0#egg=package
pip install --no-deps -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567
pip install --no-deps -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package
pip install --no-deps -e git+https://github.com/username/repo@0123456789abcdef0123456789abcdef01234567#egg=package
pip install --no-deps -e git+http://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package
pip install --no-deps -e git+ssh://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package
pip install --no-deps -e git+git://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package
pip install --no-deps -e git://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package
pip install -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package
pip install --no-deps -e . git+https://github.com/username/repo.git
pip install --no-deps -e . git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package

python -m pip install --no-deps -e git+https://github.com/username/repo.git
python -m pip install --no-deps -e git+https://github.com/username/repo.git@0123456789abcdef0123456789abcdef01234567#egg=package

nuget install some-package
nuget install some-package -Version 1.2.3
nuget install packages.config
dotnet add package some-package
dotnet add package some-package -v 1.2.3
dotnet add package some-package --version 1.2.3