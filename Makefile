.SUFFIXES:

run:
	gin run
.PHONY: run

deploy:
	for dc in ams1 ams2; do \
		kubectl --namespace e-recruiting-api-team \
		        --context "kubernetes.misc.production.$$dc" \
		        apply -f "$$(nix-build ./olympus)"; \
	done
.PHONY: deploy
