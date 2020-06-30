![Static Tests](https://github.com/little-angry-clouds/kubernetes-binaries-managers/workflows/Generic%20tests/badge.svg) ![Int Test Linux](https://github.com/little-angry-clouds/kubernetes-binaries-managers/workflows/Int%20Test%20Linux/badge.svg) ![Int Test MacOS](https://github.com/little-angry-clouds/kubernetes-binaries-managers/workflows/Int%20Test%20MacOS/badge.svg) ![Int Test Windows](https://github.com/little-angry-clouds/kubernetes-binaries-managers/workflows/Int%20Test%20Windows/badge.svg)

# kbenv
[Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) version
manager inspired by [tfenv](https://github.com/tfutils/tfenv/).

If you are coming from the kbenv bash version, you should read the [FAQ](#how-to-migrate-from-the-bash-version)!

## Features

- Install kubectl versions in a reproducible and easy way
- Enforce version in your git repositories with a `.kubectl_version` file
- Use automatic mode and download the version that matches the detected
  kubernetes cluster. Feature shamelessly copied from
  [kuberlr](https://github.com/flavio/kuberlr).

## Supported OS

Currently kbenv supports the following OSes
- Mac OS
- Linux
- Windows

## Installation

There are two components in `kbenv`. One is the `kbenv` binary, the other one
is a `kubectl` wrapper. It works as if were `kubectl`, but it has some logic to choose
the version to execute. You should take care and ensure that you don't have any
`kubectl` binary in your path. To check which binary you're executing, you can see
it with:

``` bash
$ which kubectl
/opt/brew/bin/kubectl
```

### Homebrew

This is the recomended way, since it provides upgrades. It should work in Mac,
Linux and Windows with WSL.

``` bash
# Just the first time, activate the repository
brew tap little-angry-clouds/homebrew-my-brews
# To install
brew install kbenv
# To upgrade
brew upgrade kbenv
```

You should add your `homebrew` binary path to your PATH:

``` bash
echo 'export PATH="$(brew --prefix)/bin/:$PATH"' >> ~/.bashrc
# Or
echo 'export PATH="$(brew --prefix)/bin/:$PATH"' >> ~/.zshrc
```

For Windows you should do the weird stuff that it needs to set an environmental variable.

### Manually

1. Add `~/.bin` to your `$PATH` and create it if doesn't exist

```bash
echo 'export PATH="$HOME/.bin:$PATH"' >> ~/.bashrc
# Or
echo 'export PATH="$HOME/.bin:$PATH"' >> ~/.zshrc

mkdir -p ~/.bin
```

2. Download the binaries and put them on your path

Go to [the releases
page](https://github.com/little-angry-clouds/kubernetes-binaries-managers/releases)
and download the version you want. For example:

```bash
wget https://github.com/little-angry-clouds/kubernetes-binaries-managers/releases/download/0.0.4/kbenv-linux-amd64.tar.gz
tar xzf kbenv-linux-amd64.tar.gz
mv kbenv-linux-amd64 ~/.bin/kbenv
mv kubectl-wrapper-linux-amd64 ~/.bin/kubectl
```

And that's it!

## Usage
### Help

``` bash
$ kbenv help
Kubectl version manager

Usage:
  kbenv [command]

Available Commands:
  help        Help about any command
  install     Install kubectl binary
  list        Lists local and remote versions
  uninstall   Uninstall kubectl binary
  use         Set the default version to use

Flags:
  -h, --help     help for kbenv

Use "kbenv [command] --help" for more information about a command.
```

### List installable versions

This option uses Github API to paginate all versions. Github API has some usage
limitations. It usually works, but if you happen to do a lot of requests to
github or are on an office or similar, chances are that this command will fail.
You can still install binaries if you know the version you want, thought.

```bash
$ kbenv list remote
1.18.2
1.18.1
1.18.0
1.17.5
...
```

### List installed versions

```bash
$ kbenv list local
1.18.0
1.16.4
1.14.8
```

### Install version

```bash
$ kbenv install 1.16.5
Downloading binary...
Done! Saving it at /home/user/.bin/kubectl-v1.16.5
```

### Use version

```bash
$ kbenv use 1.16.5
Done! Using 1.16.5 version.
```

To use the automatic detection of the cluster and forget about it, just set it
to `auto`:

```bash
$ kbenv use auto
Done! Using auto version.
```

### Uninstall version

```bash
$ kbenv uninstall 1.16.5
Done! 1.16.5 version uninstalled from /home/ap/.bin/kubectl-v1.16.5.
```

## FAQ
### Why migrate from bash to go?
The project just as a way of downloading the binary versions. Progressively it
began to grow a little. And then they came some PR for different stuff, but the
hard ones where the ones for adding better support for MacOS. I don't own a Mac,
so I couldn't test them properly.

Also, `helmenv` and `kbenv` where pretty much a copy paste, but they didn't have
the same code, so any change from one place I would have to add it to the other.

So, with this to problems (and because I was bored) I decided to migrate them
for Go. Go is cool because it lets you have self contained binaries, so no more
worries about the OS! I even add support for Windows, because why not. And also,
being Go a real programming language, I could add tests.

### How to migrate from the bash version
For doing so you have to:
- Delete the `kbenv` repository: `rm -r ~/.kbenv`
- Delete the line that sources the bash script: `source $HOME/.kbenv/kbenv.sh`

And that's it. The way how the Go version works is very similar. The changed
beehaviours are:

- You don't have to set the `v` before the versions. For example:

``` bash
$ kbenv install v2.0.1
# Would be
$ kbenv install 2.0.1
```

- The listing commands have been separed:

``` bash
# Before
$ kbenv list
$ kbenv list-remote
# After
$ kbenv list local
$ kbenv list remote
```

## How to enforce a kubectl version
Just create a `.kubectl_version` in your directory pointing to the version you want
to use. For example:

``` bash
$ kbenv install 1.18.0
...
$ kbenv install 1.18.2
...
$ kbenv use 1.18.2
...
$ kubectl version --client
Client Version: version.Info{Major:"1", Minor:"18", GitVersion:"v1.18.2", GitCommit:"52c56ce7a8272c798dbc29846288d7cd9fbae032", GitTreeState:"clean", BuildDate:"2020-04-16T11:56:40Z", GoVersion:"go1.13.9", Compiler:"gc", Platform:"linux/amd64"}
$ echo 1.18.0 > .kubectl_version
$ kubectl version --client
Client Version: version.Info{Major:"1", Minor:"18", GitVersion:"v1.18.0", GitCommit:"9e991415386e4cf155a24b1da15becaa390438d8", GitTreeState:"clean", BuildDate:"2020-03-25T14:58:59Z", GoVersion:"go1.13.8", Compiler:"gc", Platform:"linux/amd64"}
```

## License
GPL3
