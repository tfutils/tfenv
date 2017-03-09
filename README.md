[![Build Status](https://travis-ci.org/kamatama41/tfenv.svg?branch=master)](https://travis-ci.org/kamatama41/tfenv)

# tfenv
[Terraform](https://www.terraform.io/) version manager inspired by [rbenv](https://github.com/rbenv/rbenv)

## Support
Currently tfenv supports the following OSes
- Mac OS X (64bit)
- Linux (64bit)
- Windows (64bit) - only tested in git-bash

## Installation
1. Check out tfenv into any path (here is `${HOME}/.tfenv`)

  ```sh
  $ git clone https://github.com/kamatama41/tfenv.git ~/.tfenv
  ```

2. Add `~/.tfenv/bin` to your `$PATH` any way you like

  ```sh
  $ echo 'export PATH="$HOME/.tfenv/bin:$PATH"' >> ~/.bash_profile
  ```

  OR you can make symlinks for `tfenv/bin/*` scripts into a path that is already added to your `$PATH` (e.g. `/usr/local/bin`) `OSX/Linux Only!`

  ```sh
  ln -s ~/.tfenv/bin/* /usr/local/bin
  ``` 

## Usage
### tfenv install
Install a specific version of Terraform  
`latest` is a syntax to install latest version
`latest:<regex>` is a syntax to install latest version matching regex (used by grep -e)
```sh
$ tfenv install 0.7.0
$ tfenv install latest
$ tfenv install latest:^0.8
```

If you use [.terraform-version](#terraform-version), `tfenv install` (no argument) will install the version written in it.

### tfenv use
Switch a version to use
`latest` is a syntax to use the latest installed version
`latest:<regex>` is a syntax to use latest installed version matching regex (used by grep -e)
```sh
$ tfenv use 0.7.0
$ tfenv use latest
$ tfenv use latest:^0.8
```

### tfenv list
List installed versions
```sh
% tfenv list
0.9.0-beta2
0.8.8
0.8.4
0.7.0
0.7.0-rc4
0.6.16
0.6.2
0.6.1
```

### tfenv list-remote
List installable versions
```sh
% tfenv list-remote
0.9.0-beta2
0.9.0-beta1
0.8.8
0.8.7
0.8.6
0.8.5
0.8.4
0.8.3
0.8.2
0.8.1
0.8.0
0.8.0-rc3
0.8.0-rc2
0.8.0-rc1
0.8.0-beta2
0.8.0-beta1
0.7.13
0.7.12
...
```

## .terraform-version
If you put `.terraform-version` file on your project root, tfenv detects it and use the version written in it. If the version is `latest` or `latest:<regex>`, the latest matching version currently installed will be selected.

```sh
$ cat .terraform-version
0.6.16

$ terraform --version
Terraform v0.6.16

Your version of Terraform is out of date! The latest version
is 0.7.3. You can update by downloading from www.terraform.io

$ echo 0.7.3 > .terraform-version

$ terraform --version
Terraform v0.7.3

$ echo latest:^0.8 > .terraform-version

$ terraform --version
Terraform v0.8.8
```

## Upgrading
```sh
$ git --git-dir=~/.tfenv/.git pull
```

## Uninstalling
```sh
$ rm -rf /some/path/to/tfenv
```

## LICENSE
- [tfenv itself](https://github.com/kamatama41/tfenv/blob/master/LICENSE)
- [rbenv](https://github.com/rbenv/rbenv/blob/master/LICENSE)
  - tfenv partially uses rbenv's source code

