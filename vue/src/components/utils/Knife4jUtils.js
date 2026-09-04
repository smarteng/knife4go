/**
 * 根据url菜单查找组件名称,用于打开Tab选项卡
 * @param {*} path 路径
 * @param {*} menuData 菜单集合
 */
export function findComponentsByPath(path, menuData) {
  path = decodeURIComponent(path)
  var tmpComp = null;
  for (var i = 0; i < menuData.length; i++) {
    if (menuData[i].path == path) {
      tmpComp = menuData[i];
      break;
    }
    if (tmpComp == null) {
      var chds = menuData[i].children;
      if (chds != undefined && chds !== null) {
        tmpComp = findComponentsByPath(path, chds);
      }
    }
  }
  return tmpComp;
}

/**
 * 根据菜单主键key查找菜单项
 * @param {*} key key值
 * @param {*} menuData 菜单集合
 */
export function findMenuByKey(key, menuData) {
  var tmpComp = null;
  for (var i = 0; i < menuData.length; i++) {
    if (menuData[i].key == key) {
      tmpComp = menuData[i];
      break;
    }
    if (tmpComp == null) {
      var chds = menuData[i].children;
      if (chds != undefined && chds !== null) {
        tmpComp = findMenuByKey(key, chds);
      }
    }
  }
  return tmpComp;
}

/**
 * 计算字符串显示宽度（中文/全角字符计 2，其余字符计 1）。
 * 替代原先挂在 String.prototype 上的 gblen 猴子补丁，避免污染全局原型。
 * @param {string} text 待计算字符串
 * @returns {number} 显示宽度
 */
export function gblen(text) {
  if (text == null) {
    return 0;
  }
  var len = 0;
  var str = String(text);
  for (var i = 0; i < str.length; i++) {
    if (str.charCodeAt(i) > 127 || str.charCodeAt(i) == 94) {
      len += 2;
    } else {
      len++;
    }
  }
  return len;
}
