download:
	rsync --archive --delete --itemize-changes --exclude='*-*' sibirocobombus:alexschroeder.ch/wiki/'*.html' alexschroeder.ch/
	rsync --archive sibirocobombus:alexschroeder.ch/css/oddmu-2023.css alexschroeder.ch/oddmu.css
	sed --in-place=~ --expression='s/\/css\/oddmu-2023\.css/oddmu.css/' alexschroeder.ch/*.html
	rsync --archive --delete --itemize-changes sibirocobombus:flying-carpet.ch/wiki/'*.html' flying-carpet.ch/
	rsync --archive --delete --itemize-changes sibirocobombus:campaignwiki.org/data/'*.html' campaignwiki.org/
	rsync --archive --delete --itemize-changes sibirocobombus.root:/home/oddmu/'*.html' transjovian.org/

# (ediff-directories "alexschroeder.ch" "/ssh:sibirocobombus:alexschroeder.ch/wiki/" "html$")
# (ediff-directories "flying-carpet.ch" "/ssh:sibirocobombus.root|sudo:claudia@sibirocobombus.root:/home/alex/flying-carpet.ch/wiki/" "html$")
# (ediff-directories "campaignwiki.org" "/ssh:sibirocobombus:campaignwiki.org/data/" "html$")
# (ediff-directories "transjovian.org" "/ssh:sibirocobombus.root:/home/oddmu/" "html$")

upload:
	rsync --archive --delete --itemize-changes --exclude=Makefile --exclude='*~' . sibirocobombus:alexschroeder.ch/wiki/oddmu/themes/

