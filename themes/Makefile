SHELL=/usr/bin/fish

# Manually figure out what needs to change:
# (ediff-directories "alexschroeder.ch" "/ssh:sibirocobombus:alexschroeder.ch/wiki/" "html$")
# (ediff "alexschroeder.ch/oddmu.css" "/ssh:sibirocobombus:alexschroeder.ch/css/oddmu-2025.css")
# (ediff-directories "flying-carpet.ch" "/sudo:claudia@sibirocobombus.root:/home/alex/flying-carpet.ch/wiki/" "html$")
# (ediff-directories "campaignwiki.org" "/ssh:sibirocobombus:campaignwiki.org/data/" "html$")
# (ediff-directories "transjovian.org" "/ssh:sibirocobombus.root:/home/oddmu/" "html$")

# Upload the theme for the web site. This does not deploy the themes!
upload:
	rsync --archive --delete --itemize-changes --exclude=Makefile --exclude='*~' \
		. sibirocobombus:alexschroeder.ch/wiki/oddmu/themes/
	@echo Updated the templates for the Oddmu site
