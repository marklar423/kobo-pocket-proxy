#include <NickelHook.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include <QString>
#include <QUrl>
#include <QtNetwork/QNetworkAccessManager>
#include <QtNetwork/QNetworkReply>
#include <QtNetwork/QNetworkRequest>
#include <cstddef>
#include <cstdlib>

#include "../inih/ini.h"

typedef struct WebResponseInflater_s WebResponseInflater;

static void (*originalMakeRequest)(QUrl const&, QString const&,
                                   QMap<QString, QString> const&,
                                   QByteArray const&, WebResponseInflater*, int,
                                   int, QNetworkRequest::CacheLoadControl);

static struct nh_info PocketProxy = {
    .name = "PocketProxy",
    .desc =
        "Intercept Pocket API HTTP calls and redirect them to configured URLs "
        "instead",
    .uninstall_flag = nullptr,
    .uninstall_xflag = "/mnt/onboard/.adds/pocket_proxy/DELETE_ME_TO_UNINSTALL",
};

static bool LoadedConfig = false;
static QString GetSendApiHostPort = "";
static QString TextApiHostPort = "";
const char kConfigFilePath[] =
    "/mnt/onboard/.adds/pocket_proxy/pocket_proxy.conf";

static int config_file_callback(void* /*user*/, const char* section,
                                const char* name, const char* value) {
  if (strcmp(section, "PocketProxy") == 0) {
    if (strcmp(name, "GetSendApiHostPort") == 0) {
      nh_log("GetSendApiHostPort=%s", value);
      GetSendApiHostPort = value;
    } else if (strcmp(name, "TextApiHostPort") == 0) {
      nh_log("TextApiHostPort=%s", value);
      TextApiHostPort = value;
    }
  } else {
    nh_log("Unknown section [%s]", section);
  }

  return 1;
}

static void read_config_file() {
  nh_log("Loading config;");

  // Load the config.
  FILE* file = fopen(kConfigFilePath, "r");
  if (!file) {
    nh_log("Failed to open config file: %s.", strerror(errno));
    return;
  }
  int error_line = ini_parse_file(file, config_file_callback, NULL);
  fclose(file);

  if (error_line != 0) {
    nh_log("Failed to parse config file: error on line %d.", error_line);
  }
}

static int init_proxy() {
  return 0;  //
}

extern "C" __attribute__((visibility("default"))) void _proxy_pocket_api_calls(
    QUrl const& url, QString const& param, QMap<QString, QString> headers,
    QByteArray const& output, WebResponseInflater* inflater, int param1,
    int param2, QNetworkRequest::CacheLoadControl cl) {
  if (!LoadedConfig) {
    LoadedConfig = true;
    read_config_file();

    if (GetSendApiHostPort.isEmpty() || TextApiHostPort.isEmpty()) {
      nh_log(
          "Either GetSendHostPort or TextApiHostPort is empty in "
          "pocket_proxy.conf. Proxying will not occur.");
    }
  }

  if (!GetSendApiHostPort.isEmpty() && url.host() == "getpocket.com") {
    QUrl replacement_url(GetSendApiHostPort);
    replacement_url.setPath(url.path());
    originalMakeRequest(replacement_url, param, headers, output, inflater,
                        param1, param2, cl);
    return;
  }

  if (!TextApiHostPort.isEmpty() && url.host() == "text.getpocket.com") {
    QUrl replacement_url(TextApiHostPort);
    replacement_url.setPath(url.path());
    originalMakeRequest(replacement_url, param, headers, output, inflater,
                        param1, param2, cl);
    return;
  }

  originalMakeRequest(url, param, headers, output, inflater, param1, param2,
                      cl);
}

static struct nh_hook PocketProxyHook[] = {
    {
        .sym =
            "_ZN12WebRequester11makeRequestERK4QUrlRK7QStringRK4QMapIS3_S3_"
            "ERK10QByteArrayP19WebResponseInflateriiN15QNetworkRequest16CacheLo"
            "adControlE",
        .sym_new = "_proxy_pocket_api_calls",
        .lib = "libnickel.so.1.0.0",
        .out = nh_symoutptr(originalMakeRequest),
        .desc = "calls to metadata",
        .optional = true,
    },
    {0},
};

static struct nh_dlsym PocketProxyDlsym[] = {
    {0},
};

NickelHook(                        //
        .init = &init_proxy,       //
        .info = &PocketProxy,      //
        .hook = PocketProxyHook,   //
        .dlsym = PocketProxyDlsym  //
);
