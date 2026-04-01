// Mock API Server - 提供模拟数据让前端可以预览
// 运行方式: node mock-server.js

const http = require('http');
const {parse} = require('url');
const crypto = require('crypto');

const PORT = 5000;
const ADMIN_PORT = 9090;

// 生成 UUID 的简单方法
function uuid() {
    return crypto.randomUUID();
}

const now = new Date().toISOString();

// Mock 数据
const users = new Map();
const media = new Map();
const categories = new Map();
const tags = new Map();
const comments = new Map();
const likes = new Map();
const favorites = new Map();
const playlists = new Map();

// 初始化数据
function initData() {
    // Admin 用户
    const adminId = uuid();
    users.set(adminId, {
        id: adminId,
        username: 'admin',
        email: 'admin@origcms.local',
        nickname: 'Administrator',
        avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin',
        role: 'admin',
        status: 'active',
        created_at: now,
        password_hash: '$2a$10$fakehash'
    });
    users.set('admin', users.get(adminId));

    // 普通用户
    const userId = uuid();
    users.set(userId, {
        id: userId,
        username: 'testuser',
        email: 'test@origcms.local',
        nickname: 'Test User',
        avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=test',
        role: 'user',
        status: 'active',
        created_at: now,
        password_hash: '$2a$10$fakehash'
    });
    users.set('testuser', users.get(userId));

    // 分类
    const cats = [
        {name: 'Technology', slug: 'technology', description: 'Tech videos', order: 1},
        {name: 'Music', slug: 'music', description: 'Music videos', order: 2},
        {name: 'Gaming', slug: 'gaming', description: 'Gaming content', order: 3},
        {name: 'Education', slug: 'education', description: 'Educational videos', order: 4},
        {name: 'Entertainment', slug: 'entertainment', description: 'Fun videos', order: 5},
    ];
    cats.forEach(c => {
        const id = uuid();
        categories.set(id, {...c, id, created_at: now});
    });

    // 标签
    const tagList = [
        {name: 'Tutorial', slug: 'tutorial'},
        {name: 'Review', slug: 'review'},
        {name: 'News', slug: 'news'},
        {name: 'Comedy', slug: 'comedy'},
        {name: 'Documentary', slug: 'documentary'},
    ];
    tagList.forEach(t => {
        const id = uuid();
        tags.set(id, {...t, id, created_at: now});
    });

    // 媒体
    const mediaList = [
        {
            title: 'Introduction to Go Programming',
            description: 'Learn the basics of Go programming language in this comprehensive tutorial.',
            type: 'video',
            url: 'https://storage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4',
            thumbnail: 'https://storage.googleapis.com/gtv-videos-bucket/sample/images/BigBuckBunny.jpg',
            size: 158008374,
            duration: 596,
            views: 1250,
            likes: 89,
            tags: ['Tutorial', 'News']
        },
        {
            title: 'Web Development Best Practices',
            description: 'Tips and tricks for building modern web applications.',
            type: 'video',
            url: 'https://storage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4',
            thumbnail: 'https://storage.googleapis.com/gtv-videos-bucket/sample/images/ElephantsDream.jpg',
            size: 89846523,
            duration: 653,
            views: 856,
            likes: 45,
            tags: ['Tutorial']
        },
        {
            title: 'Nature Documentary',
            description: 'Explore the beauty of nature in this stunning documentary.',
            type: 'video',
            url: 'https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4',
            thumbnail: 'https://storage.googleapis.com/gtv-videos-bucket/sample/images/ForBiggerBlazes.jpg',
            size: 2345678,
            duration: 15,
            views: 234,
            likes: 12,
            tags: ['Documentary']
        },
        {
            title: 'Relaxing Piano Music',
            description: 'Beautiful piano compositions for relaxation and study.',
            type: 'audio',
            url: 'https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3',
            thumbnail: '',
            size: 12345678,
            duration: 375,
            views: 567,
            likes: 34,
            tags: ['Music']
        },
        {
            title: 'Amazing Photography',
            description: 'Stunning photographs from around the world.',
            type: 'image',
            url: 'https://picsum.photos/800/600',
            thumbnail: 'https://picsum.photos/400/300',
            size: 2345678,
            duration: 0,
            views: 445,
            likes: 23,
            tags: ['News']
        },
        {
            title: 'JavaScript Advanced Patterns',
            description: 'Deep dive into advanced JavaScript design patterns and techniques.',
            type: 'video',
            url: 'https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4',
            thumbnail: 'https://storage.googleapis.com/gtv-videos-bucket/sample/images/ForBiggerEscapes.jpg',
            size: 12345678,
            duration: 180,
            views: 678,
            likes: 56,
            tags: ['Tutorial', 'Review']
        },
        {
            title: 'Cooking Masterclass',
            description: 'Learn to cook delicious meals from a professional chef.',
            type: 'video',
            url: 'https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerFun.mp4',
            thumbnail: 'https://storage.googleapis.com/gtv-videos-bucket/sample/images/ForBiggerFun.jpg',
            size: 45678901,
            duration: 420,
            views: 890,
            likes: 78,
            tags: ['Education']
        },
        {
            title: 'Stand-up Comedy Special',
            description: 'Hilarious comedy performance that will make you laugh out loud.',
            type: 'video',
            url: 'https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerJoyrides.mp4',
            thumbnail: 'https://storage.googleapis.com/gtv-videos-bucket/sample/images/ForBiggerJoyrides.jpg',
            size: 34567890,
            duration: 360,
            views: 1100,
            likes: 156,
            tags: ['Comedy', 'Entertainment']
        }
    ];

    const catArray = Array.from(categories.values());
    mediaList.forEach((m, i) => {
        const id = uuid();
        media.set(id, {
            ...m,
            id,
            status: 'published',
            user_id: i % 2 === 0 ? adminId : userId,
            category_id: catArray[i % catArray.length].id,
            created_at: now,
            updated_at: now
        });
    });
}

initData();

// Mock Token (简化版)
const MOCK_SECRET = 'mock-secret-key';

function createToken(payload) {
    // 简化: 直接 base64 编码 JSON，不做加密（仅用于 mock）
    return Buffer.from(JSON.stringify({...payload, exp: Date.now() + 86400000})).toString('base64');
}

function verifyToken(token) {
    try {
        const decoded = JSON.parse(Buffer.from(token, 'base64').toString());
        if (decoded.exp < Date.now()) return null;
        return decoded;
    } catch {
        return null;
    }
}

// 解析 JSON Body
function parseBody(req) {
    return new Promise((resolve, reject) => {
        let body = '';
        req.on('data', chunk => body += chunk);
        req.on('end', () => {
            try {
                resolve(body ? JSON.parse(body) : {});
            } catch (e) {
                reject(e);
            }
        });
        req.on('error', reject);
    });
}

// 发送 JSON 响应
function json(res, data, status = 200) {
    res.writeHead(status, {
        'Content-Type': 'application/json',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, PATCH, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization'
    });
    res.end(JSON.stringify(data));
}

// CORS 预检
function handleCORS(req, res) {
    res.writeHead(204, {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, PATCH, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization'
    });
    res.end();
}

// 路由处理
async function handleRequest(req, res) {
    const parsedUrl = parse(req.url, true);
    const pathname = parsedUrl.pathname;
    const method = req.method;

    // CORS 预检
    if (method === 'OPTIONS') {
        return handleCORS(req, res);
    }

    console.log(`${method} ${pathname}`);

    // 公开接口
    if (pathname === '/api/v1/health') {
        return json(res, {status: 'ok', time: now});
    }

    if (pathname === '/api/v1/stats') {
        let storage = 0;
        media.forEach(m => storage += m.size);
        const storageStr = storage > 1024 * 1024 * 1024
            ? (storage / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
            : (storage / (1024 * 1024)).toFixed(1) + ' MB';

        return json(res, {
            users: users.size - 2, // 减去 username key
            media: media.size,
            content: categories.size + tags.size,
            storage: storageStr,
            views: Array.from(media.values()).reduce((sum, m) => sum + m.views, 0)
        });
    }

    // Categories API
    if (pathname === '/api/v1/categories' && method === 'GET') {
        return json(res, Array.from(categories.values()));
    }

    if (pathname.startsWith('/api/v1/categories/') && method === 'GET') {
        const id = pathname.split('/').pop();
        const cat = categories.get(id);
        if (!cat) return json(res, {error: 'Not found'}, 404);
        return json(res, cat);
    }

    // Tags API
    if (pathname === '/api/v1/tags' && method === 'GET') {
        return json(res, Array.from(tags.values()));
    }

    // Media API
    if (pathname === '/api/v1/media' && method === 'GET') {
        const parsed = parsedUrl.query;
        let list = Array.from(media.values());
        if (parsed.category_id) {
            list = list.filter(m => m.category_id === parsed.category_id);
        }
        if (parsed.type) {
            list = list.filter(m => m.type === parsed.type);
        }
        if (parsed.status === undefined) {
            list = list.filter(m => m.status === 'active');
        }
        const page = parseInt(parsed.page
        as
        string
    ) ||
        1;
        const pageSize = parseInt(parsed.page_size
        as
        string
    ) ||
        20;
        const total = list.length;
        list = list.slice((page - 1) * pageSize, page * pageSize);
        return json(res, {list, total, page, page_size: pageSize});
    }

    if (pathname.startsWith('/api/v1/media/') && method === 'GET') {
        const id = pathname.split('/').pop();
        const m = media.get(id);
        if (!m) return json(res, {error: 'Not found'}, 404);
        // 增加浏览数
        media.set(id, {...m, views: m.views + 1});
        return json(res, m);
    }

    // Comments API
    if (pathname === '/api/v1/comments' && method === 'GET') {
        const mediaId = parsedUrl.query.media_id;
        let list = Array.from(comments.values());
        if (mediaId) {
            list = list.filter(c => c.media_id === mediaId);
        }
        return json(res, {list});
    }

    // 认证接口
    if (pathname === '/api/v1/auth/signin' && method === 'POST') {
        const body = await parseBody(req);
        const user = users.get(body.username);
        if (!user) {
            return json(res, {error: 'Invalid credentials'}, 401);
        }
        const token = createToken({user_id: user.id, username: user.username, role: user.role});
        const refreshToken = createToken({user_id: user.id, type: 'refresh'});
        return json(res, {
            access_token: token,
            refresh_token: refreshToken,
            token_type: 'Bearer',
            expires_in: 86400,
            user: {
                id: user.id,
                username: user.username,
                email: user.email,
                nickname: user.nickname,
                avatar: user.avatar,
                role: user.role
            }
        });
    }

    if (pathname === '/api/v1/auth/signup' && method === 'POST') {
        const body = await parseBody(req);
        if (users.get(body.username)) {
            return json(res, {error: 'Username already exists'}, 409);
        }
        const id = uuid();
        const user = {
            id,
            username: body.username,
            email: body.email,
            nickname: body.nickname || body.username,
            avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${body.username}`,
            role: 'user',
            status: 'active',
            created_at: now
        };
        users.set(id, user);
        users.set(body.username, user);
        return json(res, user, 201);
    }

    if (pathname === '/api/v1/auth/refresh' && method === 'POST') {
        const body = await parseBody(req);
        const decoded = verifyToken(body.refresh_token);
        if (!decoded) {
            return json(res, {error: 'Invalid refresh token'}, 401);
        }
        const token = createToken({user_id: decoded.user_id, username: decoded.username, role: decoded.role});
        return json(res, {access_token: token, token_type: 'Bearer', expires_in: 86400});
    }

    // 验证 Token 中间件
    const authHeader = req.headers.authorization;
    let userId = null;
    let userRole = null;
    if (authHeader && authHeader.startsWith('Bearer ')) {
        const token = authHeader.slice(7);
        const decoded = verifyToken(token);
        if (decoded) {
            userId = decoded.user_id;
            userRole = decoded.role;
        }
    }

    // 媒体列表
    if (pathname === '/api/v1/media' && method === 'GET') {
        const status = parsedUrl.query.status || 'published';
        const type = parsedUrl.query.type;
        const categoryId = parsedUrl.query.category_id;
        const keyword = parsedUrl.query.keyword;
        const page = parseInt(parsedUrl.query.page) || 1;
        const pageSize = parseInt(parsedUrl.query.page_size) || 20;

        let result = Array.from(media.values()).filter(m => {
            if (status && m.status !== status) return false;
            if (type && m.type !== type) return false;
            if (categoryId && m.category_id !== categoryId) return false;
            if (keyword) {
                const kw = keyword.toLowerCase();
                if (!m.title.toLowerCase().includes(kw) && !m.description.toLowerCase().includes(kw)) {
                    return false;
                }
            }
            return true;
        });

        const total = result.length;
        const start = (page - 1) * pageSize;
        result = result.slice(start, start + pageSize);

        return json(res, {list: result, total, page, page_size: pageSize});
    }

    // 媒体详情
    const mediaMatch = pathname.match(/^\/api\/v1\/media\/(.+)$/);
    if (mediaMatch) {
        const id = mediaMatch[1];
        if (method === 'GET') {
            const m = media.get(id);
            if (!m) return json(res, {error: 'Media not found'}, 404);
            m.views++;
            media.set(id, m);
            return json(res, m);
        }
    }

    // 分类列表
    if (pathname === '/api/v1/categories' && method === 'GET') {
        return json(res, {list: Array.from(categories.values()), total: categories.size});
    }

    // 标签列表
    if (pathname === '/api/v1/tags' && method === 'GET') {
        return json(res, {list: Array.from(tags.values()), total: tags.size});
    }

    // 搜索
    if (pathname === '/api/v1/search' && method === 'GET') {
        const q = parsedUrl.query.q || '';
        const qLower = q.toLowerCase();
        const mediaResults = Array.from(media.values()).filter(m =>
            m.status === 'published' &&
            (q === '' || m.title.toLowerCase().includes(qLower) || m.description.toLowerCase().includes(qLower))
        );
        const catResults = Array.from(categories.values()).filter(c =>
            q === '' || c.name.toLowerCase().includes(qLower) || c.description?.toLowerCase().includes(qLower)
        );
        const tagResults = Array.from(tags.values()).filter(t =>
            q === '' || t.name.toLowerCase().includes(qLower)
        );
        return json(res, {media: mediaResults, categories: catResults, tags: tagResults});
    }

    // 需要认证的接口
    if (!userId) {
        return json(res, {error: 'Authorization required'}, 401);
    }

    // 用户信息
    if (pathname === '/api/v1/user/me' && method === 'GET') {
        const user = users.get(userId);
        if (!user) return json(res, {error: 'User not found'}, 404);
        return json(res, {
            id: user.id,
            username: user.username,
            email: user.email,
            nickname: user.nickname,
            avatar: user.avatar,
            role: user.role,
            status: user.status,
            created_at: user.created_at
        });
    }

    // 点赞
    const likeMatch = pathname.match(/^\/api\/v1\/media\/(.+)\/like$/);
    if (likeMatch) {
        const mediaId = likeMatch[1];
        if (method === 'POST') {
            const id = uuid();
            likes.set(id, {id, media_id: mediaId, user_id: userId, created_at: now});
            const m = media.get(mediaId);
            if (m) {
                m.likes++;
                media.set(mediaId, m);
            }
            return json(res, {id, media_id: mediaId, user_id: userId, created_at: now}, 201);
        }
        if (method === 'DELETE') {
            for (const [id, l] of likes) {
                if (l.media_id === mediaId && l.user_id === userId) {
                    likes.delete(id);
                    const m = media.get(mediaId);
                    if (m && m.likes > 0) {
                        m.likes--;
                        media.set(mediaId, m);
                    }
                    return json(res, {message: 'Unliked'});
                }
            }
            return json(res, {error: 'Like not found'}, 404);
        }
    }

    // 收藏
    const favMatch = pathname.match(/^\/api\/v1\/media\/(.+)\/favorite$/);
    if (favMatch) {
        const mediaId = favMatch[1];
        if (method === 'POST') {
            const id = uuid();
            favorites.set(id, {id, media_id: mediaId, user_id: userId, created_at: now});
            return json(res, {id, media_id: mediaId, user_id: userId, created_at: now}, 201);
        }
        if (method === 'DELETE') {
            for (const [id, f] of favorites) {
                if (f.media_id === mediaId && f.user_id === userId) {
                    favorites.delete(id);
                    return json(res, {message: 'Unfavorited'});
                }
            }
            return json(res, {error: 'Favorite not found'}, 404);
        }
    }

    // 评论
    const commentMediaMatch = pathname.match(/^\/api\/v1\/media\/(.+)\/comments$/);
    if (commentMediaMatch) {
        const mediaId = commentMediaMatch[1];
        if (method === 'POST') {
            const body = await parseBody(req);
            const id = uuid();
            const comment = {
                id,
                media_id: mediaId,
                user_id: userId,
                content: body.content,
                parent_id: body.parent_id || '',
                likes: 0,
                created_at: now
            };
            comments.set(id, comment);
            return json(res, comment, 201);
        }
    }

    if (pathname === '/api/v1/comments' && method === 'GET') {
        const mediaId = parsedUrl.query.media_id;
        const result = Array.from(comments.values()).filter(c => !mediaId || c.media_id === mediaId);
        return json(res, {list: result, total: result.length});
    }

    // 管理后台接口
    if (pathname === '/api/v1/admin/users' && method === 'GET') {
        if (userRole !== 'admin') return json(res, {error: 'Admin only'}, 403);
        const result = Array.from(users.values())
            .filter(u => u.username) // 过滤掉 username key
            .map(u => ({
                id: u.id,
                username: u.username,
                email: u.email,
                nickname: u.nickname,
                avatar: u.avatar,
                role: u.role,
                status: u.status,
                created_at: u.created_at
            }));
        return json(res, {list: result, total: result.length});
    }

    if (pathname === '/api/v1/admin/media' && method === 'GET') {
        if (userRole !== 'admin') return json(res, {error: 'Admin only'}, 403);
        const page = parseInt(parsedUrl.query.page) || 1;
        const pageSize = parseInt(parsedUrl.query.page_size) || 20;
        const result = Array.from(media.values());
        const total = result.length;
        return json(res, {
            list: result.slice((page - 1) * pageSize, page * pageSize),
            total,
            page,
            page_size: pageSize
        });
    }

    // 404
    json(res, {error: 'Not found'}, 404);
}

const server = http.createServer(handleRequest);

server.listen(PORT, () => {
    console.log(`
========================================
  Mock API Server (OrigCMS)
  Listen: http://localhost:${PORT}
  Admin: admin / admin123
  User:  testuser / user123
========================================
  `);
});
