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

