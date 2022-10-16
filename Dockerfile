# Copyright (C) 2022 The Takeout Authors.
#
# This file is part of Takeout.
#
# Takeout is free software: you can redistribute it and/or modify it under the
# terms of the GNU Affero General Public License as published by the Free
# Software Foundation, either version 3 of the License, or (at your option)
# any later version.
#
# Takeout is distributed in the hope that it will be useful, but WITHOUT ANY
# WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
# FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for
# more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

FROM golang:1.19.2-bullseye as builder
ARG src=/go/src/github.com/defsub/takeout

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading
# them in subsequent builds if they change
WORKDIR $src/
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN make install

FROM debian:bullseye-slim
ARG src=/go/src/github.com/defsub/takeout
ARG user=takeout
ARG group=$user
ARG etc=/etc/$user/
ARG home=/var/lib/$user
ARG uid=1023
ARG gid=1023

RUN groupadd -g $gid $group && useradd -r -d $home -m -g $group -u $uid $user
COPY --from=builder $src/doc/takeout.yaml $etc
COPY --from=builder /go/bin/takeout /go/bin/takeout

# go home
WORKDIR $home
USER $user
CMD ["/go/bin/takeout", "serve", "--config", "/etc/takeout/takeout.yaml"]

# sudo docker run -name takeout -p 3000:3000 -v /home/mboyns/takeout:/var/lib/takeout takeout:latest /bin/bash
