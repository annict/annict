/* eslint-disable @typescript-eslint/no-explicit-any */

export default (obj: any) => {
  return Object.keys(obj)
    .map((k) => encodeURIComponent(k) + "=" + encodeURIComponent(obj[k]))
    .join("&");
};
