function ResponseError(this: any, response: any) {
  this.message = `Request failed with status code ${response.status}`;
  this.response = response;
}

const request = async (url: string, method: string, options: any = {}) => {
  const headers = {
    ...(options.headers || {}),
    'Content-Type': 'application/json',
    'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content'),
  };

  delete options.headers;

  const res = await fetch(url, { ...{ method }, headers, ...options });

  if (!res.ok) {
    throw new (ResponseError as any)(res);
  }

  const data = await res.text();
  return data === '' ? {} : JSON.parse(data);
};

const requestWithData = async (url: string, method: string, data = {}, options = {}) => {
  const body = JSON.stringify(data);

  return await request(url, method, {
    ...{ body },
    ...options,
  });
};

export default {
  get: async (url: string, options = {}) => {
    return await request(url, 'GET', { ...options });
  },

  post: async (url: string, data = {}, options = {}) => {
    return await requestWithData(url, 'POST', data, { ...options });
  },

  delete: async (url: string, data = {}, options = {}) => {
    return await requestWithData(url, 'DELETE', data, { ...options });
  },
};
