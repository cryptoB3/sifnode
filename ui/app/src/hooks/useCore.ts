import {
  createStore,
  createApi,
  createActions,
  createPoolFinder,
  getConfig,
} from "ui-core";

const config = getConfig(
  process.env.VUE_APP_DEPLOYMENT_TAG,
  process.env.VUE_APP_SIFCHAIN_ASSET_TAG,
  process.env.VUE_APP_ETHEREUM_ASSET_TAG
);

const api = createApi(config);
const store = createStore();
const actions = createActions({ store, api });
const poolFinder = createPoolFinder(store);

// expose store on the window so it is easy to inspect
Object.defineProperty(window, "store", {
  get: function() {
    return JSON.parse(
      JSON.stringify(store, function replacer(key, value) {
        if (value.amount && value.quotient) {
          return value.toString();
        }
        return value;
      })
    );
  },
});

export function useCore() {
  return {
    store,
    api,
    actions,
    poolFinder,
    config,
  };
}
