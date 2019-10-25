## Support Status

Alpha

## v0.1.1

* [#61](https://github.com/kubernetes-csi/external-resizer/pull/61) Verify claimref associated with PV before doing volume expansion.

## v0.1.0

* [#1](https://github.com/kubernetes-csi/external-resizer/pull/1) Add a external resize controller which monitors Persistent volume claims and performs CSI `ControllerExpandVolume` as needed.
* [#26](https://github.com/kubernetes-csi/external-resizer/pull/26) If plugin does not implement `ControllerExpandVolume` it performs a no-op expansion and updates PV object.
* [#31](https://github.com/kubernetes-csi/external-resizer/pull/31) Use lease based leader election.
