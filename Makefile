.PHONY: gifs
# Run from repo root so vhs/_env.tape resolves and lipgloss gets color (see README).
gifs:
	mkdir -p screenshots
	vhs < vhs/simple.tape
	vhs < vhs/confirm.tape
	vhs < vhs/form.tape
	vhs < vhs/spinner.tape
	vhs < vhs/colors.tape
	vhs < vhs/transparency.tape
