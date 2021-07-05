# Terraform Provider Anypoint

Run the following command to build the provider

```bash
$ go build -o terraform-provider-anypoint
```

**N.B:** As of Go 1.13 make sure that your `GOPRIVATE` environment variable includes `github.com/mulesoft-consulting` 

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

## Debugging mode

First build the project using
```bash
$ go build
```

You should have a new file `terraform-provider-anypoint` in the root of the project. To start the provider in debug mode execute the following: 
```bash
$ dlv exec --headless ./terraform-provider-anypoint -- --debug
```

Once executed, connect your debugger (whether it's your IDE or the debugger client) to the debugger server. The following is an example of how to start a client debugger:
```bash
$ dlv connect 127.0.0.1:51495
```

Then have your client debugger `continue` execution (check the help for more info) then your provider should print something like: 
```bash
TF_REATTACH_PROVIDERS='{"anypoint.mulesoft.com/automation/anypoint":{"Protocol":"grpc","Pid":69612,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/yc/k0_j_x0945jdthsw7fzw5ysh0000gp/T/plugin598168131"}}}'
```

Now you can run terraform using the debugger, here's an example: 

```bash
$ TF_REATTACH_PROVIDERS='{"anypoint.mulesoft.com/automation/anypoint":{"Protocol":"grpc","Pid":69612,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/yc/k0_j_x0945jdthsw7fzw5ysh0000gp/T/plugin598168131"}}}' terraform apply --auto-approve -var-file="params.tfvars.json"
```