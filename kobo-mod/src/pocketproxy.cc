#include <cstddef>
#include <cstdlib>

#include <NickelHook.h>

static struct nh_info PocketProxy = {
    .name           = "PocketProxy",
    .desc           = "Intercept Pocket API HTTP calls and redirect them to configured URLs instead",
    .uninstall_flag = nullptr,
    .uninstall_xflag = "/mnt/onboard/.adds/pocket_proxy/DELETE_ME_TO_UNINSTALL",
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
