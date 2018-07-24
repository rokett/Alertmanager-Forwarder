# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2018-07-24
### Changed
 - #3 Amended to track errors when sending individual alerts to vRO, and return them all once all of the alerts have been iterated over.  This stops one bad alert from causing none of the subsequent alerts being processed.

## [1.0.1] - 2018-07-17
### Fixed
 - #1 --vro-port flag was not being passed to the service configuration when installing.
 - #2 Fixed bad logic whereby not all alerts were being passed to vRO.

## [1.0.0] - 2018-07-16
Initial release
