// 从统一配置文件获取应用主页地址
const APP_URL = window.LinkDockConfig.baseUrl; 

document.getElementById('app-frame').src = `${APP_URL}?newtab=true`;
