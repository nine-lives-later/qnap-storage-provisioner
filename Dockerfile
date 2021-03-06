FROM golang:1.14 AS builder

ENV CGO_ENABLED=0

COPY . /build

WORKDIR /build

RUN go vet
RUN go build -a -ldflags '-extldflags "-static"'




FROM scratch

ENV QNAP_MOUNTOPTIONS "hard:fg:suid:nfsvers=3:proto=udp:intr:rsize=8192:wsize=8192"

COPY --from=builder /build/qnap-storage-provisioner /qnap-storage-provisioner

CMD ["/qnap-storage-provisioner"]
