.PHONY: default
default: all

.PHONY: all
all: package-apex package-vf package-lightning

vendor:
	dep ensure

.PHONY: run-apex
run-apex: clean-index vendor
	go run ./SFDashC/*.go apexcode

.PHONY: run-vf
run-vf: clean-index vendor
	go run ./SFDashC/*.go  pages

.PHONY: run-lightning
run-lightning: clean-index vendor
	go run ./SFDashC/*.go lightning

.PHONY: package-apex
package-apex: run-apex
	./package-docset.sh apexcode

.PHONY: package-vf
package-vf: run-vf
	./package-docset.sh pages

.PHONY: package-lightning
package-lightning: run-lightning
	./package-docset.sh lightning

.PHONY: archive-apex
archive-apex: package-apex
	./archive-docset.sh apexcode

.PHONY: archive-vf
archive-vf: package-vf
	./archive-docset.sh pages

.PHONY: archive-lightning
archive-lightning: package-lightning
	./archive-docset.sh lightning

.PHONY: archive-all
archive-all:  archive-apex archive-vf archive-lightning

.PHONY: clean-index
clean-index:
	rm -f ./build/docSet.dsidx

.PHONY: clean-package
clean-package:
	rm -fr *.docset

.PHONY: clean-archive
clean-archive:
	rm -f *.tgz
	rm -fr ./archive

.PHONY: clean
clean: clean-index clean-package clean-archive

.PHONY: clean-build
clean-build:
	rm -fr ./build

.PHONY: clean-vendor
clean-vendor:
	rm -fr ./vendor

.PHONY: clean-all
clean-all: clean clean-build clean-vendor
