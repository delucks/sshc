# sshc

[![Build Status](https://travis-ci.org/delucks/sshc.svg?branch=master)](https://travis-ci.org/delucks/sshc)

sshc is a tool for interaction with SSH config files. It retrieves host definitions from your configuration easily and allows you to copy those definitions to the `~/.ssh/config` files on remote hosts.

## Installation

`sshc` is distributed as a statically compiled binary for a number of target platforms. You can find compressed binaries for `(arm, 386, amd64) x (openbsd, netbsd, freebsd, linux) + osx-amd64` over at [the releases page](https://github.com/delucks/sshc/releases).

If you have the golang toolchain installed, you can install with `go get`:

```
$ go get github.com/delucks/sshc
```

## Usage

You can find the names of all the host definitions in your file by using `sshc hosts`:

```
$ sshc hosts
myremotehost
*
boxinthecloud
```

To return the definition for a single host from your config, use the `get` command. The single argument to this command is the name of a host defined in your config file:

```
$ sshc get myremotehost
Host myremotehost
  HostName 127.0.0.1
  Port 22
  User delucks
  IdentityFile %d/.ssh/rsa_key
```

You can also optionally convert the host definition to JSON, with the following result:

```
$ sshc get -j myremotehost
{
  "HostName" "127.0.0.1",
  "Port" "22",
  "User" "delucks",
  "IdentityFile" "%d/.ssh/rsa_key"
}
```

This is useful for pulling out specific pieces of information using a json query tool like `jq`. For example, `sshc get -j myremotehost | jq .HostName` is a quick way of returning just the hostname field of an entry in your config file. I use this for constructing URLs on remote servers like `curl -s "http://$(sshc get -j myremotehost | jq .HostName):8080/path/index.html"`.

For a more precise look through your configuration, you can search through all lines with a regular expression using `sshc find`. Why is this better than using `egrep`? `sshc` will print the host definition that match your regex cleanly, without use of the `-C` context flags. If you use ANSI color (with `--color`), the matching parts of the host definitions will get *fancy terminal colors*.

```
$ sshc find rsa
Host myremotehost
  HostName 127.0.0.1
  Port 22
  User delucks
  IdentityFile %d/.ssh/rsa_key

Host *
  IdentityFile %d/.ssh/the_other_rsa_key

```

`sshc edit` is self-explanatory.

`sshc copy` allows you to copy a host definition across a SSH tunnel onto a remote hosts' SSH config. If you haven't encountered this situation before it sounds strange; let me illustrate why it's useful. Let's say you have two workstations at work: one laptop and one desktop. Both have access to the same staging & production environments via SSH. When you get access to a new system in a different environment, you will set up your SSH session to the new system on either the laptop or desktop first. If you want access to it from the other workstation, isolating the config block in `~/.ssh/config` and copying it to the other workstation is a manual process that gets really annoying when you do it a few times.

```
$ sshc copy boxinthecloud user@myremotehost:9999
... ssh session follows ...
```

Note that the second argument to this command is a full SSH connection string; you can override users, ports, etc here because it's used in the invocation of `ssh`.

## Prior Art

### ssh-config

Code: https://github.com/dbrady/ssh-config

`ssh-config` is a command-line only Ruby app which retrieves and modifies host definitions from your configuration. It's got a well-documented command-line interface that outputs to stdout and makes intuitive sense. It's also actively maintained from the commit log as of July 2018.

### stormssh

Code: https://github.com/emre/storm

stormssh is a similar management tool that has a ton of functionality including a web UI, tray indicator, and full CRUD on every entry in the file- allowing you to modify individual fields from the command-line. It's really powerful and can accomplish some truly fancy things with your configuration.

### advanced-ssh-config

Code: https://github.com/moul/advanced-ssh-config

This is a YAML file syntax and mini-DSL for implementing buildable ssh config files. It focuses on the creation of those files and not querying them after the fact.

## How `sshc` Differs

From **stormssh**: I considered using stormssh for my use-case but was put off by the large dependencies (Flask, paramiko) and the abundance of features I don't consider necessary. I wanted a more focused and command-line oriented tool that doesn't re-implement simple things for you like `storm delete all` (could be done with `rm ~/.ssh/config`). `stormssh` also appears to be unmaintained.

From **ssh-config**: I like this tool a lot and started off using it, but the language choice eventually drove me away from it. For a non-Ruby developer, the whole runtime is a big dependency to install on every machine for this tool. I also don't feel comfortable adding more features like the copy-definition feature in `sshc`.

`sshc` is also novel in a couple ways:
- Copying definitions from the local host to a remote `~/.ssh/config`
- Choice of golang; small footprint & installation from a single binary
