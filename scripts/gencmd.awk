#!/usr/bin/env -S awk -f
/<!-- gen:.* -->/ {
        cmd = $0
        sub(".*<!-- gen:", "", cmd)
        sub(" -->.*", "", cmd)
        print
        printf "\n"
        ignore = 1
}
/<!-- genend -->/ {
        system(cmd " | sed '/./s/^/    /'")
        printf "\n"
        ignore = 0
}

!ignore { print }
