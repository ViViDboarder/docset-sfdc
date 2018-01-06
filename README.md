SFDashC
=======

SFDashC is a go application for downloading and constructing Dash docsets from the Salesforce online documentation

Everything is wrapped with a Makefile and can be completely built by simply executing:

    make

That's it!

It will generate 3 docsets: Salesforce Apex, Salesforce Visualforce, and Salesforce Lightning

Dependencies
------------

All dependencies are being managed by [dep](https://github.com/golang/dep). Dep must be installed for the vendor folder to be built.

To Do
-----

 - [ ] Now that new `ForceCascadeType` is available, some of the entries in `./SFDashC/supportedtypes.go` can be simplified
 - [ ] Allow archiving of multiple versions and fetching pre-releases
