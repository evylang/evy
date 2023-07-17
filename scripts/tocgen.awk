#!/usr/bin/env -S awk -f
function mklink(heading) {
	link = tolower(heading)
	gsub(" ", "-", link)
	return link
}
/^## / {
	heading = substr($0, 4)
	link = mklink(heading)
	printf "\n%d. [**%s**](#%s)  \n   ", ++heading_count, heading, link
	extra_spaces = int(log(heading_count)/log(10))
	while (extra_spaces-- > 0) {
		printf " "
	}
	separator = ""
	next
}
/^### / {
	heading = substr($0, 5) # strip heading marker
	gsub("`", "", heading)
	link = mklink(heading)
	occurrence = funcs[heading]++
	if (occurrence > 0) {
		link = link "-" occurrence
	}
	printf "%s[%s](#%s)", separator, heading, link
	separator = ", "
}
END { printf "\n" }
