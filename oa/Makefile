# artisan/oa/Makefile

CC = gcc
# -D_GNU_SOURCE per le estensioni GNU (come asprintf o mount)
CFLAGS = -Wall -Wextra -Iinclude -D_GNU_SOURCE -Wno-format-truncation
LDFLAGS = -lm

# Struttura delle cartelle
SRC_DIR = src
OBJ_DIR = obj
INC_DIR = include

# Trova tutti i file .c ricorsivamente in src/
SOURCES = $(shell find $(SRC_DIR) -name '*.c')

# Mappa i sorgenti agli oggetti (es: src/utils/logger.c -> obj/utils/logger.o)
OBJECTS = $(SOURCES:$(SRC_DIR)/%.c=$(OBJ_DIR)/%.o)

# Nome del binario finale (rimane dentro la cartella oa/)
TARGET = oa

# Target principale
all: $(TARGET)

# Linker: unisce tutti gli oggetti nel binario finale
$(TARGET): $(OBJECTS)
	@echo "  LD    $@"
	@$(CC) $(OBJECTS) -o $@ $(LDFLAGS)

# Compilatore: trasforma ogni .c in un .o
# Crea le sottocartelle in obj/ se non esistono (es: obj/actions/)
$(OBJ_DIR)/%.o: $(SRC_DIR)/%.c
	@mkdir -p $(dir $@)
	@echo "  CC    $<"
	@$(CC) $(CFLAGS) -c $< -o $@

# Pulizia chirurgica: rimuove solo gli oggetti e il binario di oa
clean:
	@echo "  Cleaning oa..."
	@rm -rf $(OBJ_DIR) $(TARGET)
	@rm -f plan_coa_tmp.json

.PHONY: all clean