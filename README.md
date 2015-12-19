Box creates box plots for the plan9 plot(1) command.
It reads data sets of the form `<name> <number>*` from standard input,
and outputs a series of box plots for plot(1) on standard output.

Example:
`echo "linear 1 2 3 4 5 6 exponential 2 4 8 16 32 64" | box -t Title | plot`
shows two box plots,
one labeled "linear", showing the distribution of the numbers 1 2 3 4 5 6,
and one labeled "exponential" showing the distribution of 2 4 8 16 32 64.
