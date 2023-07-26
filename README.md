# OmogenExec
OmogenExec is a component that can be used for implementing programming problem judging systems.
It consists of two things: a sandbox component and a Go library for evaluating problems.

It is only tested on Ubuntu LTS 22.04.

## Setup
First, you must enable filesystem quotas.
The sandbox uses this functionality to prevent submissions from writing too much data to disk.

1. Install quota by running `sudo apt install quota`.
1. Open `/etc/fstab` and add `usrquota,grpquota` to the options of the filesystem containing `/var/lib/omogen` (by default mount point `/`).
1. Remount your filesystem by running `sudo mount -o remount /`.
1. Enable quota tracking by running `sudo quotacheck -ugm /`.
1. Turn on quota by running `sudo quotaon -v /`.

Next, install the sandbox from the latest `omogenexec-debian.deb` [release](https://github.com/jsannemo/omogenexec/releases/tag/v1.3.1).

You can verify that the sandbox is working by running `printf '\x01/bin/true\x00' | omogenexec --sandbox-id 1 --blocks 1024 --inodes 1024 --memory-mb 1024 --time-lim-ms 1000 --wall-time-lim-ms 1000 --pid-limit 1 2>/dev/null`.
The output should be
```
code 0
cpu <some integer>
done
```
if everything works.
