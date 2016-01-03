# Korwer

Korwer is a simple builder server (static website generators oriented).  
It can build your website based on Github webhook or via simple HTTP API.

Just set a `korwer.toml` configuration file like that:

```
[settings]
port = 5556
git = "/usr/bin/git"

[[site]]
name = "socketubs.org"
path = "/home/socketubs/Repositories/socketubs.github.io"
secret = "my-webhook-or-simple-secret"
build_cmd = "/usr/bin/git pull && /home/socketubs/.rvm/gems/ruby-2.2.1/bin/jekyll build"
deploy_cmd = "scp /home/socketubs/Repositories/socketubs.github.io/_site socketubs@my-server:"
```

You can review your configured website at `http://example.com/list`

Configure your Github webhook with url like that:
`http://example.com/webhook/socketubs.org`.  
Based on your website name declared in `korwer.toml`.

Or just do a `POST` request on `http://example.com/build` with a Json body like that:
```
{
  "site": "socketubs.org",
  "secret": "my-webhook-or-simple-secret"
}
```

And run korwer like that `./korwer`

Korwer will do a `git pull` in your site directory and run `jekyll build`.
This is an example for Jekyll but it can work for Pelican or any static website generator.

License is MIT.
