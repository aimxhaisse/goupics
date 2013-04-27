CMD	:= daemonize
SRC	:= daemonizer/daemonizer.c
OBJ	:= $(SRC:.c=.o)
CFLAGS	:= -W -Wall -pedantic -ansi -O3 -Wno-unused-result
CC	:= gcc

all: $(CMD)
	go build

$(CMD):	$(OBJ)
	$(CC) $(OBJ) -o $(CMD)

clean:
	rm -f $(OBJ) $(CMD) gow
