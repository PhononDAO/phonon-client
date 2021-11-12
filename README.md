# Phonon-client Alpha
phonon-client is a library, basic user interface, and set of utilities for interacting with phonon cards and the phonon network as a whole.
# Usage
For fa
### Release Conventions
See https://github.com/marketplace/actions/autotag

All commits to master will automatically create a tag with a bumped patch version, and upload the generated binaries under a release corresponding to the created tag.

Commit with #minor or #major to bump minor and major version numbers rather than the patch number. Commits that reference an issue with the "enhancement" label will also automatically bump the minor version.
# Building
## Requirements:
- a go compiler
- a c compiler for whatever architecture you are compiling for
## Process
Run make build
note: for windows compilation from mac or linux, you will need to install mingw-64 from your favorite package manager
