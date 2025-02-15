FROM dpokidov/imagemagick:7.1.1-41-bookworm AS build

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

ENV GOLANG_VERSION 1.22.10

RUN set -eux; \
	now="$(date '+%s')"; \
	arch="$(dpkg --print-architecture)"; arch="${arch##*-}"; \
	url=; \
	case "$arch" in \
		'amd64') \
			url='https://dl.google.com/go/go1.22.10.linux-amd64.tar.gz'; \
			sha256='736ce492a19d756a92719a6121226087ccd91b652ed5caec40ad6dbfb2252092'; \
			;; \
		'armhf') \
			url='https://dl.google.com/go/go1.22.10.linux-armv6l.tar.gz'; \
			sha256='a7bbbc80fe736269820bbdf3555e91ada5d18a5cde2276aac3b559bc1d52fc70'; \
			;; \
		'arm64') \
			url='https://dl.google.com/go/go1.22.10.linux-arm64.tar.gz'; \
			sha256='5213c5e32fde3bd7da65516467b7ffbfe40d2bb5a5f58105e387eef450583eec'; \
			;; \
		'i386') \
			url='https://dl.google.com/go/go1.22.10.linux-386.tar.gz'; \
			sha256='2ae9f00e9621489b75494fa2b8abfc5d09e0cae6effdd4c13867957ad2e4deba'; \
			;; \
		'mips64el') \
			url='https://dl.google.com/go/go1.22.10.linux-mips64le.tar.gz'; \
			sha256='e66c440c03dd19bf8423034cbde7f6813321beb18d3fcf2ef234c13a25467952'; \
			;; \
		'ppc64el') \
			url='https://dl.google.com/go/go1.22.10.linux-ppc64le.tar.gz'; \
			sha256='db05b9838f69d741fb9a5301220b1a62014aee025b0baf341aba3d280087b981'; \
			;; \
		'riscv64') \
			url='https://dl.google.com/go/go1.22.10.linux-riscv64.tar.gz'; \
			sha256='aef9b186c1b9b58c0472dbf54978f97682852a91b2e8d6bf354e59ba9c24438a'; \
			;; \
		's390x') \
			url='https://dl.google.com/go/go1.22.10.linux-s390x.tar.gz'; \
			sha256='4ab2286adb096576771801b5099760b1d625fd7b44080449151a4d9b21303672'; \
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
	go version;

ENV GOTOOLCHAIN=local

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

FROM dpokidov/imagemagick:7.1.1-41-bookworm

ENV IM_HOME /usr/local/bin

USER 65534
COPY --from=build --chown=nobody:nogroup /illustration /usr/local/bin/illustration
COPY --from=build --chown=nobody:nogroup /transformimgs /transformimgs

ENTRYPOINT ["/transformimgs", "-imConvert=/usr/local/bin/convert", "-imIdentify=/usr/local/bin/identify"]
