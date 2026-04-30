# Sourced by other tapes. termenv treats CI=1 as non-TTY and strips ANSI; many
# setups also set NO_COLOR. Force a normal color-capable terminal for recordings.
Hide
Env CI ""
Env NO_COLOR ""
Set Env "TERM" "xterm-256color"
Set Env "COLORTERM" "truecolor"
