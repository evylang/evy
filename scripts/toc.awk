#!/usr/bin/env -S awk -f
/<!-- gen:toc -->/ {
	in_toc = 1
	print
}
/<!-- genend:toc -->/ {
	in_toc = 0
	system("awk -f scripts/tocgen.awk " FILENAME); printf "\n";
}
in_toc {
	next
}

{ print }
