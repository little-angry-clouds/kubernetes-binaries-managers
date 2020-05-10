# helmenv
[Helm](https://helm.sh/) version manager inspired by
[tfenv](https://github.com/tfutils/tfenv/).

If you come from the helmenv bash version, you should read the [FAQ](#how-to-migrate-from-the-bash-version)!

## Features

- Install helm versions in a reproducible and easy way
- Enforce version in your git repositories with a `.helm_version` file

## Supported OS

Currently helmenv supports the following OSes
- Mac OS
- Linux
- Windows

## Installation

1. Add `~/.bin` to your `$PATH` and create it if doesn't exist

```bash
echo 'export PATH="$HOME/.bin:$PATH"' >> ~/.bashrc
# Or
echo 'export PATH="$HOME/.bin:$PATH"' >> ~/.zshrc

mkdir -p ~/.bin
```

For Windows you should do the weird stuff that it needs to to set an environmental variable.

2. Download the binary and put it on your path

Go to [the releases
page](https://github.com/little-angry-clouds/kubernetes-binaries-managers/releases)
and download the version you want. For example:

```bash
wget https://github.com/little-angry-clouds/kubernetes-binaries-managers/releases/download/0.0.2/helmenv-linux-amd64.tar.gz
tar xzf helmenv-linux-amd64.tar.gz
mv helmenv-linux-amd64 ~/.bin/helmenv
```

And that's it!

## Usage
### Help

``` bash
$ helmenv help
Kubectl version manager

Usage:
  helmenv [command]

Available Commands:
  help        Help about any command
  install     Install helm binary
  list        Lists local and remote versions
  uninstall   Uninstall helm binary
  use         Set the default version to use

Flags:
  -h, --help     help for helmenv

Use "helmenv [command] --help" for more information about a command.
```

### List installable versions

```bash
$ helmenv list remote
1.18.2
1.18.1
1.18.0
1.17.5
...
```

### List installed versions

```bash
$ helmenv list local
1.18.0
1.16.4
1.14.8
```

### Install version

```bash
$ helmenv install 1.16.5
Downloading binary...
Done! Saving it at /home/user/.bin/kubectl-v1.16.5
```

### Use version

```bash
$ helmenv use 1.16.5
Done! Using 1.16.5 version.
```

### Uninstall version

```bash
$ helmenv uninstall 1.16.5
Done! 1.16.5 version uninstalled from /home/ap/.bin/kubectl-v1.16.5.
```

## FAQ
### Why migrate from bash to go?
The project just as a way of downloading the binary versions. Progressively it
began to grow a little. And then they came some PR for different stuff, but the
hard ones where the ones for adding better support for MacOS. I don't own a Mac,
so I couldn't test them properly.

Also, `kbenv` and `helmenv` where pretty much a copy paste, but they didn't have
the same code, so any change from one place I would have to add it to the other.

So, with this to problems (and because I was bored) I decided to migrate them
for Go. Go is cool because it lets you have self contained binaries, so no more
worries about the OS! I even add support for Windows, because why not. I only
had to do a little specific development for Windows be able to use the
`.helmenv_version` file, but it was'nt traumatic. And also, being Go a real
programming language, I could add tests. Not that there's any right now, but I'm
on it.

### How to migrate from the bash version
For doing so you have to:
- Delete the `helmenv` repository: `rm -r ~/.helmenv`
- Delete the line that sources the bash script: `source $HOME/.helmenv/helmenv.sh`

And that's it. The way how the Go version works is very similar. The changed
beehaviours are:

- You don't have to set the `v` before the versions. For example:

``` bash
$ helmenv install v2.0.1
# Would be
$ helmenv install 2.0.1
```

- The listing commands have been separed:

``` bash
# Before
$ helmenv list
$ helmenv list-remote
# After
$ helmenv list local
$ helmenv list remote
```

## How to enforce a helm version
Just create a `.helm_version` in your directory pointing to the version you want
to use. For example:

``` bash
$ helmenv install 2.14.3
...
$ helmenv install 2.14.2
...
$ helmenv use 2.14.3
...
$ helm version --client
Client: &version.Version{SemVer:"v2.14.3", GitCommit:"0e7f3b6637f7af8fcfddb3d2941fcc7cbebb0085", GitTreeState:"clean"}
$ echo 2.14.2 > .helm_version
$ helm version --client
Client: &version.Version{SemVer:"v2.14.2", GitCommit:"a8b13cc5ab6a7dbef0a58f5061bcc7c0c61598e7", GitTreeState:"clean"}
```

## License
GPL3
