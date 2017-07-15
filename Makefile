.PHONY: all clean-index package-apex clean-index package-vf clean-index package-combined

default: all

all: clean-index package-apex clean-index package-vf clean-index package-combined

run-apex: clean-index
	dep ensure
	go run ./SFDashC/*.go apexcode

run-vf: clean-index
	dep ensure
	go run ./SFDashC/*.go pages

run-combined: clean-index
	dep ensure
	go run ./SFDashC/*.go apexcode pages

package-apex: run-apex
	$(eval name = Apex)
	$(eval package = Salesforce $(name).docset)
	$(eval version = $(shell cat ./build/apexcode-version.txt))
	cat ./SFDashC/docset-apexcode.json | sed s/VERSION/$(version)/ > ./build/docset-apexcode.json
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r ./build/atlas.en-us.apexcode.meta "$(package)/Contents/Resources/Documents/"
	cp ./build/*.html "$(package)/Contents/Resources/Documents/"
	cp ./build/*.css "$(package)/Contents/Resources/Documents/"
	cp ./SFDashC/Info-$(name).plist "$(package)/Contents/Info.plist"
	cp ./build/docSet.dsidx "$(package)/Contents/Resources/"
	@echo "Docset generated!"

package-vf: run-vf
	$(eval name = Pages)
	$(eval package = Salesforce $(name).docset)
	$(eval version = $(shell cat ./build/pages-version.txt))
	cat ./SFDashC/docset-pages.json | sed s/VERSION/$(version)/ > ./build/docset-pages.json
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r ./build/atlas.en-us.pages.meta "$(package)/Contents/Resources/Documents/"
	cp ./build/*.html "$(package)/Contents/Resources/Documents/"
	cp ./build/*.css "$(package)/Contents/Resources/Documents/"
	cp ./SFDashC/Info-$(name).plist "$(package)/Contents/Info.plist"
	cp ./build/docSet.dsidx "$(package)/Contents/Resources/"
	@echo "Docset generated!"

package-combined: run-combined
	$(eval name = Combined)
	$(eval package = Salesforce $(name).docset)
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r ./build/*.meta "$(package)/Contents/Resources/Documents/"
	cp ./build/*.html "$(package)/Contents/Resources/Documents/"
	cp ./build/*.css "$(package)/Contents/Resources/Documents/"
	cp ./SFDashC/Info-$(name).plist "$(package)/Contents/Info.plist"
	cp ./build/docSet.dsidx "$(package)/Contents/Resources/"
	@echo "Docset generated!"

archive:
	find *.docset -depth 0 | xargs -I '{}' sh -c 'tar --exclude=".DS_Store" -czf "$$(echo {} | sed -e "s/\.[^.]*$$//" -e "s/ /_/").tgz" "{}"'
	@echo "Archives created!"

clean-index:
	rm -f ./build/docSet.dsidx

clean: clean-index
	rm -fr ./build
	rm -fr *.docset
	rm -f *.tgz
