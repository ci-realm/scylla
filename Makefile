.SUFFIXES:

run:
	gin run
.PHONY: run

# REVISION=$$(git rev-parse --verify HEAD) kubernetes-deploy --template-dir olympus/misc.production e-recruiting-api-team kubernetes.misc.production.$$dc; \

deploy: olympus/misc.production/deployment.json
	docker load -i "$$(nix-build ./ci.nix -A docker)" && \
  docker push "quay.dc.xing.com/e-recruiting-api-team/scylla:$$(git rev-parse --verify HEAD)" && \
	for dc in ams1 ams2; do \
	  kubectl --namespace e-recruiting-api-team \
	          --context "kubernetes.misc.production.$$dc" \
	          apply -f "$$(nix-build ./olympus)"; \
	done
.PHONY: deploy

sync-secrets:
	for dc in ams1 ams2; do \
	  public_key=$$(jq -r ._public_key < olympus/misc.production/secrets.ejson); \
	  private_key=$$(< /opt/ejson/keys/$$public_key); \
	  kubectl create secret generic ejson-keys "--from-literal=$$public_key=$$private_key" --context "kubernetes.misc.production.$$dc" --namespace e-recruiting-api-team || \
	  kubectl patch secret ejson-keys -p "{\"data\":{\"public_key\":\"$$( echo -n $$private_key | base64 -w0)}}\"}}" --context "kubernetes.misc.production.$$dc" --namespace e-recruiting-api-team; \
	done
.PHONY: sync-secrets

olympus/misc.production/deployment.json:
	nix build --show-trace -f olympus -o olympus/misc.production/deployment.json
.PHONY: olympus/misc.production/deployment.json

olympus/misc.production/deployment.yml: olympus/misc.production/deployment.json
	remarshal -if json -of yaml -i olympus/misc.production/deployment.json -o olympus/misc.production/deployment.yml
.PHONY: olympus/misc.production/deployment.json
