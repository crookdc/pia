# pia 

* Configurations 
  * YAML files for structured data
  * Simple Pipsqueak scripting capabilities for hooks (before request, after request)
  * Project, environment, session, and runtime variables.
    * Project variables are persisted to disk. 
    * Environment variables are persisted in a human-editable format and can be hot-swapped between files acting as 
      different environments for the same project.
    * Session variables exists in memory for as long as the TUI application is running.
    * Runtime variables are either provided in the run command or prompted for at runtime
* TUI application
  * VIM-like command execution (for quitting application and run configurations)
  * Status output for Pipsqueak as well as HTTP output (unless call is configured as quiet)

## Configuration
The configuration will always go through a substitution pass before being parsed. This means that a user can use 
variables freely anywhere in the configuration. 
```yml
version: 1.0
name: Authenticate
request:
    url: ${environment.base}/auth
    method: POST
    headers:
      Authorization: Bearer ${session.token}
      Content-Type: application/json
    query:
      version: ${project.version}
    body:
      text: >
        {
          "username": "${environment.username}",
          "password": "${environment.password}"
        }
hooks:
  post: save-token.sqk
---
version: 1.0
name: Get Students
request:
  url: ${environment.base}/schools/${runtime.school}/students
  method: POST
  headers:
    Authorization: Bearer ${session.token}
    Content-Type: application/json
  query:
    version: ${project.version}
    class: ${runtime.class}
```

save-token.sqk:
```
let session = import "session";

// assert causes the script execution to halt and prints an error message to the pia output 
assert(response["code"] == 200, "bad status code: " + response["code"]);

let token = response["body"]["token"]
assert(token != "", "no token in response");

let expires = response["body"]["_expires"]
assert(expires != "", "no expiry in response");

session.save("token", token);
session.save("expires", expires):
```

## Scratchpad
* Automatic generation of project from SOAP WSDL
```
project/
  .pia/
  authenticate/
    spec.yml
    post.sqk
  students/
    create/
      spec.yml
    delete/
      spec.yml
```