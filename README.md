Rtail [![Travis](https://travis-ci.org/chonthu/rtail.svg?branch=master)](https://travis-ci.org/chonthu/rtail)
===============================
Remote multiplex server tailing & command execution runner

## Install

### OSX Install with Homebrew

	brew tap chonthu/tap
	brew install rtail

### From source

	go get github.com/chonthu/rtail
	go install

## Usage

### Use it just as you would regular tail locally
rtail will stream the log down locally from the remote client, by default the command your running is tail -f

	rtail ec2-user@webserver1.myserver.com:/var/log/httpd/error_log

You can run any command you want by using a '%' istead of ':' in the rtail parameter

	rtail "ec2-user@webserver1.myserver.com%cat /var/log/mail.log"

### Works with multiple servers!
Pass as many server as you would like

	rtail ec2-user@webserver1.myserver.com:/var/log/httpd/error_log johndoe@webserver2.myserver.com:/var/log/httpd/access_log

### Allows custom range selector

	rtail "root@webserver[1-5].myserver.com:/var/log/httpd/error_log"

### Custom aliases & commands
create a rtail.yml file in your working directory or in your home folder

or use the built in shortcut `rtail init`

	---
	aliases:
		error_log: /var/log/httpd/error_log

	commands:
		varnish: varnishlog

Now you can run these custom command like so:

	rtail myserver.web1:error_log
	rtail myserver.web2%varnish