#!/bin/sh
#
# Copyright 2016 The kuard Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script is run as root to initialize our data volume.  This includes
# creating directories and setting permissions.

set -o errexit
set -o nounset
set -o pipefail

mkdir -p /data/go

for ARCH in ${ALL_ARCH}; do
  mkdir -p "/data/std/${ARCH}"
done

mkdir -p "${npm_config_cache}"

chown -R ${TARGET_UIDGID} /data
chmod -R a=rwX /data
