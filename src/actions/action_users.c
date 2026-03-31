/*
 * src/actions/action_users.c
 * Remastering core: User & Group Identity artisan
 * oa: eggs in my dialect🥚🥚
 *
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * License: GPL-3.0-or-later
 */
#include "oa.h"

int action_users(OA_Context *ctx) {
    // 1. Lookup a cascata (percorso, utenti, modalità)
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    if (!pathLiveFs) pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    cJSON *users = cJSON_GetObjectItemCaseSensitive(ctx->task, "users");
    if (!users) users = cJSON_GetObjectItemCaseSensitive(ctx->root, "users");
    cJSON *mode_item = cJSON_GetObjectItemCaseSensitive(ctx->task, "mode");
    if (!mode_item) mode_item = cJSON_GetObjectItemCaseSensitive(ctx->root, "mode");

    if (!cJSON_IsString(pathLiveFs)) return 1;

    char liveroot[PATH_SAFE], p_path[PATH_SAFE], s_path[PATH_SAFE], g_path[PATH_SAFE];
    snprintf(liveroot, sizeof(liveroot), "%s/liveroot", pathLiveFs->valuestring);
    snprintf(p_path, sizeof(p_path), "%s/etc/passwd", liveroot);
    snprintf(s_path, sizeof(s_path), "%s/etc/shadow", liveroot);
    snprintf(g_path, sizeof(g_path), "%s/etc/group", liveroot);

    const char *mode = cJSON_IsString(mode_item) ? mode_item->valuestring : "standard";

    // 2. PULIZIA: Rimuoviamo gli utenti host (UID/GID 1000-59999)
    if (strcmp(mode, "clone") != 0) {
        printf("\033[1;34m[oa]\033[0m Purging host identities...\n");
        yocto_sanitize_file(p_path, OE_UID_HUMAN_MIN, OE_UID_HUMAN_MAX); // passwd
        // yocto_sanitize_file(s_path, OE_UID_HUMAN_MIN, OE_UID_HUMAN_MAX); // shadow
        yocto_sanitize_file(g_path, OE_UID_HUMAN_MIN, OE_UID_HUMAN_MAX); // group
    }

    // 3. SCRITTURA: Creazione identità live tramite le tue funzioni vendors
    if (cJSON_IsArray(users)) {
        FILE *fp = fopen(p_path, "a");
        FILE *fs = fopen(s_path, "a");
        if (!fp || !fs) { if(fp) fclose(fp); if(fs) fclose(fs); return 1; }

        cJSON *u;
        cJSON_ArrayForEach(u, users) {
            cJSON *login_obj = cJSON_GetObjectItemCaseSensitive(u, "login");
            cJSON *pass_obj = cJSON_GetObjectItemCaseSensitive(u, "password");
            cJSON *home_obj = cJSON_GetObjectItemCaseSensitive(u, "home");
            cJSON *shell_obj = cJSON_GetObjectItemCaseSensitive(u, "shell");
            cJSON *gecos_obj = cJSON_GetObjectItemCaseSensitive(u, "gecos");

            if (!login_obj || !pass_obj || !home_obj) continue;

            const char *login = login_obj->valuestring;
            const char *pass = pass_obj->valuestring;
            const char *home = home_obj->valuestring;
            const char *shell = shell_obj ? shell_obj->valuestring : "/bin/bash";
            const char *gecos = gecos_obj ? gecos_obj->valuestring : "live,,,";

            printf("\033[1;32m[oa]\033[0m Handcrafting identity: %s\n", login);

            yocto_write_passwd(fp, login, 1000, 1000, gecos, home, shell);
            yocto_write_shadow(fs, login, pass);

            char home_cmd[CMD_MAX];
            snprintf(home_cmd, sizeof(home_cmd), "mkdir -p %s%s && chown 1000:1000 %s%s", 
                     liveroot, home, liveroot, home);
            system(home_cmd);
        }
        fclose(fp);
        fclose(fs);
    }

    printf("{\"status\": \"ok\", \"action\": \"users_complete\"}\n");
    return 0;
}
