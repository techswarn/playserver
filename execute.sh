#!/bin/bash

set -euxo pipefail

go build . 
./playserver