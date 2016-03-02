[![logvac logo](http://nano-assets.gopagoda.io/readme-headers/logvac.png)](http://nanobox.io/open-source#logvac)  
[![Build Status](https://travis-ci.org/nanopack/logvac.svg)](https://travis-ci.org/nanopack/logvac)

# Logvac

Simple, lightweight, api-driven log aggregation service with realtime push capabilities and historical persistence.

## Usage
Logvac can receive logs from rsyslog
>/etc/rsyslog.d/01-logvac-example.conf
>```
# rsyslog.conf style - more info look at rsyslog.conf(5)
# Single '@' sends to UDP
*.* @127.0.0.1:1234
# Double '@' sends to TCP
*.* @@127.0.0.1:1235
```
> `sudo service rsyslog restart` with the preceding config file should start dumping logs to logvac

See http examples [here](../api/README.md)  

### Contributing

Contributions to the logvac project are welcome and encouraged. Logvac is a [Nanobox](https://nanobox.io) project and contributions should follow the [Nanobox Contribution Process & Guidelines](https://docs.nanobox.io/contributing/).

### Licence

Mozilla Public License Version 2.0

[![open source](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
