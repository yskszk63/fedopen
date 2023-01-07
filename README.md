# fedopen

Open AWS Management Console With API Credential.

## Usage

```
$ AWS_PROFILE=xxx fedopen
$ # Management Console is opened by xdg-open.
$ 
$ # or
$ 
$ # URL is output to stdout if xdg-open is not present.
$ env -i AWS_PROFILE=xxx fedopen
https://signin.aws.amazon.com/federation?Action=login&Destination=https%3A%2F%2Fconsole.aws.amazon.com%2F&Issuer=fedopen&SigninToken=pv7yws-hQ...
$ 
```

## License

[MIT](LICENSE)

## Author

[yskszk63](https://github.com/yskszk63)
