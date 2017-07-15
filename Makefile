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
	$(eval version = $(shell cat SFDashC/apexcode-version.txt))
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r SFDashC/atlas.en-us.apexcode.meta "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.html "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.css "$(package)/Contents/Resources/Documents/"
	cp SFDashC/Info-$(name).plist "$(package)/Contents/Info.plist"
	cp SFDashC/docSet.dsidx "$(package)/Contents/Resources/"
	@echo "Docset generated!"

package-vf: run-vf
	$(eval name = Pages)
	$(eval package = Salesforce $(name).docset)
	$(eval version = $(shell cat SFDashC/pages-version.txt))
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r SFDashC/atlas.en-us.pages.meta "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.html "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.css "$(package)/Contents/Resources/Documents/"
	cp SFDashC/Info-$(name).plist "$(package)/Contents/Info.plist"
	cp SFDashC/docSet.dsidx "$(package)/Contents/Resources/"
	@echo "Docset generated!"

package-combined: run-combined
	$(eval name = Combined)
	$(eval package = Salesforce $(name).docset)
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r SFDashC/*.meta "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.html "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.css "$(package)/Contents/Resources/Documents/"
	cp SFDashC/Info-$(name).plist "$(package)/Contents/Info.plist"
	cp SFDashC/docSet.dsidx "$(package)/Contents/Resources/"
	@echo "Docset generated!"

archive:
	find *.docset -depth 0 | xargs -I '{}' sh -c 'tar --exclude=".DS_Store" -czf "$$(echo {} | sed -e "s/\.[^.]*$$//" -e "s/ /_/").tgz" "{}"'
	@echo "Archives created!"

clean-index:
	rm -f SFDashC/docSet.dsidx

clean: clean-index
	rm -fr SFDashC/*.meta
	rm -fr *.docset
	rm -f SFDashC/*.css
	rm -f *.tgz
