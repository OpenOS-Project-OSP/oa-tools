/*
 * oa: eggs in my dialect🥚🥚
 *
 * src/vendors/oe-yocto.c
 * Logica di classificazione utenti basata su OpenEmbedded-Core
 * e sulla filosofia di penguins-eggs.
 */
#include "oa.h"

/**
 * @brief Scrive una riga in formato passwd (USER:x:UID:GID:GECOS:HOME:SHELL)
 */
void yocto_write_passwd(FILE *f, const char *user, int uid, int gid, const char *gecos, const char *home, const char *shell) {
    if (f) fprintf(f, "%s:x:%d:%d:%s:%s:%s\n", user, uid, gid, gecos, home, shell);
}

/**
 * @brief Scrive una riga in formato shadow (USER:PASS:LAST:MIN:MAX:WARN:INACT:EXP:RES)
 */
void yocto_write_shadow(FILE *f, const char *user, const char *enc_pass) {
    // 19750 è un valore di last_change approssimativo per il 2024+
    if (f) fprintf(f, "%s:%s:19750:0:99999:7:::\n", user, enc_pass);
}

/**
 * @brief Scrive una riga in formato group (GROUP:x:GID:USERS)
 */
void yocto_write_group(FILE *f, const char *group, int gid, const char *users) {
    if (f) fprintf(f, "%s:x:%d:%s\n", group, gid, users ? users : "");
}

/**
 * @brief Filtra un file di testo (passwd/group) rimuovendo gli UID/GID umani
 */
int yocto_sanitize_file(const char *src_path, int min_id, int max_id) {
    char tmp_path[PATH_SAFE];
    snprintf(tmp_path, sizeof(tmp_path), "%s.tmp", src_path);

    FILE *src = fopen(src_path, "r");
    FILE *dst = fopen(tmp_path, "w");
    if (!src || !dst) return -1;

    char line[PATH_SAFE];
    while (fgets(line, sizeof(line), src)) {

        char line_copy[PATH_SAFE];
        strcpy(line_copy, line);

        strtok(line_copy, ":");           // Salta il nome (era 'name')
        strtok(NULL, ":");                // Salta la password (era 'pass')
        char *id_str = strtok(NULL, ":"); // Questo ci serve per l'ID

        if (id_str) {
            int id = atoi(id_str);
            // Se l'ID è fuori dal range umano (OE-Core), preserviamo la riga
            if (id < min_id || id > max_id) {
                fputs(line, dst);
            }
        }
    }

    fclose(src);
    fclose(dst);
    return rename(tmp_path, src_path);
}

/**
 * @brief Verifica se il percorso della home è in una whitelist di sistema.
 */
static bool is_path_allowed(const char *home) {
    if (home == NULL || strlen(home) < 2) {
        return false;
    }

    const char *whitelist[] = {"home", "opt", "srv", "usr", "var", NULL};
    char path_tmp[PATH_SAFE];
    
    strncpy(path_tmp, home, sizeof(path_tmp));
    path_tmp[sizeof(path_tmp) - 1] = '\0'; // Sicurezza extra per il terminatore

    char *fLevel = strtok(path_tmp, "/");
    if (fLevel == NULL) return false;

    for (int i = 0; whitelist[i] != NULL; i++) {
        if (strcmp(fLevel, whitelist[i]) == 0) {
            return true;
        }
    }

    return false;
}

/**
 * @brief yocto_is_human_user
 * Decide se un utente dell'host deve essere processato.
 */
bool yocto_is_human_user(uint32_t uid, const char *home) {
    // 1. Filtro UID basato su OE-Core (1000-59999)
    if (uid < OE_UID_HUMAN_MIN || uid > OE_UID_HUMAN_MAX) {
        return false;
    }

    // 2. Controllo Whitelist dei percorsi
    if (!is_path_allowed(home)) {
        return false;
    }

    // 3. Verifica fisica
    struct stat st;
    if (stat(home, &st) != 0 || !S_ISDIR(st.st_mode)) {
        return false;
    }

    // 4. Analisi sottocartelle vietate
    if (strstr(home, "/cache") || strstr(home, "/run") || strstr(home, "/spool")) {
        return false;
    }

    return true;
}