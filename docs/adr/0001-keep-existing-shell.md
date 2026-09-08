# 保留 AI-Manager 现有 Shell，不复刻 Kitter 的可调整大小面板

AI-Manager 已有自己的窗口壳（侧边栏 + 主区域 + 自定义标题栏 + 15 主题），Kitter 用的是 gpui 的 `h_resizable` 可调整大小面板（侧边栏 180px 可缩 160–260，内容区列表 280px 可缩 220–420）。本决策选择保留 AI-Manager 现有壳，不引入 `react-resizable-panels` 到外层结构，也不做像素级复刻。

理由：AI-Manager 的壳已经跑通、有自己的主题系统，是现有资产；Kitter 的可调整大小面板只是其 gpui 渲染栈的产物，不是它的核心价值。我们要复刻的是**领域功能 + 交互模式 + 视觉语言**，不是把 gpui 的 layout API 搬到 React。如果某页内部确实需要列表+详情可调整大小（Kitter 的 `content()` 函数），那在页内引入 `react-resizable-panels` 即可，但不改变外层壳。

后果：外层侧边栏宽度是 AI-Manager 的既有行为（可能固定、可能有自己的拖拽），不会跟 Kitter 完全一致；这是可接受的差异——目标用户看的是「像 Kitter 的桌面应用」，不是「像素级相同的 Kitter」。
