# Pia 
*/pɪæ/*
1. The open-source Postman alternative for technical people.
---

## Disclaimer
Even though this project presents itself as a Postman alternative it offers no guarantee to implement all features found 
in Postman. Though Postman has been the inspiration of user experience in several areas, Pia aims to be a tool that 
offers a user experience that can stand on its own two feet. Anyone is welcome to suggest the implementation of their 
favourite Postman features, but do not assume that just because Postman offers some feature that its in the roadmap for 
this project.

## State of Pia
The Pia project is still in **very** early development. There are no milestones or deadlines defined and there likely 
will never be any of the sort. Hopefully there will be alpha releases of Pia once development has reached a point where 
it can be used by the public in exchange for some feedback (not mandatory). You can of course clone and build Pia from 
source whenever you want but the quality of your experience cannot be guaranteed until an official release is created.

## Getting started
### Interpolation property sources
One of the core functions of Pia is to interpolate your text files and replace certain strings with values at runtime.
The aforementioned values can have one of several sources, each of which is described in this subsection. Each section
states the *context key* for the property source, this is the key used to identify which property source to fetch values
from. For example, `${session:id_token}` would be targeting the `session` property source and resolving the key 
`id_token`.

#### Environment
*Context key: `env`*

Fetches a value from the environment variables of the host machine.

#### Property file
*Context key: `props`*

Fetches a value from the property file passed to Pia as an argument during startup.

#### Session
*Context key: `session`*

Fetches a value from the current session. Squeak scripts are normally what would set these values. Examples of usage 
includes having a Squeak script store a bearer token in the session and then inserting it into the headers of each 
request made to protected endpoints.

---
*This readme is still under construction.*