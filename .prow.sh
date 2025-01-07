#! /bin/bash
# We need to hardcode e2e version for resizer for now, because
# we need fixes from latest release-1.31 branch for all e2es to pass
export CSI_PROW_E2E_VERSION="release-1.31"
. release-tools/prow.sh
main
