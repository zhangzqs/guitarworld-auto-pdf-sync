# guitarworld-auto-pdf-sync

吉他世界网自动同步下载自己已购买曲谱，并导出为 PDF 文本保存至本地

## 功能特性

- ✨ 自动爬取 Guitar World 用户中心的所有曲谱
- 📄 下载并保存为 PDF 文件到本地目录
- 🚀 **并发下载**，支持自定义并发数（默认 3）
- 💾 图片先下载到内存，使用系统临时目录生成 PDF
- 🔄 自动跳过已存在的曲谱，避免重复下载
- 🔐 支持 Cookie 认证，访问需要登录的内容
- 📁 按创建者分类组织文件：`<创建者>/<曲名> - <副标题>.pdf`
- 🎯 命令行参数配置输出路径和并发数

## 环境要求

- Go 1.24.4 或更高版本

## 安装和使用

### 1. 克隆项目

```bash
git clone https://github.com/zhangzqs/guitarworld-auto-pdf-sync.git
cd guitarworld-auto-pdf-sync
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置 Cookies

从浏览器中获取 Guitar World 的 Cookie 信息：

1. 登录 [Guitar World](https://user.guitarworld.com.cn/user/pu/my)
2. 打开浏览器开发者工具 (F12)
3. 切换到 Network 标签
4. 刷新页面
5. 找到对 `https://user.guitarworld.com.cn/user/pu/my` 的请求
6. 在请求头中找到 `Cookie` 字段，复制完整的 Cookie 值

**重要**: 确保 Cookie 中包含 `XSRF-TOKEN` 和 `guitarworld_ses`！

设置环境变量：

```bash
export GUITARWORLD_COOKIES="你的Cookie内容"
```

### 4. 运行程序

基本用法（使用默认设置）：

```bash
go run ./cmd/guitarworld-sync
```

自定义输出目录和并发数：

```bash
go run ./cmd/guitarworld-sync -output ./my_pdfs -concurrency 5
```

或编译后运行：

```bash
go build -o guitarworld-sync ./cmd/guitarworld-sync
./guitarworld-sync -output ./my_pdfs -concurrency 5
```

### 命令行参数

```
-output string
    输出目录路径 (默认 "./pdfs")

-concurrency int
    并发下载数量 (默认 3)
```

## 输出结构

所有下载的 PDF 文件将按照创建者分类保存：

```
输出目录/
├── 化身无/
│   ├── [指弹吉他谱] 昔涟 - 崩坏星穹铁道歌曲.pdf
│   ├── [指弹吉他谱] 不眠之夜 - 崩坏：星穹铁道 - 匹诺康尼主题曲.pdf
│   ├── [指弹吉他谱] 水龙吟 - Samudrartha 崩坏星穹铁道-长生梦短 Svah Sanishyu.pdf
│   └── [弹唱吉他谱] 我不曾忘记 - 游戏《原神》2023新春会同人曲.pdf
├── 吉他老杨/
│   └── [指弹吉他谱] 隐形的翅膀.pdf
├── 平台官方/
│   ├── [弹唱吉他谱] 吉姆餐厅.pdf
│   └── [弹唱吉他谱] 夏天的风.pdf
└── 大V教你弹吉他/
    └── [指弹吉他谱] 咱们屯里的人.pdf
```

文件名格式：
- 完整格式：`[曲谱类型] <曲名> - <副标题>.pdf`
- 无副标题：`[曲谱类型] <曲名>.pdf`
- 曲谱类型包括：`指弹吉他谱`、`弹唱吉他谱` 等
- 保留空格，只替换非法文件系统字符（`<>:"/\|?*` 等）

## 性能优化

### 并发下载

程序支持并发下载多个曲谱，可以显著提高下载速度：

```bash
# 使用 5 个并发线程
go run main.go -concurrency 5

# 更激进的并发（注意可能触发限流）
go run main.go -concurrency 10
```

**建议**:
- 默认值 3 是比较稳妥的选择
- 网络条件好可以尝试 5-8
- 不建议超过 10，可能触发服务器限流

### 内存管理

- 图片首先下载到内存中
- 仅在生成 PDF 时使用系统临时目录（自动清理）
- 每个曲谱处理完成后立即释放临时文件

## 使用 GitHub Actions 自动下载

### 方式一: 使用 GitHub Actions（推荐，无需本地环境）

1. Fork 本项目到你的 GitHub 账号
2. 在你的仓库中，进入 **Actions** 页面
3. 选择 **Download PDFs** 工作流
4. 点击 **Run workflow**
5. 填入参数:
   - **cookies**: 你的 Guitar World Cookie（必填）
   - **output_dir**: 输出目录名称（可选，默认 `pdfs`）
   - **concurrency**: 并发数（可选，默认 `3`）
6. 点击 **Run workflow** 开始下载
7. 等待工作流完成后，在 **Artifacts** 中下载打包好的 PDF 文件

**优点:**
- ✅ 无需本地安装 Go 环境
- ✅ 云端运行，不占用本地资源
- ✅ 自动打包，直接下载
- ✅ 保留 7 天，可随时下载

### 方式二: 下载预编译版本

在 [Releases](https://github.com/zhangzqs/guitarworld-auto-pdf-sync/releases) 页面下载对应平台的二进制文件:

- **Linux**: `guitarworld-sync_Linux_x86_64.tar.gz`
- **Windows**: `guitarworld-sync_Windows_x86_64.zip`
- **macOS**: `guitarworld-sync_Darwin_x86_64.tar.gz` (Intel) 或 `guitarworld-sync_Darwin_arm64.tar.gz` (Apple Silicon)

解压后直接运行，无需安装 Go 环境。

## 注意事项

1. 🔑 **Cookie 有时效性**，如果下载失败可能需要重新获取 Cookie
2. 🔄 程序会自动跳过已存在的文件，可以安全地重复运行
3. ⏱️ 下载过程中会有适当延迟，避免对服务器造成过大压力
4. ⚖️ 请确保你只下载自己已购买的曲谱，尊重版权
5. 🌐 建议在网络稳定的环境下运行，避免下载中断
6. 🔐 **使用 GitHub Actions 时请注意**: Cookie 信息会在工作流日志中可见，建议使用后立即删除工作流运行记录

## 工作原理

1. 使用提供的 Cookie 访问用户的曲谱列表 API
2. 自动分页获取所有曲谱信息（ID、标题、创建者等）
3. 并发访问每个曲谱的详情页面
4. 使用 goquery 解析 HTML，提取曲谱图片 URL
5. 并发下载所有图片到内存
6. 使用系统临时目录生成 PDF（自动检测图片格式）
7. 保存 PDF 到指定输出目录，按创建者分类

## 常见问题

### Q: 提示 "Please set GUITARWORLD_COOKIES environment variable"

**A**: 需要设置 `GUITARWORLD_COOKIES` 环境变量：

```bash
export GUITARWORLD_COOKIES="你的Cookie内容"
```

### Q: 下载失败或返回 403/401 错误

**A**: Cookie 可能已过期，需要：
1. 重新登录 Guitar World
2. 从浏览器获取最新的 Cookie
3. 确保 Cookie 中包含 `XSRF-TOKEN` 和 `guitarworld_ses`

### Q: 找不到曲谱或下载数量为 0

**A**: 可能的原因：
- 网页结构变化导致解析失败
- 你的账号下确实没有曲谱
- Cookie 认证失败

### Q: 程序运行很慢

**A**: 尝试增加并发数：

```bash
go run main.go -concurrency 5
```

### Q: 出现 "too many requests" 错误

**A**: 降低并发数，避免触发服务器限流：

```bash
go run main.go -concurrency 2
```

## 示例输出

```
2025/11/29 15:40:47 Fetching sheet music list from API...
2025/11/29 15:40:47 Fetching page 1...
2025/11/29 15:40:47 Page 1: found 20 items (total so far: 20)
...

=== Found 74 sheet music items ===

Using concurrency: 5

[1/74] Processing: 昔涟 - 崩坏星穹铁道歌曲 (ID: 193812) by 化身无
[2/74] Processing: 不眠之夜 - 崩坏：星穹铁道 - 匹诺康尼主题曲 (ID: 89809) by 化身无
[1/74] Found 4 images, downloading...
[2/74] Found 5 images, downloading...
[1/74] Creating PDF with 4 pages...
[1/74] Successfully created: my_pdfs/化身无/昔涟 - 崩坏星穹铁道歌曲.pdf
...

=== Summary ===
Total: 74
Success: 70
Skipped: 2
Errors: 2
```

## 项目结构

```
guitarworld-auto-pdf-sync/
├── cmd/
│   └── guitarworld-sync/    # 主程序入口
│       └── main.go
├── internal/
│   ├── models/              # 数据模型
│   │   └── sheet.go
│   ├── scraper/             # 爬虫逻辑
│   │   ├── scraper.go       # HTTP 请求和数据获取
│   │   └── processor.go     # 曲谱处理和文件管理
│   └── pdf/                 # PDF 生成
│       └── generator.go
├── go.mod
├── run.sh                   # 运行脚本
└── README.md
```

## 技术栈

- **语言**: Go 1.24.4
- **HTML 解析**: [goquery](https://github.com/PuerkitoBio/goquery)
- **PDF 生成**: [gofpdf](https://github.com/jung-kurt/gofpdf)
- **并发控制**: Go 原生 goroutines + channel + sync.WaitGroup
- **项目结构**: 遵循标准 Go 项目布局 (cmd/internal)

## License

MIT License - 详见 [LICENSE](LICENSE) 文件

## 贡献

欢迎提交 Issue 和 Pull Request！
