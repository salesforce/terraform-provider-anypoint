# Terraform Provider Anypoint

Run the following command to build the provider

```bash
$ go build -o terraform-provider-anypoint
```

**N.B:** As of Go 1.13 make sure that your `GOPRIVATE` environment variable includes `github.com/mulesoft-consulting` 

```bash
$ go env -w GOPRIVATE=github.com/mulesoft-consulting
```

## Test sample configuration

First, build and install the provider.

```bash
$ make install
```

Then, navigate inside the `examples` folder, and update your credentials in `main.tf`.   
Run the following command to initialize the workspace and apply the sample configuration.

```bash
$ terraform init && terraform apply
```

If you prefer to have your credentials in a separate file, create a `params.tfvars.json` file in the `examples` folder. Then add your parameters as shown in the example below: 

```json
{
  "client_id": "REMPLACE_HERE",
  "client_secret": "REMPLACE_HERE",
  "org_id": "REMPLACE_HERE"
}
```
Make sure to add the params file when you apply your terraform configuration as follow:
```bash
$ terraform init && terraform apply -var-file="params.tfvars.json"
```
