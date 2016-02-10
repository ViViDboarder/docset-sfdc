SFDashC
=======

SFDashC is a go application for downloading and constructing Dash docsets from the Salesforce online documentation

Everything is wrapped with a Makefile and can be completely built by simply executing:

    make

That's it!

It will generate 3 docsets: Salesforce Apex, Salesforce Visualforce, and Salesforce Combined

Dependencies
------------

Currently these are not auto resolved. You must install the following:
    * github.com/coopernurse/gorp
    * github.com/mattn/go-sqlite3
