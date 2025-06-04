#include <NickelHook.h>

#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QSettings>
#include <QString>
#include <QUrl>
#include <cstddef>
#include <cstdlib>

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

static QString GetSendHostPort = "";
static QString TextApiHostPort = "";

static int init_proxy() {
  // Load the config.
  QSettings settings("/mnt/onboard/.adds/pocket_proxy/pocket_proxy.conf",
                     QSettings::IniFormat);
  settings.beginGroup("PocketProxy");
  GetSendHostPort = settings.value("GetSendHostPort", "").toString();
  TextApiHostPort = settings.value("TextApiHostPort", "").toString();
  return 0;
}

extern "C" __attribute__((visibility("default"))) void _proxy_pocket_api_calls(
    QUrl const& url, QString const& param, QMap<QString, QString> headers,
    QByteArray const& output, WebResponseInflater* inflater, int param1,
    int param2, QNetworkRequest::CacheLoadControl cl) {
  if (GetSendHostPort.isEmpty() || TextApiHostPort.isEmpty()) {
    nh_log(
        "Either GetSendHostPort or TextApiHostPort is empty in "
        "pocket_proxy.conf. Proxying will not occur.");
    originalMakeRequest(url, param, headers, output, inflater, param1, param2,
                        cl);
    return;
  }

  if (url.host() == "getpocket.com") {
    nh_log("Redirecting Pocket Call to get/send endpoint specified in config.");
    QUrl replacement_url(GetSendHostPort);
    replacement_url.setPath(url.path());
    originalMakeRequest(replacement_url, param, headers, output, inflater,
                        param1, param2, cl);
    return;
  }

  if (url.host() == "text.getpocket.com") {
    nh_log("Redirecting Pocket Call to text endpoint specified in config.");
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

NickelHook(.init = &init_proxy, .info = &PocketProxy, .hook = PocketProxyHook,
           .dlsym = PocketProxyDlsym, );
