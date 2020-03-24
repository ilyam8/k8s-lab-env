#!/usr/bin/env bash

install() {
  dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" > /dev/null 2>&1 && pwd)"

  kubectl apply -f "$dir"/env/namespace.yml
  kubectl apply -f "$dir"/env/
}

install
