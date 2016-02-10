default: complete

complete: run-combined package-apex package-vf package-combined

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
	cp -r SFDashC/atlas.en-us.200.0.apexcode.meta "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.html "$(package)/Contents/Resources/Documents/"
	cp SFDashC/Info-$(type).plist "$(package)/Contents/Info.plist"
	cp SFDashC/docSet.dsidx "$(package)/Contents/Resources/"

package-vf:
	$(eval type = Pages)
	$(eval package = Salesforce $(type).docset)
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r SFDashC/atlas.en-us.200.0.pages.meta "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.html "$(package)/Contents/Resources/Documents/"
	cp SFDashC/Info-$(type).plist "$(package)/Contents/Info.plist"
	cp SFDashC/docSet.dsidx "$(package)/Contents/Resources/"

package-combined:
	$(eval type = Combined)
	$(eval package = Salesforce $(type).docset)
	mkdir -p "$(package)/Contents/Resources/Documents"
	cp -r SFDashC/*.meta "$(package)/Contents/Resources/Documents/"
	cp SFDashC/*.html "$(package)/Contents/Resources/Documents/"
	cp SFDashC/Info-$(type).plist "$(package)/Contents/Info.plist"
	cp SFDashC/docSet.dsidx "$(package)/Contents/Resources/"

clean:
	rm -fr SFDashC/*.meta
	rm -f SFDashC/docSet.dsidx
	rm -fr *.docset
