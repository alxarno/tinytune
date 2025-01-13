.PHONY: watch clear

VENDOR = assets/vendor.bundle.js
INDEX =assets/index.min.js

$(VERBOSE).SILENT:

build: ${VENDOR} ${INDEX} ## build all assets (.js, .css)

${VENDOR}:
	npm run build-vendor

${INDEX}:
	npm run build

watch: ## build all assets in hot-reload way
	npm run watch

clear:
	rm -rf ${VENDOR} ${INDEX}