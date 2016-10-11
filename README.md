[![Build Status](https://travis-ci.org/kamatama41/tfenv.svg?branch=master)](https://travis-ci.org/kamatama41/tfenv)

# tfenv
[Terraform](https://www.terraform.io/) version manager inspired by [rbenv](https://github.com/rbenv/rbenv)

## Support
Currently tfenv supports the following OSes
- Mac OS X (64bit)
- Linux (64bit)

## Installation
1. Check out tfenv into any path
```sh
$ git clone https://github.com/kamatama41/tfenv.git /some/path/to/tfenv
```
2. Add `/some/path/to/tfenv/bin` to your `$PATH`

This is an example of adding `.zshrc`
```
if [ -f /some/path/to/tfenv/bin/tfenv ]; then
  path=(/some/path/to/tfenv/bin $path)
fi
```

## Usage
### tfenv install
Install a specific version of Terraform
```sh
$ tfenv install 0.7.0
$ tfenv install latest # latest version
```

If you use [.terraform-version](#terraform-version), `tfenv install` (no argument) will install the version written in it.

### tfenv use
Switch a version to use
```sh
$ tfenv use 0.7.0
```

### tfenv list
List installed versions
```sh
% tfenv list
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
0.7.0
0.7.0-rc4
0.7.0-rc3
0.7.0-rc2
0.7.0-rc1
0.6.16
0.6.15
0.6.14
0.6.13
...
```

## .terraform-version
If you put `.terraform-version` file on your project root, tfenv detects it and use the version written in it.

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

```

## Upgrading
```sh
$ git --git-dir=/some/path/to/tfenv/.git pull
```

## Uninstalling
```sh
$ rm -rf /some/path/to/tfenv
```

## LICENSE
- [tfenv itself](https://github.com/kamatama41/tfenv/blob/master/LICENSE)
- [rbenv](https://github.com/rbenv/rbenv/blob/master/LICENSE)
  - tfenv partially uses rbenv's source code
