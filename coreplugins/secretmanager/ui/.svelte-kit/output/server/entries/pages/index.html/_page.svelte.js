import { e as escape_html, a as attr, d as derived, s as store_get, u as unsubscribe_stores, h as head } from "../../../chunks/index.js";
import "clsx";
import { w as writable } from "../../../chunks/index2.js";
function MessageBanner($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}
const URLPattern = {};
const baseLocale = "en";
const locales = (
  /** @type {const} */
  ["en", "zh"]
);
const localStorageKey = "arupa.language";
const strategy = [
  "localStorage",
  "preferredLanguage",
  "baseLocale"
];
const routeStrategies = [];
const isServer = typeof window === "undefined";
globalThis.__paraglide = /** @type {any} */
globalThis.__paraglide ?? {};
globalThis.__paraglide.ssr = /** @type {any} */
globalThis.__paraglide.ssr ?? {};
let localeInitiallySet = false;
let getLocale = () => {
  let strategyToUse = strategy;
  if (!isServer && typeof window !== "undefined" && window.location?.href) {
    strategyToUse = getStrategyForUrl(window.location.href);
  }
  const resolved = resolveLocaleWithStrategies(strategyToUse);
  if (resolved) {
    if (!localeInitiallySet) {
      localeInitiallySet = true;
      setLocale(resolved, { reload: false });
    }
    return resolved;
  }
  throw new Error("No locale found. Read the docs https://paraglidejs.com/errors#no-locale-found");
};
function resolveLocaleWithStrategies(strategyToUse, urlForUrlStrategy) {
  let locale2;
  for (const strat of strategyToUse) {
    if (strat === "baseLocale") {
      locale2 = baseLocale;
    } else if (strat === "preferredLanguage" && !isServer) {
      locale2 = extractLocaleFromNavigator();
    } else if (strat === "localStorage" && !isServer) {
      locale2 = localStorage.getItem(localStorageKey) ?? void 0;
    } else if (isCustomStrategy(strat) && customClientStrategies.has(strat)) {
      const handler = customClientStrategies.get(strat);
      if (handler) {
        const result = handler.getLocale();
        if (result instanceof Promise) {
          continue;
        }
        if (result !== void 0) {
          return assertIsLocale(result);
        }
      }
    }
    const matchedLocale = toLocale(locale2);
    if (matchedLocale) {
      return matchedLocale;
    }
  }
  return void 0;
}
const navigateOrReload = (newLocation) => {
  {
    window.location.reload();
  }
};
let setLocale = (newLocale, options) => {
  const optionsWithDefaults = {
    reload: true,
    ...options
  };
  let currentLocale;
  try {
    currentLocale = getLocale();
  } catch {
  }
  const customSetLocalePromises = [];
  let strategyToUse = strategy;
  if (!isServer && typeof window !== "undefined" && window.location?.href) {
    strategyToUse = getStrategyForUrl(window.location.href);
  }
  for (const strat of strategyToUse) {
    if (strat === "baseLocale") {
      continue;
    } else if (strat === "localStorage" && typeof window !== "undefined") {
      localStorage.setItem(localStorageKey, newLocale);
    } else if (isCustomStrategy(strat) && customClientStrategies.has(strat)) {
      const handler = customClientStrategies.get(strat);
      if (handler) {
        let result = handler.setLocale(newLocale);
        if (result instanceof Promise) {
          result = result.catch((error) => {
            throw new Error(`Custom strategy "${strat}" setLocale failed.`, {
              cause: error
            });
          });
          customSetLocalePromises.push(result);
        }
      }
    }
  }
  const runReload = () => {
    if (!isServer && optionsWithDefaults.reload && window.location && newLocale !== currentLocale) {
      navigateOrReload();
    }
  };
  if (customSetLocalePromises.length) {
    return Promise.all(customSetLocalePromises).then(() => {
      runReload();
    });
  }
  runReload();
  return;
};
let getUrlOrigin = () => {
  if (typeof window !== "undefined") {
    return window.location.origin;
  }
  return "http://fallback.com";
};
function toLocale(value2) {
  if (typeof value2 !== "string") {
    return void 0;
  }
  const lowerValue = value2.toLowerCase();
  for (const locale2 of locales) {
    if (locale2.toLowerCase() === lowerValue) {
      return locale2;
    }
  }
  return void 0;
}
function assertIsLocale(input) {
  const locale2 = toLocale(input);
  if (locale2)
    return locale2;
  throw new Error(`Invalid locale: ${input}. Expected one of: ${locales.join(", ")}`);
}
function extractLocaleFromNavigator() {
  if (!navigator?.languages?.length) {
    return void 0;
  }
  const languages = navigator.languages.map((lang) => ({
    fullTag: lang,
    baseTag: lang.split("-")[0]
  }));
  for (const lang of languages) {
    const fullLocale = toLocale(lang.fullTag);
    if (fullLocale) {
      return fullLocale;
    }
    const baseLocale2 = toLocale(lang.baseTag);
    if (baseLocale2) {
      return baseLocale2;
    }
  }
  return void 0;
}
function deLocalizeUrl(url) {
  {
    return deLocalizeUrlDefaultPattern(url);
  }
}
function deLocalizeUrlDefaultPattern(url) {
  const urlObj = typeof url === "string" ? new URL(url, getUrlOrigin()) : new URL(url);
  const pathSegments = urlObj.pathname.split("/").filter(Boolean);
  if (pathSegments.length > 0 && toLocale(pathSegments[0])) {
    urlObj.pathname = "/" + pathSegments.slice(1).join("/");
  }
  return urlObj;
}
let cachedRouteStrategyUrl;
let cachedRouteStrategy;
function findMatchingRouteStrategy(url) {
  if (routeStrategies.length === 0) {
    return void 0;
  }
  const urlString = typeof url === "string" ? url : url.href;
  if (cachedRouteStrategyUrl === urlString) {
    return cachedRouteStrategy;
  }
  const publicUrl = new URL(urlString, "http://example.com");
  const canonicalUrl = deLocalizeUrl(publicUrl);
  const candidateUrls = canonicalUrl.href === publicUrl.href ? [publicUrl] : [publicUrl, canonicalUrl];
  let match;
  for (const candidateUrl of candidateUrls) {
    for (const routeStrategy of routeStrategies) {
      const pattern = new URLPattern(routeStrategy.match, candidateUrl.href);
      if (pattern.exec(candidateUrl.href)) {
        match = routeStrategy;
        break;
      }
    }
    if (match)
      break;
  }
  cachedRouteStrategyUrl = urlString;
  cachedRouteStrategy = match;
  return match;
}
function getStrategyForUrl(url) {
  const routeStrategy = findMatchingRouteStrategy(url);
  if (routeStrategy && routeStrategy.exclude !== true && Array.isArray(routeStrategy.strategy)) {
    return routeStrategy.strategy;
  }
  return strategy;
}
const customClientStrategies = /* @__PURE__ */ new Map();
function isCustomStrategy(strategy2) {
  return typeof strategy2 === "string" && /^custom-[A-Za-z0-9_-]+$/.test(strategy2);
}
const en_page_title = (
  /** @type {(inputs: Page_TitleInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Secrets`
    );
  }
);
const zh_page_title = (
  /** @type {(inputs: Page_TitleInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `密钥`
    );
  }
);
const page_title = (
  /** @type {((inputs?: Page_TitleInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Page_TitleInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_page_title();
    return en_page_title();
  })
);
const en_page_description = (
  /** @type {(inputs: Page_DescriptionInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Centralized secrets manager`
    );
  }
);
const zh_page_description = (
  /** @type {(inputs: Page_DescriptionInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `集中式密钥管理器`
    );
  }
);
const page_description = (
  /** @type {((inputs?: Page_DescriptionInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Page_DescriptionInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_page_description();
    return en_page_description();
  })
);
const en_secrets = (
  /** @type {(inputs: SecretsInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Secrets`
    );
  }
);
const zh_secrets = (
  /** @type {(inputs: SecretsInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `密钥`
    );
  }
);
const secrets = (
  /** @type {((inputs?: SecretsInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<SecretsInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_secrets();
    return en_secrets();
  })
);
const en_refresh = (
  /** @type {(inputs: RefreshInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Refresh`
    );
  }
);
const zh_refresh = (
  /** @type {(inputs: RefreshInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `刷新`
    );
  }
);
const refresh = (
  /** @type {((inputs?: RefreshInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<RefreshInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_refresh();
    return en_refresh();
  })
);
const en_filter_by_name = (
  /** @type {(inputs: Filter_By_NameInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Filter by name`
    );
  }
);
const zh_filter_by_name = (
  /** @type {(inputs: Filter_By_NameInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `按名称筛选`
    );
  }
);
const filter_by_name = (
  /** @type {((inputs?: Filter_By_NameInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Filter_By_NameInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_filter_by_name();
    return en_filter_by_name();
  })
);
const en_loading = (
  /** @type {(inputs: LoadingInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Loading…`
    );
  }
);
const zh_loading = (
  /** @type {(inputs: LoadingInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `加载中…`
    );
  }
);
const loading = (
  /** @type {((inputs?: LoadingInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<LoadingInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_loading();
    return en_loading();
  })
);
const en_managed_secrets = (
  /** @type {(inputs: Managed_SecretsInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Managed secrets`
    );
  }
);
const zh_managed_secrets = (
  /** @type {(inputs: Managed_SecretsInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `已管理的密钥`
    );
  }
);
const managed_secrets = (
  /** @type {((inputs?: Managed_SecretsInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Managed_SecretsInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_managed_secrets();
    return en_managed_secrets();
  })
);
const en_add_secret = (
  /** @type {(inputs: Add_SecretInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Add secret`
    );
  }
);
const zh_add_secret = (
  /** @type {(inputs: Add_SecretInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `添加密钥`
    );
  }
);
const add_secret = (
  /** @type {((inputs?: Add_SecretInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Add_SecretInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_add_secret();
    return en_add_secret();
  })
);
const en_revealed_secret_title = (
  /** @type {(inputs: Revealed_Secret_TitleInputs) => LocalizedString} */
  (i) => {
    return (
      /** @type {LocalizedString} */
      `Secret: ${i?.name}`
    );
  }
);
const zh_revealed_secret_title = (
  /** @type {(inputs: Revealed_Secret_TitleInputs) => LocalizedString} */
  (i) => {
    return (
      /** @type {LocalizedString} */
      `密钥：${i?.name}`
    );
  }
);
const revealed_secret_title = (
  /** @type {((inputs: Revealed_Secret_TitleInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Revealed_Secret_TitleInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_revealed_secret_title(inputs);
    return en_revealed_secret_title(inputs);
  })
);
const en_http_warning = (
  /** @type {(inputs: Http_WarningInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `HTTP is not safe; this secret may be compromised.`
    );
  }
);
const zh_http_warning = (
  /** @type {(inputs: Http_WarningInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `当前使用 HTTP，传输不安全，此密钥可能泄露。`
    );
  }
);
const http_warning = (
  /** @type {((inputs?: Http_WarningInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Http_WarningInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_http_warning();
    return en_http_warning();
  })
);
const en_copy = (
  /** @type {(inputs: CopyInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Copy`
    );
  }
);
const zh_copy = (
  /** @type {(inputs: CopyInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `复制`
    );
  }
);
const copy = (
  /** @type {((inputs?: CopyInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<CopyInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_copy();
    return en_copy();
  })
);
const en_close = (
  /** @type {(inputs: CloseInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Close`
    );
  }
);
const zh_close = (
  /** @type {(inputs: CloseInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `关闭`
    );
  }
);
const close = (
  /** @type {((inputs?: CloseInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<CloseInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_close();
    return en_close();
  })
);
const en_close_dialog = (
  /** @type {(inputs: Close_DialogInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Close dialog`
    );
  }
);
const zh_close_dialog = (
  /** @type {(inputs: Close_DialogInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `关闭对话框`
    );
  }
);
const close_dialog = (
  /** @type {((inputs?: Close_DialogInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Close_DialogInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_close_dialog();
    return en_close_dialog();
  })
);
const en_name = (
  /** @type {(inputs: NameInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Name`
    );
  }
);
const zh_name = (
  /** @type {(inputs: NameInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `名称`
    );
  }
);
const name = (
  /** @type {((inputs?: NameInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<NameInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_name();
    return en_name();
  })
);
const en_unique_identifier = (
  /** @type {(inputs: Unique_IdentifierInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Unique identifier`
    );
  }
);
const zh_unique_identifier = (
  /** @type {(inputs: Unique_IdentifierInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `唯一标识符`
    );
  }
);
const unique_identifier = (
  /** @type {((inputs?: Unique_IdentifierInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Unique_IdentifierInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_unique_identifier();
    return en_unique_identifier();
  })
);
const en_description_optional = (
  /** @type {(inputs: Description_OptionalInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Description (optional)`
    );
  }
);
const zh_description_optional = (
  /** @type {(inputs: Description_OptionalInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `描述（可选）`
    );
  }
);
const description_optional = (
  /** @type {((inputs?: Description_OptionalInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Description_OptionalInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_description_optional();
    return en_description_optional();
  })
);
const en_description_placeholder = (
  /** @type {(inputs: Description_PlaceholderInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Set an alias or comment`
    );
  }
);
const zh_description_placeholder = (
  /** @type {(inputs: Description_PlaceholderInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `设置别名或备注`
    );
  }
);
const description_placeholder = (
  /** @type {((inputs?: Description_PlaceholderInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Description_PlaceholderInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_description_placeholder();
    return en_description_placeholder();
  })
);
const en_value = (
  /** @type {(inputs: ValueInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Value`
    );
  }
);
const zh_value = (
  /** @type {(inputs: ValueInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `值`
    );
  }
);
const value = (
  /** @type {((inputs?: ValueInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<ValueInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_value();
    return en_value();
  })
);
const en_secret_value = (
  /** @type {(inputs: Secret_ValueInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Secret value`
    );
  }
);
const zh_secret_value = (
  /** @type {(inputs: Secret_ValueInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `密钥值`
    );
  }
);
const secret_value = (
  /** @type {((inputs?: Secret_ValueInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Secret_ValueInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_secret_value();
    return en_secret_value();
  })
);
const en_allowed_plugins = (
  /** @type {(inputs: Allowed_PluginsInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Allowed plugins`
    );
  }
);
const zh_allowed_plugins = (
  /** @type {(inputs: Allowed_PluginsInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `允许访问的插件`
    );
  }
);
const allowed_plugins = (
  /** @type {((inputs?: Allowed_PluginsInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Allowed_PluginsInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_allowed_plugins();
    return en_allowed_plugins();
  })
);
const en_allowed_plugins_placeholder = (
  /** @type {(inputs: Allowed_Plugins_PlaceholderInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `one-plugin-per-line`
    );
  }
);
const zh_allowed_plugins_placeholder = (
  /** @type {(inputs: Allowed_Plugins_PlaceholderInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `每行一个插件`
    );
  }
);
const allowed_plugins_placeholder = (
  /** @type {((inputs?: Allowed_Plugins_PlaceholderInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Allowed_Plugins_PlaceholderInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_allowed_plugins_placeholder();
    return en_allowed_plugins_placeholder();
  })
);
const en_allowed_plugins_empty = (
  /** @type {(inputs: Allowed_Plugins_EmptyInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Empty means no plugin can access this secret.`
    );
  }
);
const zh_allowed_plugins_empty = (
  /** @type {(inputs: Allowed_Plugins_EmptyInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `留空表示不允许任何插件访问此密钥。`
    );
  }
);
const allowed_plugins_empty = (
  /** @type {((inputs?: Allowed_Plugins_EmptyInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Allowed_Plugins_EmptyInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_allowed_plugins_empty();
    return en_allowed_plugins_empty();
  })
);
const en_cancel = (
  /** @type {(inputs: CancelInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Cancel`
    );
  }
);
const zh_cancel = (
  /** @type {(inputs: CancelInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `取消`
    );
  }
);
const cancel = (
  /** @type {((inputs?: CancelInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<CancelInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_cancel();
    return en_cancel();
  })
);
const en_protection = (
  /** @type {(inputs: ProtectionInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Protection`
    );
  }
);
const zh_protection = (
  /** @type {(inputs: ProtectionInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `保护方式`
    );
  }
);
const protection = (
  /** @type {((inputs?: ProtectionInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<ProtectionInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_protection();
    return en_protection();
  })
);
const en_identity_protection = (
  /** @type {(inputs: Identity_ProtectionInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Manager identity`
    );
  }
);
const zh_identity_protection = (
  /** @type {(inputs: Identity_ProtectionInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `管理器身份保护`
    );
  }
);
const identity_protection = (
  /** @type {((inputs?: Identity_ProtectionInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Identity_ProtectionInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_identity_protection();
    return en_identity_protection();
  })
);
const en_passphrase_protection = (
  /** @type {(inputs: Passphrase_ProtectionInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `Passphrase protection`
    );
  }
);
const zh_passphrase_protection = (
  /** @type {(inputs: Passphrase_ProtectionInputs) => LocalizedString} */
  () => {
    return (
      /** @type {LocalizedString} */
      `口令保护`
    );
  }
);
const passphrase_protection = (
  /** @type {((inputs?: Passphrase_ProtectionInputs, options?: { locale?: "en" | "zh" }) => LocalizedString) & import('../runtime.js').MessageMetadata<Passphrase_ProtectionInputs, { locale?: "en" | "zh" }, {}>} */
  ((inputs = {}, options = {}) => {
    const locale2 = options.locale ?? getLocale();
    if (locale2 === "zh") return zh_passphrase_protection();
    return en_passphrase_protection();
  })
);
const locale = writable(getLocale());
function SecretForm($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    let { editing, busy } = $$props;
    let name$1 = "";
    let description = "";
    let value$1 = "";
    let protection$1 = "identity";
    let allowedPlugins = "";
    let localeOptions = derived(() => ({ locale: store_get($$store_subs ??= {}, "$locale", locale) }));
    let changingValue = derived(() => editing === null);
    let canSave = derived(() => editing === null);
    $$renderer2.push(`<form class="grid gap-4 p-4"><label class="grid gap-2"><span class="text-sm font-semibold text-base-content/70">${escape_html(name({}, localeOptions()))}</span> <input class="input w-full font-mono" required="" maxlength="128"${attr("placeholder", unique_identifier({}, localeOptions()))}${attr("disabled", busy, true)}${attr("value", name$1)}/></label> <label class="grid gap-2"><span class="text-sm font-semibold text-base-content/70">${escape_html(description_optional({}, localeOptions()))}</span> <input class="input w-full" maxlength="240"${attr("placeholder", description_placeholder({}, localeOptions()))}${attr("disabled", busy, true)}${attr("value", description)}/></label> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (changingValue()) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<label class="grid gap-2"><span class="text-sm font-semibold text-base-content/70">${escape_html(value({}, localeOptions()))}</span> <textarea class="textarea min-h-32 w-full font-mono text-sm" autocomplete="off"${attr("placeholder", secret_value({}, localeOptions()))}${attr("disabled", busy, true)}>`);
      const $$body = escape_html(value$1);
      if ($$body) {
        $$renderer2.push(`${$$body}`);
      }
      $$renderer2.push(`</textarea></label> <label class="grid gap-2"><span class="text-sm font-semibold text-base-content/70">${escape_html(protection({}, localeOptions()))}</span> `);
      $$renderer2.select({ class: "select w-full", disabled: busy, value: protection$1 }, ($$renderer3) => {
        $$renderer3.option({ value: "identity" }, ($$renderer4) => {
          $$renderer4.push(`${escape_html(identity_protection({}, localeOptions()))}`);
        });
        $$renderer3.option({ value: "passphrase" }, ($$renderer4) => {
          $$renderer4.push(`${escape_html(passphrase_protection({}, localeOptions()))}`);
        });
      });
      $$renderer2.push(`</label> `);
      {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]-->`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <label class="grid gap-2"><span class="text-sm font-semibold text-base-content/70">${escape_html(allowed_plugins({}, localeOptions()))}</span> <textarea class="textarea min-h-24 w-full font-mono text-sm"${attr("placeholder", allowed_plugins_placeholder({}, localeOptions()))}${attr("disabled", busy, true)}>`);
    const $$body_1 = escape_html(allowedPlugins);
    if ($$body_1) {
      $$renderer2.push(`${$$body_1}`);
    }
    $$renderer2.push(`</textarea> <span class="text-xs text-base-content/60">${escape_html(allowed_plugins_empty({}, localeOptions()))}</span></label> <div class="flex flex-wrap justify-end gap-2"><button class="btn" type="button"${attr("disabled", busy, true)}>${escape_html(cancel({}, localeOptions()))}</button> <button class="btn btn-primary" type="submit"${attr("disabled", busy || !canSave(), true)}>${escape_html(add_secret({}, localeOptions()))}</button></div></form>`);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}
function RevealDialog($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    let { title, value: value2, warning } = $$props;
    let localeOptions = derived(() => ({ locale: store_get($$store_subs ??= {}, "$locale", locale) }));
    $$renderer2.push(`<dialog class="modal"><div class="modal-box max-w-2xl"><h3 class="text-lg font-semibold">${escape_html(title)}</h3> <p class="mt-2 text-sm text-warning">${escape_html(warning)}</p> <textarea class="textarea mt-4 min-h-40 w-full select-all font-mono text-sm" readonly="">`);
    const $$body = escape_html(value2);
    if ($$body) {
      $$renderer2.push(`${$$body}`);
    }
    $$renderer2.push(`</textarea> <div class="modal-action"><button class="btn" type="button">${escape_html(copy({}, localeOptions()))}</button> <button class="btn btn-primary" type="button">${escape_html(close({}, localeOptions()))}</button></div></div> <form method="dialog" class="modal-backdrop"><button${attr("aria-label", close_dialog({}, localeOptions()))}>${escape_html(close({}, localeOptions()))}</button></form></dialog>`);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    let keys = [];
    let editing = null;
    let filterQuery = "";
    let busyCount = 0;
    let busy = derived(() => busyCount > 0);
    let revealName = "";
    let revealValue = "";
    let localeOptions = derived(() => ({ locale: store_get($$store_subs ??= {}, "$locale", locale) }));
    head("py8068", $$renderer2, ($$renderer3) => {
      $$renderer3.title(($$renderer4) => {
        $$renderer4.push(`<title>${escape_html(page_title({}, localeOptions()))}</title>`);
      });
      $$renderer3.push(`<meta name="description"${attr("content", page_description({}, localeOptions()))}/>`);
    });
    $$renderer2.push(`<main class="mx-auto grid w-full max-w-[1320px] gap-4 px-4 py-6 sm:py-8"><header class="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"><div class="flex min-w-0 items-center gap-3"><div class="min-w-0"><h1 class="text-2xl font-semibold">${escape_html(secrets({}, localeOptions()))}</h1> <p class="truncate text-sm text-base-content/60">${escape_html(page_description({}, localeOptions()))}</p></div></div> <div class="flex flex-wrap gap-2"><button class="btn btn-primary" type="button"${attr("disabled", busy(), true)}>${escape_html(refresh({}, localeOptions()))}</button></div></header> `);
    MessageBanner($$renderer2);
    $$renderer2.push(`<!----> <section class="grid min-w-0 gap-4 lg:grid-cols-[minmax(0,1fr)_360px]"><div class="card min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm"><div class="flex flex-col gap-3 border-b border-base-300 p-4 sm:flex-row sm:items-center sm:justify-between"><h2 class="font-semibold">${escape_html(secrets({}, localeOptions()))}</h2> <input class="input w-full sm:w-64" type="search"${attr("placeholder", filter_by_name({}, localeOptions()))}${attr("disabled", busy(), true)}${attr("value", filterQuery)}/></div> <div class="min-w-0 overflow-x-auto">`);
    {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="grid min-h-48 place-items-center p-6 text-base-content/60">${escape_html(loading({}, localeOptions()))}</div>`);
    }
    $$renderer2.push(`<!--]--></div></div> <div class="grid gap-4"><div class="stat rounded-lg border border-base-300 bg-base-100 shadow-sm"><div class="stat-title">${escape_html(managed_secrets({}, localeOptions()))}</div> <div class="stat-value text-2xl">${escape_html(keys.length)}</div></div> <aside class="card h-fit min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm"><div class="border-b border-base-300 p-4"><h2 class="font-semibold">${escape_html(add_secret({}, localeOptions()))}</h2></div> `);
    SecretForm($$renderer2, {
      editing,
      busy: busy()
    });
    $$renderer2.push(`<!----></aside></div></section></main> `);
    RevealDialog($$renderer2, {
      title: revealed_secret_title({ name: revealName }, localeOptions()),
      value: revealValue,
      warning: http_warning({}, localeOptions())
    });
    $$renderer2.push(`<!---->`);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}
export {
  _page as default
};
