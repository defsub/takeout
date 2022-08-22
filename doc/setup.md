# Takeout Setup

## Overview

This setup assumes the use of a UNIX system such as Linux. Takeout is written
in Go, so most systems will work just fine. Please adjust commands below as
needed. You can setup on a virtual system in the cloud such as EC2, GCE,
Digital Ocean, Linode, or use your spiffy computer at home.

You need to have media stored in an S3 bucket somewhere. Some common services
are AWS S3, Wasabi, Backblaze, and Minio. And if you're using that spiffy home
computer, you can also install Minio at home and make your local media
available via S3 to your home network. The other bucket services cost money,
however you have the added benefit of having your media securely available
wherever you go.

Please see [bucket.md](bucket.md) for further details on how you should
organize your media in S3. [rclone](https://rclone.org) is an excellent tool to
manage S3 buckets from the command line. Once that's all done, proceed with the
steps below.

One recommended cloud setup would be:
* Linode (Nanode 1GB $5/mo) for running Takeout
* Wasabi ($5.99 TB/mo) for S3 media

## Steps

Download and install Go from [https://go.dev/](https://go.dev/) as needed. Run
the following command to ensure Go is correctly installed. You should have Go
1.18 or higher.

```console
$ go version
```

Download and build the Takeout server from Github. Precompiled versions may be
available at a later time. Check the [Takeout Releases Page](https://github.com/defsub/takeout/releases).

```console
$ git clone https://github.com/defsub/takeout.git
$ cd takeout
$ go build
```

Install Takeout in ${GOPATH}/bin. Don't worry if you don't have a GOPATH
environment variable defined, Go will default to your home directory
(~/go/bin). Ensure that ${GOPATH}/bin is in your command path. You should see a
Takeout version displayed.

```console
$ go install
$ takeout version
```

Create the takeout directory. This is the base directory where config files,
databases, and logs will be stored.

```console
$ TAKEOUT_HOME=~/takeout
$ mkdir ${TAKEOUT_HOME}
```

Create a media sub-directory to store bucket specific configuration and
databases.  Change "mymedia" to whatever name you like (here and below). Your
media isn't stored here, just related config files and databases.

```console
$ mkdir ${TAKEOUT_HOME}/mymedia
```

Copy sample start script

```console
$ cp start.sh ${TAKEOUT_HOME}
$ chmod 755 ${TAKEOUT_HOME}/start.sh
```

Copy sample config files

```console
$ cp doc/takeout.yaml ${TAKEOUT_HOME}
$ cp doc/config.yaml ${TAKEOUT_HOME}/mymedia
```
Sync your media. This may take multiple hours depending on the amount of media
files. Repeat the sync command for other media directories you may have
created.

```console
$ cd ${TAKEOUT_HOME}/mymedia
$ ~/go/bin/takeout sync
```

You may encounter sync errors like the following:

    2022/08/11 11:08:41 artist not found: Billy F Gibbons
	2022/08/11 11:08:41 track release not found: Billy F Gibbons/Hardware/12/1

In this example the artist "Billy F Gibbons" was used to tag and store the
media, however, MusicBrainz knows this artist as "Billy Gibbons". You can fix
this with RewriteRules or adjust your media file names and/or directory names.

Troubleshooting this stuff may be difficult so ask for help as needed.

Create your first user. Change the example user "ozzy" and password. Please use
a strong password to protect access to your media. Note that "mymedia" must
match the takeout sub-directory name used above. The idea here is that there
can be multiple users and users can use the same or different buckets of
media. Indie for Dad, scary movies for Mom, and some emo for the teenager.

```console
$ cd ${TAKEOUT_HOME}
$ ~/go/bin/takeout user --add --user="ozzy" --pass="changeme" --media="mymedia"
```

Setup a secure TLS front-end to the Takeout server. [Nginx](http://nginx.org/)
is a great option. [Let's Encrypt](https://letsencrypt.org/)) can be used to
get a free TLS certificate. An Nginx config example would be:

    server {
        listen 0.0.0.0:80;
        server_name myserver.org;
        return 301 https://$host$request_uri;
    }

    upstream myapp {
        server 127.0.0.1:3000;
        keepalive 8;
    }

    server {
        listen 443 ssl http2;
        server_name myserver.org;

        ssl_certificate /etc/letsencrypt/live/myserver.org/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/myserver.org/privkey.pem;
        ssl_trusted_certificate /etc/letsencrypt/live/myserver.org/chain.pem;

        access_log /var/log/nginx/myserver.org.log;

        location / {
          proxy_set_header X-Real-IP $remote_addr;
          proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
          proxy_set_header Host $http_host;
          proxy_set_header X-NginX-Proxy true;
          proxy_pass http://myapp/;
          proxy_redirect off;
        }
    }


Start the server.

```console
$ cd ${TAKEOUT_HOME}
$ ./start.sh
```
