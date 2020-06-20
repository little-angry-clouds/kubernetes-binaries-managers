#!/bin/bash

export CREATE_CLUSTER_COOLDOWN="${BATS_CREATE_CLUSTER_COOLDOWN:-60s}"

function create_cluster() {
  # If BATS_TEST_ID not set, create it and create cluster
  # If set, don't destroy the cluster when finishing testing
  if [[ -z ${BATS_TEST_ID+x} ]]
  then
    BATS_TEST_ID=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 7 | head -n 1)
    echo "# Test cluster with ID: $BATS_TEST_ID ..." >&3
    kind create cluster --name "test-$BATS_TEST_ID" --wait "$CREATE_CLUSTER_COOLDOWN" &> /dev/null
  else
    BATS_CLEAN_TEST=false
  fi

  export KUBECONFIG=/tmp/test-$BATS_TEST_ID

  # Delete cluster if BATS_CLEAN_TEST != false
  if [[ $BATS_CLEAN_TEST != "false" ]]
  then
    echo "kind delete cluster --name test-$BATS_TEST_ID"
  fi
}
