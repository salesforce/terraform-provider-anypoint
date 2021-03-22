# Terraform Provider Hashicups

Run the following command to build the provider

## Dependencies
* Python 3.7  
```bash
brew install pyenv && pyenv install 3.7.10 && pyenv global 3.7.3 && $ echo -e 'if command -v pyenv 1>/dev/null 2>&1; then\n  eval "$(pyenv init -)"\nfi' >> ~/.zshrc
```
* pkg-config `brew install pkg-config`


```shell
go build -o terraform-provider-hashicups
```

## Test sample configuration

First, build and install the provider.

```shell
make install
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```
