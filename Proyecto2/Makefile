all:
	$(MAKE) -C kernel-module
	$(MAKE) -C ebpf-probe

clean:
	$(MAKE) -C kernel-module clean
	$(MAKE) -C ebpf-probe clean

compose-up:
	docker compose -f infra/docker-compose.yml up -d

compose-down:
	docker compose -f infra/docker-compose.yml down
