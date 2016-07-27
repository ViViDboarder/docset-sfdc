default: complete

complete: run-combined package-combined clean-index run-apex package-apex clean-index run-vf package-vf

run-apex:
	(cd SFDashC && go run *.go --silent apexcode)

run-vf:
	(cd SFDashC && go run *.go --silent pages)

run-combined:
	(cd SFDashC && go run *.go --silent apexcode pages)

package-apex:
	$(eval type = Apex)
	$(eval package = Salesforce $(type).docset)
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r SFDashC/atlas.en-us.apexcode.meta "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.html "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.css "$(package)/Contents/Resources/Documents/"
	cp SFDashC/Info-$(type).plist "$(package)/Contents/Info.plist"
	cp SFDashC/docSet.dsidx "$(package)/Contents/Resources/"

package-vf:
	$(eval type = Pages)
	$(eval package = Salesforce $(type).docset)
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r SFDashC/atlas.en-us.pages.meta "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.html "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.css "$(package)/Contents/Resources/Documents/"
	cp SFDashC/Info-$(type).plist "$(package)/Contents/Info.plist"
	cp SFDashC/docSet.dsidx "$(package)/Contents/Resources/"

package-combined:
	$(eval type = Combined)
	$(eval package = Salesforce $(type).docset)
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r SFDashC/*.meta "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.html "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.css "$(package)/Contents/Resources/Documents/"
	cp SFDashC/Info-$(type).plist "$(package)/Contents/Info.plist"
	cp SFDashC/docSet.dsidx "$(package)/Contents/Resources/"

clean-index:
	rm -f SFDashC/docSet.dsidx

clean: clean-index
	rm -fr SFDashC/*.meta
	rm -f SFDashC/*.css
	rm -fr *.docset
