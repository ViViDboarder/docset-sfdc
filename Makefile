default: complete

complete: clean-index run-apex package-apex clean-index run-vf package-vf

run-apex:
	dep ensure
	(cd SFDashC && go run *.go --silent apexcode)

run-vf:
	dep ensure
	(cd SFDashC && go run *.go --silent pages)

run-combined:
	dep ensure
	(cd SFDashC && go run *.go --silent apexcode pages)

package-apex:
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

package-vf:
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

package-combined:
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
