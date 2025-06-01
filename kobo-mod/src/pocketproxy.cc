#include <cstddef>
#include <cstdlib>

#include <NickelHook.h>

static struct nh_info PocketProxy = {
    .name           = "PocketProxy",
    .desc           = "",
    .uninstall_flag = "/mnt/onboard/pp_uninstall",
};

static struct nh_hook PocketProxyHook[] = {
    {0},
};

static struct nh_dlsym PocketProxyDlsym[] = {
    {0},
};

NickelHook(
    .init  = nullptr,
    .info  = &PocketProxy,
    .hook  = PocketProxyHook,
    .dlsym = PocketProxyDlsym,
)
