obj-m += continfo.o

KERNEL_BUILD ?= /lib/modules/$(shell uname -r)/build

all:
	$(MAKE) -C $(KERNEL_BUILD) M=$(CURDIR) modules

clean:
	$(MAKE) -C $(KERNEL_BUILD) M=$(CURDIR) clean
