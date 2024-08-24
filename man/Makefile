TEXT=$(wildcard *.txt)
MAN=$(patsubst %.txt,%,${TEXT})
HTML=$(patsubst %.txt,%.html,${TEXT})
MD=$(patsubst %.txt,%.md,${TEXT})

help:
	@echo Help for Oddmu Documentation
	@echo ============================
	@echo make man
	@echo "    regenerate man pages"
	@echo make html
	@echo "    generate HTML pages"
	@echo make md
	@echo "    generate Markdown pages"
	@echo make clean
	@echo "    delete HTML and Markdown pages"
	@echo make realclean
	@echo "    delete HTML, Markdown and man pages"

man: ${MAN}

%: %.txt
	scdoc < $< > $@

html: ${HTML}

%.html: %.md
	@echo Making $@
	@echo '<!DOCTYPE html>' > $@
	@oddmu html $(basename $<) | sed --regexp-extended \
	  -e 's/<a href="(oddmu[a-z.-]*.[1-9])">([^<>]*)<\/a>/<a href="\1.html">\2<\/a>/g' >> $@

md: ${MD}

%.md: %.txt
	@echo Making $@
	@perl scdoc-to-markdown < $< > $@

README.md: ../README.md
	@echo Making $@
	@sed --regexp-extended \
	  -e 's/\]\(.*\/(.*)\.txt\)/](\1)/' \
	  < $< > $@

upload: ${MD} README.md
	rsync --itemize-changes --archive *.md sibirocobombus:alexschroeder.ch/wiki/oddmu/
	make clean

clean:
	@echo Removing HTML and Markdown files
	@rm --force ${HTML} ${MD} README.md

realclean: clean
	@echo Removing man pages
	@rm --force ${MAN}
