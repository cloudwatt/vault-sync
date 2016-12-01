FROM alpine:3.3
MAINTAINER Rohith <gambol99@gmail.com>

ADD bin/vaultctl /usr/bin/vaultctl

CMD [ "/usr/bin/vaultctl" ]