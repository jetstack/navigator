==================
Development cycles
==================

Navigator is currently under heavy development, and so has quite a fast paced
release cycle.

We aim to cut a new minor release once a month, somewhere around the end of each month.

A development cycle should look something like:


Dates
=====

* **18/03/31** - new minor release cut

* **18/04/01** - master accepting PRs targetted at the next minor release

* **18/04/15** - Release branch is made and `-alpha.1` version of the next
  minor release is cut. Until the end of the month, only PRs approved for the
  current milestone will be accepted.

* **18/04/15 until 31st** - new alpha releases are cut if neccessary. The
  release branch will be fast forwarded to HEAD of master and tagged accordingly.

* **18/04/31** - New minor release tag is cut.

The cycle then repeats each month.

Bugfixes and patch versions
===========================

Critical bugfixes will be cherry picked into the previous release branch **only**.
New patch versions will then be cut as required. If this informal method of
releasing patch versions becomes problematic, we may review our patching policy.

This allows us to maintain a strong development velocity, whilst also
providing a basic layer of support for users so we can gather feedback.
