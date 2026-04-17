#ifndef AGG_GO_INTERNAL_CPPBRIDGE_BRIDGE_H
#define AGG_GO_INTERNAL_CPPBRIDGE_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

int agg_go_cpp_bridge_probe(void);
int agg_go_cpp_bridge_is_stub(void);
const char* agg_go_cpp_bridge_build_id(void);
const char* agg_go_cpp_bridge_last_error(void);

#ifdef __cplusplus
}
#endif

#endif
