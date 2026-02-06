#! /bin/bash
export CSI_PROW_E2E_FOCUS_1_34="\[FeatureGate:VolumeAttributesClass\]"

custom_post_install() {
    echo "!!!! Running custom post install hook... !!!!!!!!!!"
    yq -i '.VolumeAttributesClass = {"FromExistingClassName": "testvac"}' "${CSI_PROW_WORK}/test-driver.yaml"
    cat "${CSI_PROW_WORK}/test-driver.yaml"

    # Create the VAC Class
    cat >"${CSI_PROW_WORK}/sample-vac.yaml" <<EOF
apiVersion: storage.k8s.io/v1
kind: VolumeAttributesClass
metadata:
  name: testvac
driverName: hostpath.csi.k8s.io
parameters:
  e2eVacTest: e2eVacTestValue
EOF
    info "!!!!!!!!!! Creating sample VAC class !!!!!!!!!! "
    cat "${CSI_PROW_WORK}/sample-vac.yaml"
    kubectl create -f "${CSI_PROW_WORK}/sample-vac.yaml"
    info "########## Modifying CSI driver to enable VAC ##########"
    kubectl get statefulsets csi-hostpathplugin  -o json|jq '(.spec.template.spec.containers[] | select(.name == "hostpath").args) += ["-enable-controller-modify-volume", "--accepted-mutable-parameter-names=e2eVacTest"]'|kubectl apply -f -

    info "!!!!!!!!!! Rolling out statefulset !!!!!!!!!!"
    kubectl rollout restart sts csi-hostpathplugin -n default
    # Wait for the rollout to complete
    info "########## Waiting for rollout to complete..."
    kubectl rollout status sts csi-hostpathplugin -n default
}

export CSI_PROW_DRIVER_POSTINSTALL="custom_post_install"
. release-tools/prow.sh
main

