# Sourced by other tapes. termenv treats CI=1 as non-TTY and strips ANSI; many
# setups also set NO_COLOR. Force a normal color-capable terminal for recordings.
Env CI ""
Env NO_COLOR ""
Env TERM "xterm-256color"
Env COLORTERM "truecolor"
