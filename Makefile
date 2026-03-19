.PHONY: gifs
gifs:
	mkdir -p screenshots
	vhs < vhs/simple.tape
	vhs < vhs/confirm.tape
	vhs < vhs/form.tape
	vhs < vhs/spinner.tape
