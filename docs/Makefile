# You can set these variables from the command line.
SPHINXOPTS    =
SPHINXBUILD   = $(VENV_PATH)/bin/python -msphinx
SPHINXPROJ    = Navigator
SOURCEDIR     = .
VENV_PATH     = .venv
BUILDDIR      = _build

# Put it first so that "make" without argument is like "make help".
help:
	@$(SPHINXBUILD) -M help "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

.PHONY: help Makefile

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: .venv Makefile
	@$(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

.venv:
	virtualenv -p $(shell which python3) $(VENV_PATH)
	$(VENV_PATH)/bin/pip install -r requirements.txt
	touch .venv

check: .venv
	$(SPHINXBUILD) "$(SOURCEDIR)" "$(BUILDDIR)" -b linkcheck
	$(SPHINXBUILD) "$(SOURCEDIR)" "$(BUILDDIR)" -b spelling
