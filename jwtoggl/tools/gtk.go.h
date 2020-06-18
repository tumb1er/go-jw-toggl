#pragma once
#include <stdlib.h>

static const char* escape_markup(const char *string, int len) {
  return g_markup_escape_text(string, len);
}
