#ifndef SNIPBOX_ENGINE_H
#define SNIPBOX_ENGINE_H
#include <stdint.h>
#ifdef __cplusplus
extern "C" {
#endif
/* Returned strings are owned by the engine; release them exactly once. */
char *snipbox_search(const char *commands_json, const char *query);
char *snipbox_parse_template(const char *command_json);
char *snipbox_render_template(const char *command, const char *values_json);
void snipbox_free_string(char *value);
#ifdef __cplusplus
}
#endif
#endif
