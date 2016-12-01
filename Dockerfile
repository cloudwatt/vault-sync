FROM alpine:3.3
MAINTAINER Félix Cantournet <felix.cantournet@gmail.com>

ADD bin/vault-sync /usr/bin/vault-sync

CMD [ "/usr/bin/vault-sync" ]