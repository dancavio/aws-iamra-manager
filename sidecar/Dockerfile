FROM alpine:3.20

ARG WORKDIR="/iamram"
WORKDIR $WORKDIR

RUN apk add --no-cache wget bash jq gcompat aws-cli

ARG TARGETPLATFORM

ADD install-aws-tools /tmp/
RUN /tmp/install-aws-tools "$TARGETPLATFORM" && rm /tmp/install-aws-tools
ADD update-credentials update-credentials

ENTRYPOINT ["sleep", "infinity"]

ENV PATH="$WORKDIR:$PATH"

ENV AWS_PAGER=""
ENV AWS_EC2_METADATA_SERVICE_ENDPOINT="http://127.0.0.1:9911/"
