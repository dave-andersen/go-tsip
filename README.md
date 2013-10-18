go-tsip
=======

A very skeletal program to talk Trimble GPS TSIP packets, with a focus
on Trimble's Thunderbolt timing GPS receiver.
It currently supports Software Version, Primary, and Secondary timing
packets.  The latter are sent each second by the Thunderbolt.
It has a rudimentary command-sending framework that currently
knows only how to request the software version.

I usually run it as 'go run tsip.go'

Notes:
  - It assumes you have some kind of telnet server set up to access
your GPS.  You can hack this up with a serial port using socat:

   http://www.dest-unreach.org/socat/doc/socat-ttyovertcp.txt

Or you can change the program to open the device file (with appropriate
shelling out to set the serial speed, etc, using stty).  I can probably
be convinced to add this if someone wants it - drop me an email.

  - The address for that telnet server is currently hardcoded.
That should convey the degree to which this was a quick hack to
be able to determine if my GPS timing receiver was properly
synchronized.
