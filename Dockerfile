FROM alpine:3.23

ENV WORKDIR=/workdir

RUN mkdir -p ${WORKDIR}

WORKDIR ${WORKDIR}
VOLUME ${WORKDIR}

ENTRYPOINT ["/bin/gt"]

COPY gt /bin/gt
