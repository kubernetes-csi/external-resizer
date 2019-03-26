# v0.1.0

## Initial release Changelog

* Add a external resize controller which monitors Persistent volume claims and performs
  CSI `ControllerExpandVolume` as needed. If plugin does not implement `ControllerExpandVolume` it performs a no-op expansion
  and updates PV object.

