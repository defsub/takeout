# Takeout Setup

## Overview

This setup assumes the use of a UNIX system such as Linux. Takeout is written
in Go, so most systems will work just fine. Please adjust commands as
needed. You can setup on a virtual system in the cloud such as EC2, GCE,
Digital Ocean, Linode, or use your spiffy computer at home.

You need to have media stored in an S3 bucket somewhere. Some common services
are AWS S3, Wasabi, Backblaze, and Minio. And if you're using that spiffy home
computer, you can also install Minio at home and make your local media
available via S3 to your home network. The other bucket services cost money but
you have the added benefit of having your media securely available wherever you
go.

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

Create the takeout directory. There is the base directory will config files,
databases, and logs will be stored.

```console
$ TAKEOUT_HOME=~/takeout
$ mkdir ${TAKEOUT_HOME}
```

Create media sub-directory to store bucket specific configuration and
databases.  Change "mymedia" to whatever name you like (here and below).

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

Create your first user. Change the example user "ozzy" and password. Please use
a strong password to protect access to your media. Note that "mymedia" must
match a takeout sub-directory name used above. The idea here is that there can
be multiple users and users can use the same or different buckets of
media. Indie for Dad, scary movies for Mom, and some emo for the teenager.

```console
$ cd ${TAKEOUT_HOME}
$ ~/go/bin/takeout user --add --user="ozzy" --pass="changeme" --media="mymedia"
```

Start the server.

```console
$ cd ${TAKEOUT_HOME}
$ ./start.sh
```
