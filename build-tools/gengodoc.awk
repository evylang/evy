#!/usr/bin/env -S awk -f
$0 ~ "//\tUsage:" {
        in_usage = 1
        cmd = $3
}
$0 ~ "^(package |// [[])" && in_usage {
        system(cmd " --help | sed -e '/./s|^|//\t|' -e 's|^$|//|'")
        if ($1 == "//") printf "//\n"
        in_usage = 0
}

!in_usage { print }
