# Reproducible PDF build for the OpMem manuscript.
.PHONY: pdf clean check

MAIN = main
LATEXMK = latexmk

pdf:
	$(LATEXMK) -pdf -interaction=nonstopmode -halt-on-error $(MAIN).tex

check: pdf
	@test -f $(MAIN).pdf

clean:
	$(LATEXMK) -C $(MAIN).tex
