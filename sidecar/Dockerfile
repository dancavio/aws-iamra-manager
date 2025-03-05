FROM alpine:3.20

ARG WORKDIR="/iamram"
WORKDIR $WORKDIR

RUN apk add --no-cache wget curl bash jq gcompat aws-cli

ARG TARGETPLATFORM

ADD install-aws-tools /tmp/
RUN /tmp/install-aws-tools "$TARGETPLATFORM" && rm /tmp/install-aws-tools
ADD scripts/* .

ARG release_version="DEV"
RUN echo "$release_version" > version

ENTRYPOINT ["sleep", "infinity"]

ENV PATH="$WORKDIR:$PATH"

ENV AWS_PAGER=""
ENV AWS_EC2_METADATA_SERVICE_ENDPOINT="http://127.0.0.1:9911/"
