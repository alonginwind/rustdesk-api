'use strict';
const MANIFEST = 'flutter-app-manifest';
const TEMP = 'flutter-temp-cache';
const CACHE_NAME = 'flutter-app-cache';

const RESOURCES = {"index.html": "d08aafa3ea873fc431d780aca9a82538",
"/": "d08aafa3ea873fc431d780aca9a82538",
"canvaskit/canvaskit.js.symbols": "48c83a2ce573d9692e8d970e288d75f7",
"canvaskit/chromium/canvaskit.js.symbols": "a012ed99ccba193cf96bb2643003f6fc",
"canvaskit/chromium/canvaskit.wasm": "b1ac05b29c127d86df4bcfbf50dd902a",
"canvaskit/chromium/canvaskit.js": "671c6b4f8fcc199dcc551c7bb125f239",
"canvaskit/skwasm.worker.js": "89990e8c92bcb123999aa81f7e203b1c",
"canvaskit/skwasm.wasm": "9f0c0c02b82a910d12ce0543ec130e60",
"canvaskit/canvaskit.wasm": "1f237a213d7370cf95f443d896176460",
"canvaskit/canvaskit.js": "66177750aff65a66cb07bb44b8c6422b",
"canvaskit/skwasm.js": "694fda5704053957c2594de355805228",
"canvaskit/skwasm.js.symbols": "262f4827a1317abb59d71d6c587a93e2",
"icons/Icon-512.png": "96e752610906ba2a93c65f8abe1645f1",
"icons/Icon-192.png": "ac9a721a12bbc803b44f645561ecb1e1",
"icons/Icon-maskable-512.png": "301a7604d45b3e739efc881eb04896ea",
"icons/Icon-maskable-192.png": "c457ef57daa1d16f64b27b786ec2ea3c",
"pkg/web_bridge.js": "6f9706cbf910ec4e932d8627b63dfca1",
"pkg/web_bridge_bg.wasm": "275e59670de3b5201f560f9adb61002a",
"load_bridge.js": "95701bfdf1d3e3f4d70e20d971008d9a",
"load_bridge.tpl.js": "55c157aa0714f94af3181a76673fe171",
"tweetnacl.min.js": "f5d2adfc82ea03ecff05df1b241181b9",
"manifest.json": "4b5dfbb08964d9dba0a0ccef7efa2ed3",
"favicon.png": "5dcef449791fa27946b3d35ad8803796",
"flutter.js": "f393d3c16b631f36852323de8e583132",
"version.json": "b20ebee53792b779b0eb92bb0bc77ae8",
"fzstd.js": "8ddb24fbb64f4f4c05ddf22db97149c3",
"share.js": "9fc4672b4910371afa2e2163a54c6293",
"flutter_bootstrap.js": "a98182edde978afaad658a4b2c08ccc5",
"assets/AssetManifest.json": "5b9b4b29531b22bb35af66f6a5bbea7b",
"assets/AssetManifest.bin": "917644686808dc844d51a83e4d00a519",
"assets/FontManifest.json": "f4f31d18ca443c7bfb50340e84b860d5",
"assets/packages/wakelock_plus/assets/no_sleep.js": "7748a45cd593f33280669b29c2c8919a",
"assets/packages/flex_color_picker/assets/opacity.png": "49c4f3bcb1b25364bb4c255edcaaf5b2",
"assets/packages/dash_chat_2/assets/placeholder.png": "ce1fece6c831b69b75c6c25a60b5b0f3",
"assets/packages/dash_chat_2/assets/profile_placeholder.png": "77f5794e2eb49f7989b8f85e92cfa4e0",
"assets/packages/window_manager/images/ic_chrome_minimize.png": "4282cd84cb36edf2efb950ad9269ca62",
"assets/packages/window_manager/images/ic_chrome_close.png": "75f4b8ab3608a05461a31fc18d6b47c2",
"assets/packages/window_manager/images/ic_chrome_unmaximize.png": "4a90c1909cb74e8f0d35794e2f61d8bf",
"assets/packages/window_manager/images/ic_chrome_maximize.png": "af7499d7657c8b69d23b85156b60298c",
"assets/shaders/ink_sparkle.frag": "ecc85a2e95f5e9f53123dcaf8cb9b6ce",
"assets/AssetManifest.bin.json": "a7e320ec42d2cd3ce17f1b55689251b8",
"assets/NOTICES": "223308dfb930b3c3fd76ff1f23d79097",
"assets/fonts/MaterialIcons-Regular.otf": "fa8584739cafdd75934ac77528c5a4ce",
"assets/assets/kb_layout_not_iso.svg": "c0903bc16f7a3584677d1129c627f79e",
"assets/assets/auth-auth0.svg": "563377c91144cb7d1a02ff8ee5aa58bb",
"assets/assets/actions_mobile.svg": "840555746fccec5610779500eec942be",
"assets/assets/auth-facebook.svg": "ee6b89b302080bcef63b9779b6c9bea4",
"assets/assets/search.svg": "492f2847532faf43169300c524aa7ccb",
"assets/assets/display_switcher.svg": "e0fb2ce351b287e3adf5df93224c4f49",
"assets/assets/peer_searchbar.ttf": "f189a84ff155487e21892166a8a1a49b",
"assets/assets/address_book.ttf": "612eb0515c3bca0ea7e661cb74c14fcc",
"assets/assets/auth-okta.svg": "2f417879cb6ff130f0b6294c6710fa72",
"assets/assets/unpinned.svg": "d529613e4ab67bc9049861fcd8646606",
"assets/assets/auth-microsoft.svg": "22f8f6cc1c8e024fe205d3f56f2be58d",
"assets/assets/android.svg": "2172ec92c0da8b72dfe31565d5cee45c",
"assets/assets/more.ttf": "d1d2c5484a00438591c383e6ff514747",
"assets/assets/close.svg": "e2f1a0bd1efe55a14fc21e70d14265e8",
"assets/assets/checkbox-outline.svg": "f0d7b636853657cc21df676e2f473e1f",
"assets/assets/auth-github.svg": "3c30cd99eb7418fe737b717536672201",
"assets/assets/mac.svg": "aea75d88e94e3feda3a8e85fe8f7e83b",
"assets/assets/folder_new.svg": "bf774d77fd7614a7d3f26f98f3e5d8d6",
"assets/assets/message_24dp_5F6368.svg": "b0f463bd4acc2aa9ead75c7e652a0543",
"assets/assets/tabbar.ttf": "593f286bbe900c64016ed23dc8ba91d6",
"assets/assets/chat2.svg": "3cb54943f6e44b5edf015a6e80955d43",
"assets/assets/insecure_relay.svg": "459680c84934bd3e0ed583df62617080",
"assets/assets/dots.svg": "10dc01da59bbf56899fe3dfb24de4588",
"assets/assets/win.svg": "e6032ece6ba082b1ab125503577d9525",
"assets/assets/scam.png": "a6d290cead819b6bfe4248b5d542567f",
"assets/assets/arrow.svg": "ef5695b014466fc2f60dc3a201fa9435",
"assets/assets/call_end.svg": "82190401a712a19798a416252071b2ae",
"assets/assets/auth-google.svg": "26dd5d10ffb0d02eb37a3858f2e533f3",
"assets/assets/record_screen.svg": "ccb57d890c59dad9e4b7d22422ed9b21",
"assets/assets/transfer.svg": "bde9d26182e2684ae721ecb5a55963fc",
"assets/assets/trash.svg": "32cb1ca9e5561cc9d545e185d8bf17f1",
"assets/assets/file_transfer.svg": "7530de7176b77ed19834c6dd87a692c4",
"assets/assets/actions.svg": "2a43efdc71e95b7db43c1fc5ae34ed87",
"assets/assets/secure.svg": "bf1b40b1a434039d84856285af6708ed",
"assets/assets/linux.svg": "94446987e0c95123c5549fb26affdd32",
"assets/assets/chat.svg": "10595dcc8484d6f7f140f54d6db32877",
"assets/assets/secure_relay.svg": "cd3fe0afb93a1978651efc96f723b203",
"assets/assets/pinned.svg": "0a2acd10ced63f0224be16aa8c0f3b5f",
"assets/assets/fullscreen.svg": "ddb2613848048e0b1ac089eacd72b8b6",
"assets/assets/refresh.svg": "1d14356f6aec82f5708c1d7bc3538ff4",
"assets/assets/auth-default.svg": "6c1a2702edb372085ca4371c4ef09aa5",
"assets/assets/display.svg": "bf281b8b99eac956f7ab03028e7c235b",
"assets/assets/insecure.svg": "58385c3079d8d21b891582151147a144",
"assets/assets/keyboard_mouse.svg": "df5d9c69c2ecc23735f4919b416d19d0",
"assets/assets/folder.svg": "eeaa976f375cc066c7fb7cc99733cd30",
"assets/assets/rec.svg": "8cbc54b941a213eee9017f723787b3c0",
"assets/assets/chevron_up_chevron_down.svg": "9ca13e04aa9ce89f591d775461e827d2",
"assets/assets/auth-apple.svg": "f96f69a4e4e7f5415d0888e249198b8a",
"assets/assets/file.svg": "8e000fe8e15b1ecb0c88d46ee64f6265",
"assets/assets/device_group.ttf": "0c14da6ee5eab333bfacf34943d68887",
"assets/assets/home.svg": "2320ca704fa8dfa8fadedd4c181b1d34",
"assets/assets/fullscreen_exit.svg": "e9adde63a9a4613ce646c2b68a81554a",
"assets/assets/voice_call_waiting.svg": "cbbab0ecd11a1c20ca16ebd41e72d86a",
"assets/assets/icon.svg": "9673d0a1dd44d81bc31c76a56857d787",
"assets/assets/call_wait.svg": "ca0cdeaa2a498f901db4973419f56e49",
"assets/assets/kb_layout_iso.svg": "19463c9f2c8c2ce6db8a57d2bcf206ac",
"assets/assets/gestures.ttf": "a70c60208ba07ce378ed9a5cf8aa586b",
"assets/assets/voice_call.svg": "b261a6c0ce34fd6a31bba07f8e32f5fa",
"assets/assets/auth-gitlab.svg": "b9b4d037f50930335a0e503cd9a35294",
"assets/assets/screen.svg": "c6b0ea973c050844c2c2cba27eaffd23",
"main.dart.js": "d9dcf8ab28d11f53f7023e05b7d71813"};
// The application shell files that are downloaded before a service worker can
// start.
const CORE = ["main.dart.js",
"index.html",
"flutter_bootstrap.js",
"assets/AssetManifest.bin.json",
"assets/FontManifest.json"];

// During install, the TEMP cache is populated with the application shell files.
self.addEventListener("install", (event) => {
  self.skipWaiting();
  return event.waitUntil(
    caches.open(TEMP).then((cache) => {
      return cache.addAll(
        CORE.map((value) => new Request(value, {'cache': 'reload'})));
    })
  );
});
// During activate, the cache is populated with the temp files downloaded in
// install. If this service worker is upgrading from one with a saved
// MANIFEST, then use this to retain unchanged resource files.
self.addEventListener("activate", function(event) {
  return event.waitUntil(async function() {
    try {
      var contentCache = await caches.open(CACHE_NAME);
      var tempCache = await caches.open(TEMP);
      var manifestCache = await caches.open(MANIFEST);
      var manifest = await manifestCache.match('manifest');
      // When there is no prior manifest, clear the entire cache.
      if (!manifest) {
        await caches.delete(CACHE_NAME);
        contentCache = await caches.open(CACHE_NAME);
        for (var request of await tempCache.keys()) {
          var response = await tempCache.match(request);
          await contentCache.put(request, response);
        }
        await caches.delete(TEMP);
        // Save the manifest to make future upgrades efficient.
        await manifestCache.put('manifest', new Response(JSON.stringify(RESOURCES)));
        // Claim client to enable caching on first launch
        self.clients.claim();
        return;
      }
      var oldManifest = await manifest.json();
      var origin = self.location.origin;
      for (var request of await contentCache.keys()) {
        var key = request.url.substring(origin.length + 1);
        if (key == "") {
          key = "/";
        }
        // If a resource from the old manifest is not in the new cache, or if
        // the MD5 sum has changed, delete it. Otherwise the resource is left
        // in the cache and can be reused by the new service worker.
        if (!RESOURCES[key] || RESOURCES[key] != oldManifest[key]) {
          await contentCache.delete(request);
        }
      }
      // Populate the cache with the app shell TEMP files, potentially overwriting
      // cache files preserved above.
      for (var request of await tempCache.keys()) {
        var response = await tempCache.match(request);
        await contentCache.put(request, response);
      }
      await caches.delete(TEMP);
      // Save the manifest to make future upgrades efficient.
      await manifestCache.put('manifest', new Response(JSON.stringify(RESOURCES)));
      // Claim client to enable caching on first launch
      self.clients.claim();
      return;
    } catch (err) {
      // On an unhandled exception the state of the cache cannot be guaranteed.
      console.error('Failed to upgrade service worker: ' + err);
      await caches.delete(CACHE_NAME);
      await caches.delete(TEMP);
      await caches.delete(MANIFEST);
    }
  }());
});
// The fetch handler redirects requests for RESOURCE files to the service
// worker cache.
self.addEventListener("fetch", (event) => {
  if (event.request.method !== 'GET') {
    return;
  }
  var origin = self.location.origin;
  var key = event.request.url.substring(origin.length + 1);
  // Redirect URLs to the index.html
  if (key.indexOf('?v=') != -1) {
    key = key.split('?v=')[0];
  }
  if (event.request.url == origin || event.request.url.startsWith(origin + '/#') || key == '') {
    key = '/';
  }
  // If the URL is not the RESOURCE list then return to signal that the
  // browser should take over.
  if (!RESOURCES[key]) {
    return;
  }
  // If the URL is the index.html, perform an online-first request.
  if (key == '/') {
    return onlineFirst(event);
  }
  event.respondWith(caches.open(CACHE_NAME)
    .then((cache) =>  {
      return cache.match(event.request).then((response) => {
        // Either respond with the cached resource, or perform a fetch and
        // lazily populate the cache only if the resource was successfully fetched.
        return response || fetch(event.request).then((response) => {
          if (response && Boolean(response.ok)) {
            cache.put(event.request, response.clone());
          }
          return response;
        });
      })
    })
  );
});
self.addEventListener('message', (event) => {
  // SkipWaiting can be used to immediately activate a waiting service worker.
  // This will also require a page refresh triggered by the main worker.
  if (event.data === 'skipWaiting') {
    self.skipWaiting();
    return;
  }
  if (event.data === 'downloadOffline') {
    downloadOffline();
    return;
  }
});
// Download offline will check the RESOURCES for all files not in the cache
// and populate them.
async function downloadOffline() {
  var resources = [];
  var contentCache = await caches.open(CACHE_NAME);
  var currentContent = {};
  for (var request of await contentCache.keys()) {
    var key = request.url.substring(origin.length + 1);
    if (key == "") {
      key = "/";
    }
    currentContent[key] = true;
  }
  for (var resourceKey of Object.keys(RESOURCES)) {
    if (!currentContent[resourceKey]) {
      resources.push(resourceKey);
    }
  }
  return contentCache.addAll(resources);
}
// Attempt to download the resource online before falling back to
// the offline cache.
function onlineFirst(event) {
  return event.respondWith(
    fetch(event.request).then((response) => {
      return caches.open(CACHE_NAME).then((cache) => {
        cache.put(event.request, response.clone());
        return response;
      });
    }).catch((error) => {
      return caches.open(CACHE_NAME).then((cache) => {
        return cache.match(event.request).then((response) => {
          if (response != null) {
            return response;
          }
          throw error;
        });
      });
    })
  );
}
