FROM ubuntu:jammy
RUN apt-get -y update
# Use with these env vars otherwise tzdata gets stuck selecting area
RUN DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get install -y git make curl sqlite3 php php-sqlite3
WORKDIR /letswifi-portal
RUN git clone https://github.com/geteduroam/letswifi-portal .
# Pin to specific commit
RUN git checkout 6b7043d5a6aae05e0c06e4e706a5365eff9daf73
RUN make check-php
RUN make vendor
RUN make etc/letswifi.conf.php
RUN mkdir -p var
RUN sqlite3 var/letswifi-dev.sqlite < sql/letswifi.sqlite.sql
RUN php bin/add-realm.php example.com 1
CMD php -S 0.0.0.0:8080 -t www
