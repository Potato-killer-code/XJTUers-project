const BASE_URL = '/api/v1'

/**
 * 通用请求封装
 */
async function request(url, options = {}) {
  const config = {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  }

  try {
    const res = await fetch(BASE_URL + url, config)
    const data = await res.json()
    return data
  } catch (err) {
    return { code: -1, message: '网络请求失败，请检查网络连接' }
  }
}

/**
 * 存入物品
 * @param {string} code - 4位数字密码
 */
export function storeItem(code) {
  return request('/cabinet/store', {
    method: 'POST',
    body: JSON.stringify({ code }),
  })
}

/**
 * 取出物品
 * @param {string} code - 4位数字密码
 */
export function retrieveItem(code) {
  return request('/cabinet/retrieve', {
    method: 'POST',
    body: JSON.stringify({ code }),
  })
}

/**
 * 查询柜子状态
 */
export function getStatus() {
  return request('/cabinet/status')
}
