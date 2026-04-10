# Encoding Status 页面设计文档

## 1. 页面结构

```markdown
┌──────────────────────────────────────────────────────────────────────────────┐
│ 顶部导航栏                                                                 │
├───────────────┬──────────────────────────────────────────────────────────────┤
│               │                                                              │
│ 侧边栏        │ 主内容区                                                     │
│               │                                                              │
│ - Dashboard   │ ┌──────────────────────────────────────────────────────────┐ │
│ - Users       │ │ 页面标题：Encoding Status                               │ │
│ - Media       │ │ 页面描述：Manage and monitor encoding tasks             │ │
│ - **Encoding**│ ├──────────────────────────────────────────────────────────┤ │
│   - Status    │ │ 状态概览卡片                                            │ │
│   - Tasks     │ │ ┌─────────┬─────────┬─────────┬─────────┬─────────┐    │ │
│   - Profiles  │ │ │ 总任务  │ 待处理  │ 处理中  │ 已完成  │ 失败    │    │ │
│ - Settings    │ │ │ 100     │ 5       │ 3       │ 80      │ 12      │    │ │
│               │ │ └─────────┴─────────┴─────────┴─────────┴─────────┘    │ │
│               │ ├──────────────────────────────────────────────────────────┤ │
│               │ │ 操作按钮：                                              │ │
│               │ │ [刷新状态] [重试失败任务] [导出报告]                     │ │
│               │ ├──────────────────────────────────────────────────────────┤ │
│               │ │ 编码任务列表                                            │ │
│               │ │ ┌──────────────────────────────────────────────────────┐ │ │
│               │ │ │ 搜索框                                              │ │ │
│               │ │ ├──────────────────────────────────────────────────────┤ │ │
│               │ │ │ 状态筛选：[全部] [待处理] [处理中] [已完成] [失败]   │ │ │
│               │ │ ├──────────────────────────────────────────────────────┤ │ │
│               │ │ │ 表格：任务ID | 媒体文件 | 状态 | 开始时间 | 完成时间 │ │ │
│               │ │ │       | 进度 | 操作                                  │ │ │
│               │ │ └──────────────────────────────────────────────────────┘ │ │
│               │ ├──────────────────────────────────────────────────────────┤ │
│               │ │ 分页控件                                                │ │
│               │ └──────────────────────────────────────────────────────────┘ │
└───────────────┴──────────────────────────────────────────────────────────────┘
```

## 2. 组件设计

### 2.1 状态概览卡片

```tsx
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-4 mb-6">
  <Card>
    <CardContent className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">总任务数</p>
          <p className="text-2xl font-bold">100</p>
        </div>
        <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center">
          <TasksIcon className="h-5 w-5 text-blue-600" />
        </div>
      </div>
    </CardContent>
  </Card>
  
  <Card>
    <CardContent className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">待处理</p>
          <p className="text-2xl font-bold text-yellow-600">5</p>
        </div>
        <div className="w-10 h-10 rounded-full bg-yellow-100 flex items-center justify-center">
          <ClockIcon className="h-5 w-5 text-yellow-600" />
        </div>
      </div>
    </CardContent>
  </Card>
  
  <Card>
    <CardContent className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">处理中</p>
          <p className="text-2xl font-bold text-blue-600">3</p>
        </div>
        <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center">
          <LoaderIcon className="h-5 w-5 text-blue-600 animate-spin" />
        </div>
      </div>
    </CardContent>
  </Card>
  
  <Card>
    <CardContent className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">已完成</p>
          <p className="text-2xl font-bold text-green-600">80</p>
        </div>
        <div className="w-10 h-10 rounded-full bg-green-100 flex items-center justify-center">
          <CheckIcon className="h-5 w-5 text-green-600" />
        </div>
      </div>
    </CardContent>
  </Card>
  
  <Card>
    <CardContent className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">失败</p>
          <p className="text-2xl font-bold text-red-600">12</p>
        </div>
        <div className="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
          <XIcon className="h-5 w-5 text-red-600" />
        </div>
      </div>
    </CardContent>
  </Card>
  
  <Card>
    <CardContent className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">断开连接</p>
          <p className="text-2xl font-bold text-purple-600">2</p>
        </div>
        <div className="w-10 h-10 rounded-full bg-purple-100 flex items-center justify-center">
          <WifiOffIcon className="h-5 w-5 text-purple-600" />
        </div>
      </div>
    </CardContent>
  </Card>
</div>
```

### 2.2 操作按钮 (Toolbar 位置)

```tsx
<div className="bg-white rounded-lg border shadow-sm">
  {/* Toolbar - 标题和操作按钮 */}
  <div className="p-4 border-b flex items-center justify-between">
    <div>
      <h2 className="text-xl font-semibold text-gray-900">编码任务</h2>
      <p className="text-sm text-gray-500">管理和监控编码任务</p>
    </div>
    <div className="flex flex-wrap gap-3">
      <Button onClick={refreshStatus}>
        <RefreshCwIcon className="mr-2 h-4 w-4" />
        刷新状态
      </Button>
      <Button variant="default" className="bg-yellow-600 hover:bg-yellow-700">
        <RotateCcwIcon className="mr-2 h-4 w-4" />
        重试失败任务
      </Button>
      <Button variant="default" className="bg-indigo-600 hover:bg-indigo-700">
        <DownloadIcon className="mr-2 h-4 w-4" />
        导出报告
      </Button>
    </div>
  </div>
  
  {/* 搜索和筛选 */}
  <div className="p-4 border-b">
    <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div className="relative w-full md:w-64">
        <SearchIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
        <Input 
          placeholder="搜索任务..." 
          className="pl-10"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
      </div>
      <div className="flex flex-wrap gap-2">
        <Button 
          variant={statusFilter === 'all' ? 'default' : 'secondary'}
          onClick={() => setStatusFilter('all')}
        >
          全部
        </Button>
        <Button 
          variant={statusFilter === 'pending' ? 'default' : 'secondary'}
          className={statusFilter === 'pending' ? 'bg-yellow-600 hover:bg-yellow-700' : ''}
          onClick={() => setStatusFilter('pending')}
        >
          待处理
        </Button>
        <Button 
          variant={statusFilter === 'processing' ? 'default' : 'secondary'}
          className={statusFilter === 'processing' ? 'bg-blue-600 hover:bg-blue-700' : ''}
          onClick={() => setStatusFilter('processing')}
        >
          处理中
        </Button>
        <Button 
          variant={statusFilter === 'completed' ? 'default' : 'secondary'}
          className={statusFilter === 'completed' ? 'bg-green-600 hover:bg-green-700' : ''}
          onClick={() => setStatusFilter('completed')}
        >
          已完成
        </Button>
        <Button 
          variant={statusFilter === 'failed' ? 'default' : 'secondary'}
          className={statusFilter === 'failed' ? 'bg-red-600 hover:bg-red-700' : ''}
          onClick={() => setStatusFilter('failed')}
        >
          失败
        </Button>
        <Button 
          variant={statusFilter === 'disconnected' ? 'default' : 'secondary'}
          className={statusFilter === 'disconnected' ? 'bg-purple-600 hover:bg-purple-700' : ''}
          onClick={() => setStatusFilter('disconnected')}
        >
          断开连接
        </Button>
      </div>
    </div>
  </div>
```

### 2.3 编码任务列表

```tsx
  {/* 表格 */}
  <div className="overflow-x-auto">
    <table className="w-full">
      <thead className="bg-gray-50">
        <tr>
          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            任务 ID
          </th>
          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            媒体文件
          </th>
          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            状态
          </th>
          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            开始时间
          </th>
          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            完成时间
          </th>
          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            进度
          </th>
          <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
            操作
          </th>
        </tr>
      </thead>
      <tbody className="bg-white divide-y divide-gray-200">
        {tasks.map((task) => (
          <tr key={task.id}>
            <td className="px-4 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
              {task.id}
            </td>
            <td className="px-4 py-4 whitespace-nowrap text-sm text-gray-500">
              {task.mediaName}
            </td>
            <td className="px-4 py-4 whitespace-nowrap">
              <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                task.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                task.status === 'processing' ? 'bg-blue-100 text-blue-800' :
                task.status === 'completed' ? 'bg-green-100 text-green-800' :
                task.status === 'failed' ? 'bg-red-100 text-red-800' :
                'bg-purple-100 text-purple-800'
              }`}>
                {task.status === 'disconnected' ? '断开连接' : task.status}
              </span>
            </td>
            <td className="px-4 py-4 whitespace-nowrap text-sm text-gray-500">
              {task.startTime}
            </td>
            <td className="px-4 py-4 whitespace-nowrap text-sm text-gray-500">
              {task.endTime || '-'}
            </td>
            <td className="px-4 py-4 whitespace-nowrap">
              <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div 
                  className={`h-2.5 rounded-full ${
                    task.status === 'processing' ? 'bg-blue-600' :
                    task.status === 'completed' ? 'bg-green-600' :
                    task.status === 'failed' ? 'bg-red-600' :
                    task.status === 'disconnected' ? 'bg-purple-600' :
                    'bg-yellow-600'
                  }`} 
                  style={{ width: `${task.progress}%` }}
                ></div>
              </div>
              <span className="text-xs text-gray-500 mt-1 block">
                {task.progress}%
              </span>
            </td>
            <td className="px-4 py-4 whitespace-nowrap text-right text-sm font-medium">
              <Button 
                variant="ghost" 
                size="sm"
                onClick={() => retryTask(task.id)}
                className="text-blue-600 hover:text-blue-900"
              >
                重试
              </Button>
              <Button 
                variant="ghost" 
                size="sm"
                onClick={() => viewTaskDetails(task.id)}
                className="text-gray-600 hover:text-gray-900 ml-2"
              >
                详情
              </Button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  </div>
  
  {/* 分页 */}
  <div className="p-4 border-t flex items-center justify-between">
    <div className="text-sm text-gray-700">
      显示 {((currentPage - 1) * pageSize) + 1} 到 {Math.min(currentPage * pageSize, totalTasks)} 条，共 {totalTasks} 条
    </div>
    <div className="flex items-center gap-2">
      <Button 
        variant="secondary" 
        size="sm"
        onClick={() => setCurrentPage(currentPage - 1)}
        disabled={currentPage === 1}
      >
        上一页
      </Button>
      <span className="text-sm text-gray-700">
        {currentPage} / {Math.ceil(totalTasks / pageSize)}
      </span>
      <Button 
        variant="secondary" 
        size="sm"
        onClick={() => setCurrentPage(currentPage + 1)}
        disabled={currentPage >= Math.ceil(totalTasks / pageSize)}
      >
        下一页
      </Button>
    </div>
  </div>
</div>
```

## 3. 数据结构

```typescript
// 编码任务类型
interface EncodingTask {
  id: number;
  mediaId: number;
  mediaName: string;
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'disconnected';
  progress: number;
  startTime: string;
  endTime: string | null;
  errorMessage: string | null;
}

// 编码状态
interface EncodingStatus {
  totalTasks: number;
  pending: number;
  processing: number;
  completed: number;
  failed: number;
  disconnected: number;
  successRate: string;
}
```

## 4. 交互逻辑

```typescript
// 状态管理
const [status, setStatus] = useState<EncodingStatus>({
  totalTasks: 0,
  pending: 0,
  processing: 0,
  completed: 0,
  failed: 0,
  disconnected: 0,
  successRate: '0%'
});

const [tasks, setTasks] = useState<EncodingTask[]>([]);
const [searchQuery, setSearchQuery] = useState('');
const [statusFilter, setStatusFilter] = useState('all');
const [currentPage, setCurrentPage] = useState(1);
const [pageSize, setPageSize] = useState(25);
const [totalTasks, setTotalTasks] = useState(0);

// API 调用
const fetchEncodingStatus = async () => {
  const response = await fetch('/api/admin/encoding/status');
  const data = await response.json();
  setStatus(data);
};

const fetchEncodingTasks = async () => {
  const response = await fetch(`/api/admin/encoding/tasks?status=${statusFilter}&page=${currentPage}&page_size=${pageSize}&search=${searchQuery}`);
  const data = await response.json();
  setTasks(data.tasks);
  setTotalTasks(data.total);
};

// 操作函数
const refreshStatus = () => {
  fetchEncodingStatus();
  fetchEncodingTasks();
};

const retryTask = async (taskId: number) => {
  await fetch(`/api/admin/encoding/tasks/${taskId}/retry`, { method: 'POST' });
  refreshStatus();
};

const retryAllFailedTasks = async () => {
  await fetch('/api/admin/encoding/retry-failed', { method: 'POST' });
  refreshStatus();
};

const viewTaskDetails = (taskId: number) => {
  // 显示任务详情对话框
  setSelectedTaskId(taskId);
  setShowDetailsDialog(true);
};
```

## 5. 响应式设计

```tailwind
/* 响应式布局类 */
@media (max-width: 768px) {
  .grid-cols-1 {
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }
  
  .md\:grid-cols-2 {
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }
  
  .lg\:grid-cols-5 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .hidden-mobile {
    display: none;
  }
}
```

## 6. 样式标准

### 6.0 样式设计原则

1. **样式与语言分离**：样式规范不涉及语言内容，语言由多语言管理系统负责。所有样式定义应使用通用的结构和布局，不绑定具体的文本内容。

2. **样式与功能分离**：样式规范只定义视觉呈现，不涉及功能逻辑。功能元素的数量（如卡片数量）由业务逻辑决定，样式应提供统一的模板，适用于不同数量的功能元素。

3. **组件化设计**：所有样式应采用组件化设计，确保在不同页面和场景中保持一致的视觉效果。

4. **响应式适配**：样式应支持不同屏幕尺寸的适配，确保在各种设备上都能提供良好的用户体验。

5. **可扩展性**：样式规范应具备良好的可扩展性，能够适应未来功能的变化和新增。

### 6.1 卡片样式

#### 标准卡片
- **容器**：`bg-white rounded-lg border border-gray-200 shadow-sm`
- **内边距**：`p-6`
- **标题**：`text-sm font-medium text-gray-500`
- **数值**：`text-2xl font-bold`
- **图标容器**：`w-10 h-10 rounded-full flex items-center justify-center`
- **底部子标题**：可选，使用 `text-xs text-gray-400 mt-1`
- **底部色条**：必填，使用 `h-1 w-full`，颜色根据卡片类型确定

#### 状态卡片颜色规范
| 状态 | 标题颜色 | 数值颜色 | 图标背景 | 图标颜色 | 底部色条（必填） |
|------|----------|----------|----------|----------|------------------|
| 总任务 | text-gray-500 | text-gray-900 | bg-blue-100 | text-blue-600 | bg-blue-500 |
| 待处理 | text-gray-500 | text-yellow-600 | bg-yellow-100 | text-yellow-600 | bg-yellow-500 |
| 处理中 | text-gray-500 | text-blue-600 | bg-blue-100 | text-blue-600 | bg-blue-500 |
| 已完成 | text-gray-500 | text-green-600 | bg-green-100 | text-green-600 | bg-green-500 |
| 失败 | text-gray-500 | text-red-600 | bg-red-100 | text-red-600 | bg-red-500 |
| 断开连接 | text-gray-500 | text-purple-600 | bg-purple-100 | text-purple-600 | bg-purple-500 |

#### 卡片元素完整性
- **必填元素**：所有卡片必须包含以下元素：容器、标题、数值、图标容器和底部色条
- **元素一致性**：所有卡片的元素结构必须保持一致，确保视觉统一
- **颜色一致性**：不同状态的卡片应使用对应的颜色，确保视觉区分度
- **布局一致性**：所有卡片的布局结构必须保持一致，确保视觉整齐
- **响应式一致性**：所有卡片在不同屏幕尺寸下的表现必须保持一致

### 6.2 标题大小标准

| 级别 | 样式类 | 字体大小 | 字重 | 颜色 | 用途 |
|------|--------|----------|------|------|------|
| H1 | text-2xl | 1.5rem | font-bold | text-gray-900 | 页面主标题 |
| H2 | text-xl | 1.25rem | font-semibold | text-gray-900 | 区域标题 |
| H3 | text-lg | 1.125rem | font-medium | text-gray-900 | 子区域标题 |
| 副标题 | text-sm | 0.875rem | font-medium | text-gray-500 | 卡片标题、描述文本 |
| 正文 | text-sm | 0.875rem | font-normal | text-gray-700 | 普通文本 |
| 小文本 | text-xs | 0.75rem | font-normal | text-gray-500 | 辅助文本、时间戳 |

### 6.3 搜索组件样式

#### 标准搜索框
- **容器**：`relative w-full md:w-64`
- **图标**：`absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400`
- **输入框**：`pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent`
- **占位符**：`text-gray-400`
- **高度**：统一为 `h-10`（40px）
- **边框**：统一为 `border-gray-300`，聚焦时为 `border-transparent`

#### 筛选按钮
- **标准状态**：`px-3 py-1.5 text-sm font-medium rounded-md border border-gray-300 bg-white text-gray-700 hover:bg-gray-50`
- **激活状态**：根据状态使用对应颜色，例如 `bg-blue-600 hover:bg-blue-700 text-white border-transparent`
- **间距**：按钮之间 `gap-2`

### 6.4 输入框(Input)样式标准

#### 标准输入框
- **基础样式**：`px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent`
- **高度**：统一为 `h-10`（40px）
- **宽度**：根据容器自适应，默认为 `w-full`
- **边框**：统一为 `border-gray-300`，聚焦时为 `border-transparent`
- **颜色**：`text-gray-900`，占位符为 `text-gray-400`
- **字体**：`text-sm font-normal`

#### 输入框变体
- **带图标的输入框**：参考搜索框样式，图标位于左侧
- **带后缀的输入框**：图标或文本位于右侧，使用 `relative` 容器
- **多行文本框**：`min-h-20 p-4 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent`

### 6.5 下拉选择框样式标准

#### 标准下拉框
- **容器**：`relative w-full`
- **选择器**：`px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent appearance-none bg-white`
- **箭头图标**：`absolute right-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400 pointer-events-none`
- **高度**：统一为 `h-10`（40px）
- **宽度**：根据容器自适应，默认为 `w-full`
- **边框**：统一为 `border-gray-300`，聚焦时为 `border-transparent`

#### 下拉菜单
- **容器**：`absolute z-10 mt-1 w-full bg-white rounded-md shadow-lg border border-gray-200`
- **选项**：`px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer`
- **选中项**：`bg-primary text-white hover:bg-primary/90`
- **间距**：选项之间无间距，菜单内边距为 `py-1`

### 6.6 按钮样式

### 6.4 按钮样式

#### 主按钮
- **样式**：`px-4 py-2 rounded-md font-medium focus:outline-none focus:ring-2 focus:ring-offset-2`
- **默认颜色**：`bg-primary hover:bg-primary/90 text-white`
- **大小**：默认 `h-10`（40px）

#### 次要按钮
- **样式**：`px-4 py-2 rounded-md font-medium border border-gray-300 focus:outline-none focus:ring-2 focus:ring-offset-2`
- **默认颜色**：`bg-white hover:bg-gray-50 text-gray-700`
- **大小**：默认 `h-10`（40px）

#### 幽灵按钮
- **样式**：`px-3 py-1.5 text-sm font-medium rounded-md focus:outline-none`
- **默认颜色**：`text-gray-600 hover:text-gray-900 hover:bg-gray-50`
- **大小**：`h-8`（32px）

### 6.5 表格样式

#### 表头
- **背景**：`bg-gray-50`
- **文字**：`text-left text-xs font-medium text-gray-500 uppercase tracking-wider`
- **内边距**：`px-4 py-3`

#### 表格行
- **默认**：`bg-white`
- **交替行**：无（统一白色背景）
- **边框**：`divide-y divide-gray-200`
- **内边距**：`px-4 py-4`
- **文字**：`text-sm text-gray-900`（主键），`text-sm text-gray-500`（普通字段）

### 6.6 状态标签样式

| 状态 | 样式类 | 显示文本 |
|------|--------|----------|
| 待处理 | `px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-yellow-100 text-yellow-800` | 待处理 |
| 处理中 | `px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800` | 处理中 |
| 已完成 | `px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800` | 已完成 |
| 失败 | `px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-red-100 text-red-800` | 失败 |
| 断开连接 | `px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-purple-100 text-purple-800` | 断开连接 |

### 6.7 进度条样式

- **容器**：`w-full bg-gray-200 rounded-full h-2.5`
- **进度**：`h-2.5 rounded-full`，颜色根据状态确定
- **百分比文本**：`text-xs text-gray-500 mt-1 block`

| 状态 | 进度条颜色 |
|------|------------|
| 待处理 | bg-yellow-600 |
| 处理中 | bg-blue-600 |
| 已完成 | bg-green-600 |
| 失败 | bg-red-600 |
| 断开连接 | bg-purple-600 |

### 6.8 分页控件样式

- **容器**：`p-4 border-t flex items-center justify-between`
- **信息文本**：`text-sm text-gray-700`
- **按钮**：使用次要按钮样式，`size="sm"`
- **页码**：`text-sm text-gray-700`
- **间距**：按钮之间 `gap-2`

### 6.9 添加功能按钮位置

#### 标准位置
- **主要添加按钮**：位于页面或区域的右上角，与其他操作按钮一起
- **表格内添加按钮**：位于表格工具栏的右侧，与搜索框和筛选控件在同一行
- **卡片内添加按钮**：位于卡片头部的右侧，与卡片标题在同一行
- **一致性**：所有添加按钮应使用相同的位置规则，确保在不同页面和场景中保持一致

#### 样式规范
- **按钮类型**：使用主按钮样式
- **图标**：添加 `PlusIcon` 图标，位于文字左侧
- **文字**：统一使用 "添加" 或 "新建"
- **大小**：默认使用标准主按钮大小（`h-10`）
- **一致性**：所有添加按钮应使用相同的样式，确保在不同页面和场景中保持一致

### 6.10 ACTION 按钮样式

#### 标准样式
- **更多操作按钮**：使用垂直省略号图标（`MoreVerticalIcon`）
- **图标按钮**：使用圆形背景的图标按钮，`w-8 h-8 flex items-center justify-center rounded-full hover:bg-gray-100`
- **下拉菜单**：点击 ACTION 按钮后显示下拉菜单，包含具体操作选项
- **一致性**：所有ACTION按钮应使用相同的样式，确保在不同页面和场景中保持一致

#### 统一规范
- **位置**：位于表格行的最后一列，靠右对齐
- **大小**：统一使用 `w-8 h-8` 的图标按钮
- **颜色**：默认 `text-gray-500 hover:text-gray-700`
- **交互**：悬停时显示背景色变化
- **一致性**：所有ACTION按钮应使用相同的位置和交互规则，确保在不同页面和场景中保持一致

### 6.11 表格选中后操作按钮位置

#### 标准位置
- **批量操作栏**：位于表格上方，工具栏区域的左侧或右侧
- **操作按钮**：与搜索框、筛选控件在同一工具栏内
- **显示条件**：仅当表格中有选中项时显示
- **一致性**：所有表格选中后操作按钮应使用相同的位置规则，确保在不同页面和场景中保持一致

#### 样式规范
- **容器**：`flex items-center gap-2`
- **按钮**：使用次要按钮样式，`size="sm"`
- **文字**：显示选中数量，例如 "已选择 3 项"
- **操作**：包含批量删除、批量编辑等常用操作
- **一致性**：所有表格选中后操作按钮应使用相同的样式，确保在不同页面和场景中保持一致

### 6.12 工具栏布局标准

#### 标准布局
- **左侧**：页面标题和描述
- **右侧**：操作按钮组（添加按钮、刷新按钮等）
- **搜索和筛选**：位于工具栏下方，单独一行
- **批量操作**：位于搜索和筛选区域的左侧
- **一致性**：所有工具栏应使用相同的布局规则，确保在不同页面和场景中保持一致

#### 响应式布局
- **桌面端**：水平排列，标题左对齐，按钮右对齐
- **移动端**：垂直排列，标题在上，按钮在下
- **间距**：组件之间保持 `gap-4` 的间距
- **一致性**：所有工具栏应使用相同的响应式布局规则，确保在不同页面和场景中保持一致

