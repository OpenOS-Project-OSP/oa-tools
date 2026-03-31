	CC = gcc
# Usiamo += per aggiungere opzioni senza cancellare quelle precedenti
CFLAGS = -Wall -Wextra -Iinclude -D_GNU_SOURCE -Wno-format-truncation
LDFLAGS = -lm

# 1. Trova tutti i file .c ricorsivamente
SOURCES = $(shell find src -name '*.c')

# 2. Genera i nomi degli oggetti
OBJECTS = $(SOURCES:src/%.c=obj/%.o)

TARGET = oa

all: $(TARGET)

$(TARGET): $(OBJECTS)
	@echo "  LD    $@"
	@$(CC) $(OBJECTS) -o $@ $(LDFLAGS)
	@echo "--------------------------------------"
	@echo "Build completata con successo: ./$(TARGET)"

# Creazione cartelle e compilazione
obj/%.o: src/%.c
	@mkdir -p $(dir $@)
	@echo "  CC    $<"
	@$(CC) $(CFLAGS) -c $< -o $@

clean:
	@echo "  Cleaning up..."
	@rm -rf obj $(TARGET)

.PHONY: all clean