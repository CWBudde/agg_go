//go:build agogo && cgo

#include "bridge.h"

#include <string>

namespace {

std::string& last_error_storage() {
  static std::string value = "stub bridge: no AGG-backed implementation linked";
  return value;
}

}  // namespace

extern "C" int agg_go_cpp_bridge_probe(void) {
  last_error_storage() = "stub bridge: no AGG-backed implementation linked";
  return -1;
}

extern "C" int agg_go_cpp_bridge_is_stub(void) { return 1; }

extern "C" const char* agg_go_cpp_bridge_build_id(void) {
  return "agogo-stub-v1";
}

extern "C" const char* agg_go_cpp_bridge_last_error(void) {
  return last_error_storage().c_str();
}
