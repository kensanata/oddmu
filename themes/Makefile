download:
	rsync --archive --delete sibirocobombus:alexschroeder.ch/wiki/'*.html' --exclude='*-*' alexschroeder.ch/
	rsync --archive sibirocobombus:alexschroeder.ch/css/oddmu-2023.css alexschroeder.ch/oddmu.css
	sed --in-place=~ --expression='s/\/css\/oddmu-2023\.css/oddmu.css/' alexschroeder.ch/*.html
	rsync --archive --delete sibirocobombus:flying-carpet.ch/wiki/'*.html' flying-carpet.ch/
	rsync --archive --delete sibirocobombus:campaignwiki.org/wiki/'*.html' campaignwiki.org/
	rsync --archive --delete sibirocobombus.root:/home/oddmu/'*.html' transjovian.org/
