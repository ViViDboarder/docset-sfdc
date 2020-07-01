SFDashC
=======

SFDashC is a go application for downloading and constructing Dash docsets from the Salesforce online documentation

Everything is wrapped with a Makefile and can be completely built by executing:

    make

That's it!

It will generate 3 docsets: Salesforce Apex, Salesforce Visualforce, and Salesforce Lightning

To Do
-----

 - [ ] Now that new `ForceCascadeType` is available, some of the entries in `./SFDashC/supportedtypes.go` can be simplified
