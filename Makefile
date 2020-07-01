.PHONY: default test
default: all

.PHONY: all
all: package-apex package-vf package-lightning


.PHONY: run-apex
run-apex: clean-index
	go run ./SFDashC/*.go apexcode

.PHONY: run-vf
run-vf: clean-index
	go run ./SFDashC/*.go  pages

.PHONY: run-lightning
run-lightning: clean-index
	go run ./SFDashC/*.go lightning

.PHONY: package-apex
package-apex: run-apex
	./scripts/package-docset.sh apexcode

.PHONY: package-vf
package-vf: run-vf
	./scripts/package-docset.sh pages

.PHONY: package-lightning
package-lightning: run-lightning
	./scripts/package-docset.sh lightning

.PHONY: archive-apex
archive-apex: package-apex
	./scripts/archive-docset.sh apexcode

./archive/Salesforce_Apex: archive-apex

.PHONY: archive-vf
archive-vf: package-vf
	./scripts/archive-docset.sh pages

./archive/Salesforce_Visualforce: archive-vf

.PHONY: archive-lightning
./archive-lightning: package-lightning
	./scripts/archive-docset.sh lightning

./archive/Salesforce_Lightning: archive-lightning

.PHONY: archive-all
archive-all: archive-apex archive-vf # archive-lightning Lightning package isn't functional

./archive: archive-all

.PHONY: create-pr
create-pr: ./archive
	./scripts/create-pr.sh

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

.PHONY: clean-pr
clean-pr:
	rm -fr ./repotmp

.PHONY: clean-all
clean-all: clean clean-build clean-pr
