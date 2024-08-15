TEXT=$(wildcard *.txt)
MAN=$(patsubst %.txt,%,${TEXT})
HTML=$(patsubst %.txt,%.html,${TEXT})
MD=$(patsubst %.txt,%.md,${TEXT})

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
	@sed --regexp-extended \
	  -e 's/\*([^*]+)\*/**\1**/g' \
	  -e 's/_(oddmu[a-z.-]*)_\(([1-9])\)/[\1(\2)](\1.\2)/g' \
	  -e 's/\b_([^_]+)_\b/*\1*/g' \
	  -e 's/^# /## /' \
	  -e 's/#([^ #])/\\#\1/' \
	  -e 's/"(http.*?)"/`\1`/' \
	  -e 's/"(\[.*?\]\(.*?\))"/`\1`/' \
	  -e 's/^([A-Z.-]*\([1-9]\))( ".*")?$$/# \1/' \
	  < $< > $@

README.md: ../README.md
	@echo Making $@
	@sed --regexp-extended \
	  -e 's/\]\(.*\/(.*)\.txt\)/](\1)/' \
	  < $< > $@

upload: ${MD} README.md
	rsync --itemize-changes --archive *.md sibirocobombus:alexschroeder.ch/wiki/oddmu/
	make clean

clean:
	rm --force ${HTML} ${MD} README.md

realclean: clean
	rm --force ${MAN}
