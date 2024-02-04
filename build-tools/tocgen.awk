#!/usr/bin/env -S awk -f
function mklink(heading) {
	link = tolower(heading)
	gsub(" ", "-", link)
	occurrence = links[link]++
	if (occurrence > 0) {
		link = link "-" occurrence
	}
	return link
}
/^## / {
	heading = substr($0, 4)
	if (h2count > 0) {
		printf "\n"
	}
	printf "%d. [**%s**](#%s)", ++h2count, heading, mklink(heading)
	h3count = 0
	next
}
/^### / {
	if (h3count++ == 0) {
		printf "  \n   "
		extra_spaces = int(log(h2count)/log(10))
		while (extra_spaces-- > 0) {
			printf " "
		}
	} else {
		printf ", "
	}
	heading = substr($0, 5) # strip heading marker
	gsub("`", "", heading)
	printf "[%s](#%s)", heading, mklink(heading)
}
END { printf "\n" }
