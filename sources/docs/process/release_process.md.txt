# Release Process
This document describes the mechanics of the release process for Edge Conductor

## Code Freeze

Once the team decides the code is ready for code freeze, a release branch is created
with a name in the form "edge_conductor_VERSION" where the VERSION is the 
semantic version number for the release.  

No new features should be committed to this branch.

## Test and Debug Phase

Once the release branch is created, the Test phase of the release process begins and
test cases are executed.  Any serious problems found during testing should be
fixed, tested and merged to the main branch
first and then cherry picked to the release branch.

The Test and Debug Phase ends when testing meets release quality objectives and 
the release is approved by the Test Lead, the Software Quality Lead and the
Security Lead.

## Release Phase

Once the release branch is tested and the release is approved, the final steps
to create the release are to tag the branch e.g. "ec_VERSION" 
and create the release in gitlab (or github).

Release artifacts should then be posted to the artifact repo.

## Post Release Updates

Should new fixes and changes be required for a release, they should be
cherry picked from the main branch to the release branch, tested and approved.
A new tag should be applied with the build (3rd) digit of the release
version number incremented and a new gitlab/github release should be created.

New release artifacts should then be posted to the artifact repo.

Copyright (C) 2022 Intel Corporation
 
SPDX-License-Identifier: Apache-2.0
