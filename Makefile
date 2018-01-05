.PHONY: default
default: all

.PHONY: all
all: package-apex package-vf package-lightning

docset-gen:
	dep ensure
	go build -x -o docset-gen ./SFDashC/

.PHONY: run-apex
run-apex: clean-index docset-gen
	./docset-gen apexcode

.PHONY: run-vf
run-vf: clean-index docset-gen
	./docset-gen pages

.PHONY: run-lightning
run-lightning: clean-index docset-gen
	./docset-gen lightning

package-apex: run-apex
	./package-docset.sh Apex

.PHONY: package-vf
package-vf: run-vf
	./package-docset.sh Pages

.PHONY: package-lightning
package-lightning: run-lightning
	./package-docset.sh Lightning

.PHONY: archive
archive:
	find *.docset -depth 0 | xargs -I '{}' sh -c 'tar --exclude=".DS_Store" -czf "$$(echo {} | sed -e "s/\.[^.]*$$//" -e "s/ /_/").tgz" "{}"'
	@echo "Archives created!"

.PHONY: clean-index
clean-index:
	rm -f ./build/docSet.dsidx

.PHONY: clean-package
clean-package:
	rm -fr *.docset

.PHONY: clean-archive
clean-archive:
	rm -f *.tgz

.PHONY: clean
clean: clean-index clean-package clean-archive
	rm -f docset-gen

.PHONY: clean-build
clean-build:
	rm -fr ./build
