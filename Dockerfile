FROM dpokidov/imagemagick:7.1.1-36-2-bookworm AS build

ARG BRANCH=main

RUN apt-get update && apt-get upgrade -y && apt-get install -y --no-install-recommends \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config \
		dirmngr \
		gpg \
		gpg-agent \
		wget \
		git \
	&& rm -rf /var/lib/apt/lists/*

#Installing golang
ENV PATH /usr/local/go/bin:$PATH

ENV GOLANG_VERSION 1.22.6

RUN set -eux; \
	now="$(date '+%s')"; \
	arch="$(dpkg --print-architecture)"; arch="${arch##*-}"; \
	url=; \
	case "$arch" in \
		'amd64') \
			url='https://dl.google.com/go/go1.22.6.linux-amd64.tar.gz'; \
			sha256='999805bed7d9039ec3da1a53bfbcafc13e367da52aa823cb60b68ba22d44c616'; \
			;; \
		'armhf') \
			url='https://dl.google.com/go/go1.22.6.linux-armv6l.tar.gz'; \
			sha256='b566484fe89a54c525dd1a4cbfec903c1f6e8f0b7b3dbaf94c79bc9145391083'; \
			;; \
		'arm64') \
			url='https://dl.google.com/go/go1.22.6.linux-arm64.tar.gz'; \
			sha256='c15fa895341b8eaf7f219fada25c36a610eb042985dc1a912410c1c90098eaf2'; \
			;; \
		'i386') \
			url='https://dl.google.com/go/go1.22.6.linux-386.tar.gz'; \
			sha256='9e680027b058beab10ce5938607660964b6d2c564bf50bdb01aa090dc5beda98'; \
			;; \
		'mips64el') \
			url='https://dl.google.com/go/go1.22.6.linux-mips64le.tar.gz'; \
			sha256='01547606c5b5c1b0e5587b3afd65172860d2f4755e523785832905759ecce2d7'; \
			;; \
		'ppc64el') \
			url='https://dl.google.com/go/go1.22.6.linux-ppc64le.tar.gz'; \
			sha256='9d99fce3f6f72a76630fe91ec0884dfe3db828def4713368424900fa98bb2bd6'; \
			;; \
		'riscv64') \
			url='https://dl.google.com/go/go1.22.6.linux-riscv64.tar.gz'; \
			sha256='30be9c9b9cc4f044d4da9a33ee601ab7b3aff4246107d323a79e08888710754e'; \
			;; \
		's390x') \
			url='https://dl.google.com/go/go1.22.6.linux-s390x.tar.gz'; \
			sha256='82f3bae3ddb4ede45b848db48c5486fadb58551e74507bda45484257e7194a95'; \
			;; \
		*) echo >&2 "error: unsupported architecture '$arch' (likely packaging update needed)"; exit 1 ;; \
	esac; \
	\
	wget -O go.tgz.asc "$url.asc"; \
	wget -O go.tgz "$url" --progress=dot:giga; \
	echo "$sha256 *go.tgz" | sha256sum -c -; \
	\
# https://github.com/golang/go/issues/14739#issuecomment-324767697
	GNUPGHOME="$(mktemp -d)"; export GNUPGHOME; \
# https://www.google.com/linuxrepositories/
	gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC  EC91 7721 F63B D38B 4796'; \
# let's also fetch the specific subkey of that key explicitly that we expect "go.tgz.asc" to be signed by, just to make sure we definitely have it
	gpg --batch --keyserver keyserver.ubuntu.com --recv-keys '2F52 8D36 D67B 69ED F998  D857 78BD 6547 3CB3 BD13'; \
	gpg --batch --verify go.tgz.asc go.tgz; \
	gpgconf --kill all; \
	rm -rf "$GNUPGHOME" go.tgz.asc; \
	\
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
# save the timestamp from the tarball so we can restore it for reproducibility, if necessary (see below)
	SOURCE_DATE_EPOCH="$(stat -c '%Y' /usr/local/go)"; \
	export SOURCE_DATE_EPOCH; \
	touchy="$(date -d "@$SOURCE_DATE_EPOCH" '+%Y%m%d%H%M.%S')"; \
# for logging validation/edification
	date --date "@$SOURCE_DATE_EPOCH" --rfc-2822; \
# sanity check (detected value should be older than our wall clock)
	[ "$SOURCE_DATE_EPOCH" -lt "$now" ]; \
	\
	if [ "$arch" = 'armhf' ]; then \
		[ -s /usr/local/go/go.env ]; \
		before="$(go env GOARM)"; [ "$before" != '7' ]; \
		{ \
			echo; \
			echo '# https://github.com/docker-library/golang/issues/494'; \
			echo 'GOARM=7'; \
		} >> /usr/local/go/go.env; \
		after="$(go env GOARM)"; [ "$after" = '7' ]; \
# (re-)clamp timestamp for reproducibility (allows "COPY --link" to be more clever/useful)
		touch -t "$touchy" /usr/local/go/go.env /usr/local/go; \
	fi; \
	\
# ideally at this point, we would just "COPY --link ... /usr/local/go/ /usr/local/go/" but BuildKit insists on creating the parent directories (perhaps related to https://github.com/opencontainers/image-spec/pull/970), and does so with unreproducible timestamps, so we instead create a whole new "directory tree" that we can "COPY --link" to accomplish what we want
	mkdir /target /target/usr /target/usr/local; \
	mv -vT /usr/local/go /target/usr/local/go; \
	ln -svfT /target/usr/local/go /usr/local/go; \
	touch -t "$touchy" /target/usr/local /target/usr /target; \
	\
# smoke test
	go version; \
# make sure our reproducibile timestamp is probably still correct (best-effort inline reproducibility test)
	epoch="$(stat -c '%Y' /usr/local/go)"; \
	[ "$SOURCE_DATE_EPOCH" = "$epoch" ]

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 1777 "$GOPATH"
WORKDIR $GOPATH

RUN mkdir -p /go/src/github.com/Pixboost/
WORKDIR /go/src/github.com/Pixboost/

RUN git clone --branch $BRANCH --single-branch https://github.com/Pixboost/transformimgs.git

WORKDIR /go/src/github.com/Pixboost/transformimgs/illustration

RUN go build -o /illustration

WORKDIR /go/src/github.com/Pixboost/transformimgs/cmd

RUN go build -o /transformimgs

FROM dpokidov/imagemagick:7.1.1-36-2-bookworm

ENV IM_HOME /usr/local/bin

USER 65534
COPY --from=build --chown=nobody:nogroup /illustration /usr/local/bin/illustration
COPY --from=build --chown=nobody:nogroup /transformimgs /transformimgs

ENTRYPOINT ["/transformimgs", "-imConvert=/usr/local/bin/convert", "-imIdentify=/usr/local/bin/identify"]
