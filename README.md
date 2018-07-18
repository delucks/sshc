# sshc

sshc is a tool for interaction with SSH config files. It gives you more flexibility in querying them for host definitions with context and a few interesting features.

## Usage

The most basic usage is returning the definition for a single host from the file:

```
$ sshc get myremotehost
Host myremotehost
  HostName 127.0.0.1
  Port 22
  User delucks
  IdentityFile %d/.ssh/rsa_key
```

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

The rest of this README is scratch space left over from planning.

```
$ sshc get -j hostname
{"hostname": {
  "HostName": "127.0.0.1",
  "Port": "22",
  "User": "delucks",
  "IdentityFile": "%d/.ssh/rsa_key",
}}
```
