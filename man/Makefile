TEXT=$(wildcard *.txt)
MAN=$(patsubst %.txt,%,${TEXT})
HTML=$(patsubst %.txt,%.html,${TEXT})
MD=$(patsubst %.txt,%.md,${TEXT})

man: ${MAN}

%: %.txt
	scdoc < $< > $@

html: ${HTML}

%.html: %
	groff -mandoc -Dutf8 -Thtml $< | sed 's/<style type="text\/css">/<style type="text\/css">\n       body {font-family: mono; max-width: 80ch }/' > $@

md: ${MD}

%.md: %.txt
	sed --regexp-extended \
	  -e 's/\*([^*]+)\*/**\1**/g' \
	  -e 's/_(oddmu[a-z.-]*)_\(([1-9])\)/[\1(\2)](\1.\2)/g' \
	  -e 's/\b_([^_]+)_\b/*\1*/g' \
	  -e 's/^#/##/' \
	  -e 's/^([A-Z.-]*\([1-9]\))( ".*")?$$/# \1/' \
	  < $< > $@

upload: ${MD}
	rsync --itemize-changes --archive *.md sibirocobombus:alexschroeder.ch/wiki/oddmu/
	make clean

clean:
	rm --force ${HTML} ${MD}
