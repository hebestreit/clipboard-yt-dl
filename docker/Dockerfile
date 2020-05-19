FROM archlinux/base
# Add mingw binary distribution repository
RUN printf  "[ownstuff]\n\
SigLevel = Optional TrustAll\n\
Server = https://martchus.no-ip.biz/repo/arch/\$repo/os/\$arch" >> /etc/pacman.conf

# install dependencies
RUN pacman -Sy --noconfirm \
  base-devel \
  ppl \
  zlib \
  libmpc \
  gcc-ada \
  gmp \
  libxml2 \
  webkit2gtk \
  patch \
  clang \
  llvm \
  mingw-w64-gcc \
  gtk3 \
  cmake \
  go \
  libappindicator-gtk3

WORKDIR /tmp

# setup go environment
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin"

WORKDIR /go/src

ADD entrypoint.sh /

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]