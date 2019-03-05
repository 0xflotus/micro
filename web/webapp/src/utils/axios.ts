import axios from "axios";
import config from "@/config";

import * as TYPES from "@/store/mutation-types";
import store from "@/store/store";

const baseURL = config.url.basicUrl;
const $axios = axios.create({
  baseURL,
  withCredentials: true, // 允许携带cookie
  timeout: 10000 // 超时时间
});

var rememberMeToken =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWJqZWN0SWQiOiIxMDAwMDkiLCJvcGVuSWQiOiJvaHZfMjFRNFR3N3JfcC1FOWp6MTN2QU5kbEdzIiwibmFtZSI6IkxpdHRsZSBGaXJzdCIsImV4cCI6MTU0MzA3NjkzMywianRpIjoiMTAwMDA5IiwiaWF0IjoxNTQwNDg0OTMzLCJpc3MiOiJ3d3cucHJpbnRmY29kZXIuY29tIiwibmJmIjoxNTQwNDg0OTMzLCJzdWIiOiIxMDAwMDkifQ.Uyegcu_ztxWvVJjRhKjiKpSyEuENyKlz5qKFXpV3l8M";
var fingerprint = "";

var adminHeader = ";laskfqwperqw1234214cfascae09-_ad-";

// 请求拦截
$axios.interceptors.request.use(
  (config: any) => {
    // 处理请求之前的配置
    // if (store.getters.token) {
    //   config.headers['token'] = getToken();
    // }

    // 加载中提示
    // store.commit(TYPES.SET_USER_LOADING_GET);
    config.headers["remember-me-token"] = rememberMeToken;
    config.headers["browser-fingerprint"] = fingerprint;
    config.headers["admin-header"] = adminHeader;
    return config;
  },
  (error: any) => {
    debugger;

    if (error.error.code) {
      console.log(error.error.code);
    }
    // 请求失败的处理
    return Promise.reject(error);
  }
);

$axios.prototype.setRememberMeToken = (token: string) => {
  rememberMeToken = token;
};

$axios.prototype.setBrowserFingerprint = (fp: string) => {
  fingerprint = fp;
};

// 响应拦截
$axios.interceptors.response.use(
  (response: any) => {
    // 加载中提示
    // store.commit(TYPES.SET_USER_LOADING_GET_DONE);

    // 处理响应数据
    if (
      response.status >= 200 &&
      response.status < 300 &&
      response.data.success == true
    ) {
      return response.data;
    } else if (response.data && !response.data.success) {
      const error = new Error(response.data.error);
      // throw error;
      return error;
    } else {
      const error = new Error(response.statusText);
      // throw error;
      return error;
    }
  },
  (error: any) => {
    //请求超时的之后，抛出 error.code = ECONNABORTED的错误..错误信息是 timeout of  xxx ms exceeded
    if (
      error.code === "ECONNABORTED" &&
      error.message.indexOf("timeout") != -1
    ) {
      const config = error.config;
      // If config does not exist or the retry option is not set, reject
      if (!config || !config.retry) return Promise.reject(error);

      // Set the variable for keeping track of the retry count
      config.__retryCount = config.__retryCount || 1;

      // Check if we've maxed out the total number of retries
      if (config.__retryCount >= config.retry) {
        return Promise.reject(error);
      }

      // Increase the retry count
      config.__retryCount += 1;

      // Create new promise to handle exponential backoff
      let backoff = new Promise(resolve => {
        setTimeout(() => {
          // console.log("resolve");
          resolve();
        }, config.retryDelay || 1);
      });

      return backoff.then(() => $axios(config));
    } else {
      // 处理响应失败
      // console.log("err" + error);
      return Promise.reject(error);
    }
  }
);

export default $axios;
