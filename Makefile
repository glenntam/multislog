.PHONY: readme release

# Helper make command to automatically generate README.md
readme:
	@\goreadme -badge-godoc -badge-goreportcard -constants -factories -functions -methods -recursive -types -import-path github.com/glenntam/multislog > README.md.tmp
	@head -n $$(( $$(wc -l < README.md.tmp | tr -d ' ') - 3 )) README.md.tmp > README.md  # truncate last 3 lines
	@rm README.md.tmp
	# README.md sucessfully overwritten

release:
ifndef v
	$(error v is not set. Usage: make release v=v1.2.0)
endif
	git tag -s $(v) -m "release $(v)"
	git push
