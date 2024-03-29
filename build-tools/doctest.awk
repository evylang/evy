#!/usr/bin/env -S awk -f
#
# doctest.awk formats all evy code blocks in a markdown file and executes that
# code feeding it the specified input and replacing the output code block with
# the output of running that evy code.
#
# Evy code is identified by a triple-backtick code block with the language
# specified as "evy". Input that should be provided to the evy program on
# stdin should follow in another code block with the language specified as
# "evy:input". If a code block with the language "evy:output" follows,
# its contents will be replaced with the output of executing the evy program.
# If the output contains the escape sequence for clearing the terminal, that
# and all preceding output are discarded.
#
# e.g.
#
# ```evy
# read name
# print "Hello" name
# ```
#
# ```evy:input
# Fox
# ```
#
# ```evy:output
# Hello Fox
# ```
#
# If an evy code block is specified with the language "evy:err" then the code
# block is taken as evy code that is executed but any evy error should not be
# treated as a failure. This is useful for evy code blocks that demonstrate
# errors. The output is still captured and placed in the next "evy:output"
# block.

BEGIN {
	reset()
	"clear" | getline clear
	close("clear")
	error_count = 0
}

function reset() {
	code = input = output = ""
	in_code = in_input = in_output = 0
	expect_err = 0
}

# accumulate lines in a buffer, leaving off the final newline
function accumulate_line(line, buffer) {
	if (buffer == "") {
		return line
	}
	return buffer RS line
}

# execute executes cmd with the given input returning the output of
# the command with the trailing newline stripped. If the command
# exited with an error, an error is written to stderr and "" is returned
# instead. An empty string should never be returned otherwise -
# a formatted program must contain text, and an output block would
# contain some output.
function execute(cmd, input) {
	tempfile = "/tmp/doctest.tmp"
	print input | (cmd ">" tempfile " 2>&1")
	rv = close(cmd ">" tempfile " 2>&1")

	o = ""
	while (getline line < tempfile) {
		o = accumulate_line(line, o)
	}
	close(tempfile)
	system("rm " tempfile)

	if (rv != 0 && expect_err == 0) {
		split(cmd, args)
		print "Line " code_line ": Error running 'evy " args[2] "' for:", builtin > "/dev/stderr"
		print o > "/dev/stderr"
		close("/dev/stderr")
		return ""
	}
	return o
}

# Builtin title. Save the name for errors and reset the state for detecting
# code and input/output blocks.
/^### / {
	reset()
	builtin = $2
	print; next
}

# See evy code block. Accumulate code, then run through `evy fmt` when
# we get to the end of the code block. If `evy fmt` returns an error,
# just print out the original code and send the error to stderr. Otherwise
# output the formatted code in the place of the code.
/^```evy(:err)?$/ {
	reset()
	in_code = 1
	code_line = NR
	if ($1 ~ /:err/) {
		expect_err = 1
	}
	print; next
}
/^```$/ && in_code {
	in_code = 0
	v = execute("evy fmt", code)
	if (v == "") {
		# error formatting code. just print out the original code.
		# an error has already been written to stderr
		print code
		code = "" # empty it so we don't try to run it later
		error_count++
	} else {
		print v
	}
	print; next
}
in_code {
	code = accumulate_line($0, code)
	next
}

# If we see an Input block, accumulate the input so we can feed it to
# `evy run` later. Only do so if we have some code to run (if `evy fmt`
# fails, we don't bother running the code.)
/^```evy:input$/ && code != "" {
	input = ""
	in_input = 1
	print; next
}
/^```$/ && in_input {
	in_input = 0
	print; next
}
in_input {
	input = accumulate_line($0, input)
	print; next
}

# If we see an Output block and we have some code to run (`evy fmt`
# succeeded), replace the contents of it with the output of `evy run`.
# We need to write the code to a file for `evy run` as stdin of evy
# needs to be attached to any Input data accumulated. We cannot feed
# both the code and the input through the same stdin.
/^```evy:output$/ && code != "" {
	output = ""
	in_output = 1
	print; next
}
/^```$/ && in_output {
	in_output = 0
	filename = "/tmp/sample.evy"

	print code > filename
	close(filename)
	
	v = execute("evy run --skip-sleep " filename, input)
	system("rm " filename)

	if (v == "") {
		# Error with evy run. Just print the original output instead
		print output
		error_count++
	} else {
		# Remove all text before a "clear" escape sequence, including the
		# escape sequence. Keep doing until there are none left.
		while ((i = index(v, clear)) > 0) {
			v = substr(v, i + length(clear))
		}
		print v
	}
	print; next
}
in_output {
	output = accumulate_line($0, output)
	next
}

{ print }

END {
	if (error_count > 0) {
		print error_count, "error(s) detected" > "/dev/stderr"
		exit 1
	}
}
