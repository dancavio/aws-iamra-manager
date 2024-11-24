FROM alpine:3.20

ARG WORKDIR="/iamram"
WORKDIR $WORKDIR

RUN apk update && apk add --no-cache wget bash jq gcompat aws-cli py3-pip && \
    pip3 install --break-system-packages aws-shell

ARG TARGETPLATFORM

ADD install-aws-tools /tmp/
RUN /tmp/install-aws-tools "$TARGETPLATFORM" && rm /tmp/install-aws-tools
ADD update-credentials update-credentials

ENTRYPOINT ["sleep", "infinity"]
ENV PATH="$WORKDIR:$PATH"
