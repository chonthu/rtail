#Rtail
Remote tail logging

## Use it just as you would regular tail locally

	rtail -f ec2-user@webserver1.myserver.com:/var/log/httpd/error_log

## Follow tag works
rtail will stream the log down locally from the remote client

## Works with multiple servers!
Pass as many server as you would like

	rtail -f ec2-user@webserver1.myserver.com:/var/log/httpd/error_log johndoe@webserver2.myserver.com:/var/log/httpd/access_log

## Allows custom range selector

	rtail -f root@webserver[1-5].myserver.com:/var/log/httpd/error_log
