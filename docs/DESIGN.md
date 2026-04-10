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
│ - Users       │ │ 统一工具栏                                                │ │
│ - Media       │ │ ┌──────────────────────────────────────────────────────┐ │ │
│ - **Encoding**│ │ │ 页面标题：Transcoding Status                         │ │ │
│   - Status    │ │ │ 页面描述：Live monitoring of video processing workflows│ │ │
│   - Tasks     │ │ │ 连接状态：CONNECTED/DISCONNECTED                   │ │ │
│   - Profiles  │ │ └──────────────────────────────────────────────────────┘ │ │
│ - Settings    │ │ 搜索框和筛选控件                                         │ │
│               │ ├──────────────────────────────────────────────────────────┤ │
│               │ │ 状态概览卡片                                            │ │
│               │ │ ┌─────────┬─────────┬─────────┬─────────┐              │ │
│               │ │ │ 总任务  │ 排队    │ 已完成  │ 失败    │              │ │
│               │ │ │ 100     │ 5       │ 80      │ 12      │              │ │
│               │ │ └─────────┴─────────┴─────────┴─────────┘              │ │
│               │ ├──────────────────────────────────────────────────────────┤ │
│               │ │ 状态标签页（椭圆形按钮）                                 │ │
│               │ │ ┌──────┬──────┬──────┬──────┐                         │ │
│               │ │ │ 全部 │ 排队 │ 完成 │ 失败 │                         │ │
│               │ │ └──────┴──────┴──────┴──────┘                         │ │
│               │ ├──────────────────────────────────────────────────────────┤ │
│               │ │ 编码任务列表                                            │ │
│               │ │ ┌──────────────────────────────────────────────────────┐ │ │
│               │ │ │ 表格：选择框 | MEDIA | PROFILE | PROGRESS | STATUS │ │ │
│               │ │ │       | OUTPUT | TIME | ACTION                      │ │ │
│               │ │ └──────────────────────────────────────────────────────┘ │ │
│               │ ├──────────────────────────────────────────────────────────┤ │
│               │ │ 分页控件                                                │ │
│               │ └──────────────────────────────────────────────────────────┘ │
└───────────────┴──────────────────────────────────────────────────────────────┘
```

## 2. 组件设计

### 2.1 状态概览卡片

```tsx
<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
  <Card className="relative overflow-hidden border-none shadow-sm bg-white dark:bg-slate-900 ring-1 ring-slate-200 dark:ring-slate-800">
    <CardContent className="p-5">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-[11px] font-black uppercase tracking-widest text-slate-400 dark:text-slate-500">Active Jobs</p>
          <h3 className="text-3xl font-bold tabular-nums text-sky-600 dark:text-sky-400">100</h3>
          <p className="text-[10px] text-slate-400 dark:text-slate-500 font-medium">Total tasks registered</p>
        </div>
        <div className="p-2.5 rounded-xl bg-sky-50 dark:bg-sky-950/30 text-sky-500 dark:text-sky-400">
          <Video className="h-6 w-6" />
        </div>
      </div>
      <div className="absolute bottom-0 left-0 h-1 bg-sky-500 w-full opacity-10" />
    </CardContent>
  </Card>
  
  <Card className="relative overflow-hidden border-none shadow-sm bg-white dark:bg-slate-900 ring-1 ring-slate-200 dark:ring-slate-800">
    <CardContent className="p-5">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-[11px] font-black uppercase tracking-widest text-slate-400 dark:text-slate-500">In Queue</p>
          <h3 className="text-3xl font-bold tabular-nums text-amber-600 dark:text-amber-400">5</h3>
          <p className="text-[10px] text-slate-400 dark:text-slate-500 font-medium">Tasks waiting for worker</p>
        </div>
        <div className="p-2.5 rounded-xl bg-amber-50 dark:bg-amber-950/30 text-amber-500 dark:text-amber-400">
          <Clock className="h-6 w-6" />
        </div>
      </div>
      <div className="absolute bottom-0 left-0 h-1 bg-amber-500 w-full opacity-10" />
    </CardContent>
  </Card>
  
  <Card className="relative overflow-hidden border-none shadow-sm bg-white dark:bg-slate-900 ring-1 ring-slate-200 dark:ring-slate-800">
    <CardContent className="p-5">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-[11px] font-black uppercase tracking-widest text-slate-400 dark:text-slate-500">Completed</p>
          <h3 className="text-3xl font-bold tabular-nums text-emerald-600 dark:text-emerald-400">80</h3>
          <p className="text-[10px] text-slate-400 dark:text-slate-500 font-medium">Successfully finished</p>
        </div>
        <div className="p-2.5 rounded-xl bg-emerald-50 dark:bg-emerald-950/30 text-emerald-500 dark:text-emerald-400">
          <CheckCircle2 className="h-6 w-6" />
        </div>
      </div>
      <div className="absolute bottom-0 left-0 h-1 bg-emerald-500 w-full opacity-10" />
    </CardContent>
  </Card>
  
  <Card className="relative overflow-hidden border-none shadow-sm bg-white dark:bg-slate-900 ring-1 ring-slate-200 dark:ring-slate-800">
    <CardContent className="p-5">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-[11px] font-black uppercase tracking-widest text-slate-400 dark:text-slate-500">Failed</p>
          <h3 className="text-3xl font-bold tabular-nums text-red-600 dark:text-red-400">12</h3>
          <p className="text-[10px] text-slate-400 dark:text-slate-500 font-medium">Tasks requiring attention</p>
        </div>
        <div className="p-2.5 rounded-xl bg-red-50 dark:bg-red-950/30 text-red-500 dark:text-red-400">
          <XCircle className="h-6 w-6" />
        </div>
      </div>
      <div className="absolute bottom-0 left-0 h-1 bg-red-500 w-full opacity-10" />
    </CardContent>
  </Card>
</div>
```

### 2.2 操作按钮 (Toolbar 位置)

```tsx
<Card>
  <CardContent className="p-6">
    <div className="space-y-4">
      {/* 页面标题和状态 */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">Transcoding Status</h2>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-1.5 flex items-center gap-2">
            <span className="inline-block w-2 h-2 rounded-full bg-sky-500 animate-pulse"/>
            Live monitoring of video processing workflows
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant="outline" className="gap-2 text-[11px] font-bold px-3 py-1 border-2 text-emerald-500 border-emerald-100 bg-emerald-50/50 dark:bg-emerald-950/20 dark:border-emerald-900">
            <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"/>
            CONNECTED
          </Badge>
        </div>
      </div>

      {/* 分隔线 */}
      <div className="border-t border-slate-200 dark:border-slate-800 my-2"/>

      {/* 搜索和筛选 */}
      <div className="flex flex-col lg:flex-row gap-4">
        <div className="flex-1 min-w-0">
          <div className="relative w-full">
            <SearchIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
            <Input
              placeholder="Search tasks by media ID, profile, or status..."
              className="pl-10 h-9 w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
            />
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Select>
            <SelectTrigger className="w-[140px] h-9 focus:ring-1 focus:ring-ring focus:ring-offset-0">
              <div className="flex items-center gap-2">
                <FilterIcon className="h-4 w-4"/>
                <SelectValue placeholder="Profile"/>
              </div>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="justify-center text-center font-medium opacity-70">--- All ---</SelectItem>
              <SelectItem value="h264-240">H264 240p</SelectItem>
              <SelectItem value="h264-360">H264 360p</SelectItem>
              <SelectItem value="h264-480">H264 480p</SelectItem>
              <SelectItem value="h264-720">H264 720p</SelectItem>
              <SelectItem value="h264-1080">H264 1080p</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="default" size="sm" className="h-9 px-4">
            <SearchIcon className="h-4 w-4 mr-2"/>
            Search
          </Button>
        </div>
      </div>
    </div>
  </CardContent>
</Card>
```

### 2.3 编码任务列表

```tsx
<Card>
  <CardHeader className="pb-3">
    <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <CardTitle>Task List</CardTitle>
        <CardDescription>
          {filteredTasks.length} task{filteredTasks.length !== 1 ? 's' : ''} found
        </CardDescription>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        {selectedRows.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">
              {selectedRows.length} selected
            </span>
            <Separator orientation="vertical" className="h-6"/>
            <Button
              variant="outline"
              size="sm"
              onClick={handleBatchRetry}
            >
              <RotateCcw className="h-4 w-4 mr-1"/>
              Retry All
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setSelectedRows([])}
            >
              <XCircle className="h-4 w-4 mr-1"/>
              Clear
            </Button>
          </div>
        )}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm">
              <Settings className="h-4 w-4 mr-2"/>
              Options
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Batch Actions</DropdownMenuLabel>
            <DropdownMenuSeparator/>
            <DropdownMenuItem>
              <Download className="h-4 w-4 mr-2"/>
              Export Tasks
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  </CardHeader>
  <CardContent className="px-0">
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow className="border-b hover:bg-transparent bg-muted/50">
            <TableHead className="w-[50px]">
              <Checkbox
                checked={filteredTasks.length > 0 && selectedRows.length === filteredTasks.length}
                onCheckedChange={toggleSelectAll}
                aria-label="Select all"
              />
            </TableHead>
            <TableHead
              className="w-[200px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">MEDIA</TableHead>
            <TableHead
              className="w-[120px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted/80"
              onClick={() => handleSort('profile_id')}
            >
              <div className="flex items-center gap-1">
                PROFILE
                <ArrowUpDown className="h-3 w-3"/>
              </div>
            </TableHead>
            <TableHead
              className="w-[120px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted/80"
              onClick={() => handleSort('progress')}
            >
              <div className="flex items-center gap-1">
                PROGRESS
                <ArrowUpDown className="h-3 w-3"/>
              </div>
            </TableHead>
            <TableHead
              className="w-[120px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted/80"
              onClick={() => handleSort('status')}
            >
              <div className="flex items-center gap-1">
                STATUS
                <ArrowUpDown className="h-3 w-3"/>
              </div>
            </TableHead>
            <TableHead
              className="w-[80px] py-2.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">OUTPUT</TableHead>
            <TableHead
              className="w-[120px] py-2.5 text-right text-[11px] font-semibold uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted/80"
              onClick={() => handleSort('update_time' as any)}
            >
              <div className="flex items-center justify-end gap-1">
                TIME
                <ArrowUpDown className="h-3 w-3"/>
              </div>
            </TableHead>
            <TableHead
              className="w-[180px] py-2.5 text-right text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
              ACTION
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {filteredTasks.map((task) => (
            <TaskRow key={task.id}
                     task={task}
                     onRetry={() => handleRetryTask(task.id)}
                     isRetrying={retryingTaskId === task.id}
                     isSelected={selectedRows.includes(task.id)}
                     onToggleSelect={() => toggleSelectRow(task.id)}
            />
          ))}
        </TableBody>
      </Table>
    </div>
  </CardContent>
</Card>

{/* 分页控件 */}
{data && data.total_filtered > data.page_size && (
  <div className="flex items-center justify-between pt-1 text-xs text-muted-foreground">
    <span className="tabular-nums">
      Page {data.page} of {totalPages} · {data.total_filtered} total
    </span>
    <div className="flex gap-1.5">
      <Button variant="outline" size="sm" className="h-8 text-xs px-3"
              disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
        ← Previous
      </Button>
      <Button variant="outline" size="sm" className="h-8 text-xs px-3"
              disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
        Next →
      </Button>
    </div>
  </div>
)
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

### 6.10 标签页样式（椭圆形按钮）

#### 标准样式
- **按钮类型**：使用 `Button` 组件，`variant` 根据激活状态切换
- **激活状态**：`variant="default"`，使用主按钮样式
- **非激活状态**：`variant="outline"`，使用次要按钮样式
- **形状**：`rounded-full`，实现椭圆形按钮效果
- **大小**：`size="sm"`，`px-4 py-1.5`
- **间距**：按钮之间 `gap-2`
- **一致性**：所有标签页按钮应使用相同的样式，确保在不同页面和场景中保持一致

#### 统一规范
- **位置**：位于状态概览卡片和任务表格之间
- **内容**：包含标签文本和数量，数量使用 `text-xs opacity-60` 样式
- **交互**：点击切换标签，激活状态显示主按钮样式
- **一致性**：所有标签页应使用相同的样式和交互规则，确保在不同页面和场景中保持一致

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
- **搜索和筛选**：位于工具栏下方，单独一行
- **Options 按钮**：位于搜索和筛选区域的右侧，包含上传、添加等操作按钮
- **批量操作**：位于搜索和筛选区域的左侧
- **一致性**：所有工具栏应使用相同的布局规则，确保在不同页面和场景中保持一致

#### 响应式布局
- **桌面端**：水平排列，标题左对齐，Options 按钮右对齐
- **移动端**：垂直排列，标题在上，Options 按钮在下
- **间距**：组件之间保持 `gap-4` 的间距
- **一致性**：所有工具栏应使用相同的响应式布局规则，确保在不同页面和场景中保持一致

#### Options 按钮位置标准
- **参考页面**：/admin/transcoding/status 页面
- **位置**：搜索和筛选区域的右侧，与筛选控件在同一行
- **布局**：使用 `flex items-center gap-2 ml-auto lg:ml-0` 确保按钮在右侧
- **示例**：上传媒体按钮、添加配置文件按钮等应放置在该位置
- **一致性**：所有页面的操作按钮都应遵循此布局规则

## 7. 样式修复说明

### 7.1 修复的问题

1. **卡片颜色问题**：
   - Total Subscribers 卡片：图标和底部色条缺少颜色
   - Active Users 卡片：图标和底部色条缺少颜色
   - 趋势指示颜色：下降趋势显示灰色而非红色，缺少持平趋势指示

2. **布局问题**：
   - TranscodingProfiles 页面：布局被破坏，Dialog 组件结束后缺少正确的闭合标签
   - 上传媒体按钮位置：位置不正确，需要统一到 Options 位置

3. **样式一致性问题**：
   - 各页面搜索和筛选样式不一致
   - 表格显示总记录数问题

### 7.2 修复方案

1. **卡片颜色修复**：
   - 为 StatCard 组件添加颜色映射，确保图标和底部色条正确显示对应颜色
   - 修复趋势指示颜色：上升趋势绿色，下降趋势红色，持平趋势灰色
   - 确保颜色类正确应用到图标上

2. **布局修复**：
   - 修复 TranscodingProfiles 页面的布局问题，确保 Dialog 组件正确闭合
   - 统一上传媒体按钮位置到 Options 位置
   - 确保所有页面的搜索和筛选区域与 Transcoding Status 页面保持一致

3. **样式一致性修复**：
   - 统一所有页面的 Card 样式
   - 统一搜索和筛选组件的样式
   - 确保表格显示总记录数

### 7.3 注意事项

1. **样式与逻辑分离**：
   - 所有修复仅涉及样式变更，不修改任何业务逻辑
   - 确保样式修复不会影响现有功能

2. **颜色规范**：
   - 使用 Tailwind CSS 的默认颜色类（如 pink, cyan, amber 等）
   - 确保颜色类正确应用到图标和底部色条

3. **布局一致性**：
   - 以 Transcoding Status 页面为标准，统一所有页面的布局
   - 确保操作按钮位置一致

4. **构建验证**：
   - 所有修复完成后运行 `npm run build` 确保没有引入错误
   - 确保代码编译成功，无语法错误

### 7.4 修复后的效果

- Total Subscribers 卡片：粉色图标和底部色条
- Active Users 卡片：青色图标和底部色条
- Total Revenue 卡片：琥珀色图标和底部色条
- 所有卡片的趋势指示颜色正确
- 所有页面的布局和样式保持一致
- 上传媒体按钮位置统一到 Options 位置
- 表格显示总记录数

