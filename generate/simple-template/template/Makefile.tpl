cur_makefile_path := $(abspath $(lastword ./))
binName = $(shell echo $(cur_makefile_path)|awk -F '/' '{ print $$NF }')

ifeq ($(shell uname),Windows_NT)
	suffix = .exe
endif

.PHONY : build

build :
    go mod tidy
	go build -o ./bin/$(binName)$(suffix)
