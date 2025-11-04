# 文件上传API文档

## 概述

管理后台文件上传接口，支持单文件、批量文件、图片上传等功能。所有接口需要管理员权限（Bearer Token认证）。

## 基础路径

所有接口的基础路径：`/api/admin/upload`

## 认证

所有接口都需要在请求头中携带管理员Token：

```
Authorization: Bearer {your-admin-token}
```

## 上传目录结构

```
uploads/
├── images/    # 图片文件
└── files/     # 其他文件
```

文件访问URL：`http://your-domain/uploads/images/filename.jpg`

---

## API列表

### 1. 上传单个文件

**接口**：`POST /api/admin/upload/file`

**功能**：上传单个文件（图片、文档、压缩包等）

**Content-Type**：`multipart/form-data`

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | File | 是 | 上传的文件 |

**支持的文件类型**：
- 图片：`.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, `.svg`
- 文档：`.pdf`, `.doc`, `.docx`
- 压缩包：`.zip`, `.rar`

**文件大小限制**：10MB

**请求示例**：

```bash
curl -X POST http://localhost:8080/api/admin/upload/file \
  -H "Authorization: Bearer your-token" \
  -F "file=@/path/to/file.jpg"
```

**前端示例**：
```javascript
const formData = new FormData();
formData.append('file', fileInput.files[0]);

const response = await fetch('/api/admin/upload/file', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});

const data = await response.json();
console.log(data.url); // 文件访问URL
```

**响应示例**：
```json
{
  "url": "/uploads/images/20241102-abc12345-文件名.jpg",
  "path": "uploads/images/20241102-abc12345-文件名.jpg",
  "filename": "20241102-abc12345-文件名.jpg",
  "original": "原始文件名.jpg",
  "size": 123456,
  "type": "image/jpeg"
}
```

---

### 2. 上传图片（专用）

**接口**：`POST /api/admin/upload/image`

**功能**：专门用于上传图片文件，自动验证图片格式

**Content-Type**：`multipart/form-data`

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| image | File | 是 | 图片文件（也支持file字段名） |

**支持格式**：`.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, `.svg`

**文件大小限制**：5MB

**请求示例**：

```bash
curl -X POST http://localhost:8080/api/admin/upload/image \
  -H "Authorization: Bearer your-token" \
  -F "image=@/path/to/image.jpg"
```

**前端示例**：
```javascript
const formData = new FormData();
formData.append('image', imageFile);

const response = await fetch('/api/admin/upload/image', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});

const data = await response.json();
// 使用 data.url 作为文章封面图
```

**响应示例**：
```json
{
  "url": "/uploads/images/20241102-abc12345-image.jpg",
  "path": "uploads/images/20241102-abc12345-image.jpg",
  "filename": "20241102-abc12345-image.jpg",
  "original": "image.jpg",
  "size": 45678,
  "type": "image/jpeg"
}
```

---

### 3. 批量上传文件

**接口**：`POST /api/admin/upload/files`

**功能**：一次上传多个文件

**Content-Type**：`multipart/form-data`

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| files | File[] | 是 | 文件数组 |

**文件大小限制**：总大小50MB，单个文件10MB

**请求示例**：

```bash
curl -X POST http://localhost:8080/api/admin/upload/files \
  -H "Authorization: Bearer your-token" \
  -F "files=@file1.jpg" \
  -F "files=@file2.png"
```

**前端示例**：
```javascript
const formData = new FormData();
// 添加多个文件
for (let file of fileInput.files) {
  formData.append('files', file);
}

const response = await fetch('/api/admin/upload/files', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});

const data = await response.json();
console.log(data.files); // 上传成功的文件列表
console.log(data.errors); // 上传失败的文件错误（如果有）
```

**响应示例**：
```json
{
  "success": 2,
  "total": 3,
  "files": [
    {
      "url": "/uploads/images/20241102-abc123-file1.jpg",
      "path": "uploads/images/20241102-abc123-file1.jpg",
      "filename": "20241102-abc123-file1.jpg",
      "original": "file1.jpg",
      "size": 123456,
      "type": "image/jpeg"
    },
    {
      "url": "/uploads/images/20241102-def456-file2.png",
      "path": "uploads/images/20241102-def456-file2.png",
      "filename": "20241102-def456-file2.png",
      "original": "file2.png",
      "size": 78901,
      "type": "image/png"
    }
  ],
  "errors": [
    "file3.exe: 不支持的文件类型"
  ]
}
```

---

### 4. 删除文件

**接口**：`DELETE /api/admin/upload/file`

**功能**：删除已上传的文件

**请求参数**（Query参数）：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 文件路径或URL |

**请求示例**：

```bash
curl -X DELETE "http://localhost:8080/api/admin/upload/file?path=uploads/images/file.jpg" \
  -H "Authorization: Bearer your-token"
```

**前端示例**：
```javascript
const response = await fetch(
  `/api/admin/upload/file?path=${encodeURIComponent(filePath)}`,
  {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  }
);

const data = await response.json();
```

**响应示例**：
```json
{
  "message": "删除成功"
}
```

---

### 5. 获取文件列表

**接口**：`GET /api/admin/upload/files`

**功能**：获取已上传的文件列表（可选功能）

**请求参数**（Query参数）：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 否 | 文件类型筛选：image/file/all（默认all） |
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |

**请求示例**：

```bash
curl "http://localhost:8080/api/admin/upload/files?type=image&page=1&page_size=20" \
  -H "Authorization: Bearer your-token"
```

**响应示例**：
```json
{
  "files": [
    {
      "name": "20241102-abc123-image.jpg",
      "path": "uploads/images/20241102-abc123-image.jpg",
      "url": "/uploads/images/20241102-abc123-image.jpg",
      "size": 123456,
      "modified": "2024-11-02T10:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

---

## 文件命名规则

上传的文件会自动重命名为唯一文件名：

格式：`{日期}-{唯一ID}-{原文件名}.{扩展名}`

示例：
- 原文件名：`我的图片.jpg`
- 新文件名：`20241102-abc12345-wo-de-tu-pian.jpg`

**优势**：
- 避免文件名冲突
- 包含日期便于管理
- 保留部分原文件名便于识别

---

## 文件访问

上传成功后，文件可以通过以下URL访问：

```
http://your-domain/uploads/images/filename.jpg
http://your-domain/uploads/files/document.pdf
```

**静态文件服务**：已在路由中配置 `r.Static("/uploads", "./uploads")`

---

## 前端集成示例

### React/Vue组件示例

```javascript
// 图片上传组件
const ImageUpload = () => {
  const [uploading, setUploading] = useState(false);
  const [imageUrl, setImageUrl] = useState('');

  const handleUpload = async (file) => {
    setUploading(true);
    const formData = new FormData();
    formData.append('image', file);

    try {
      const response = await fetch('/api/admin/upload/image', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      });

      const data = await response.json();
      setImageUrl(data.url);
      // 使用 data.url 更新文章封面图字段
    } catch (error) {
      console.error('上传失败:', error);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <input
        type="file"
        accept="image/*"
        onChange={(e) => handleUpload(e.target.files[0])}
        disabled={uploading}
      />
      {imageUrl && <img src={imageUrl} alt="预览" />}
      {uploading && <p>上传中...</p>}
    </div>
  );
};
```

### 使用富文本编辑器（如TinyMCE、CKEditor）

```javascript
// TinyMCE图片上传配置
tinymce.init({
  selector: '#content',
  images_upload_handler: async (blobInfo, progress) => {
    const formData = new FormData();
    formData.append('image', blobInfo.blob(), blobInfo.filename());

    const response = await fetch('/api/admin/upload/image', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      },
      body: formData
    });

    const data = await response.json();
    return data.url; // 返回图片URL供编辑器使用
  }
});
```

---

## 错误处理

### 常见错误

**文件大小超限**：
```json
{
  "error": "文件上传失败: http: request body too large"
}
```

**不支持的文件类型**：
```json
{
  "error": "不支持的文件类型: .exe"
}
```

**未授权**：
```json
{
  "error": "未提供认证信息"
}
```

**权限不足**：
```json
{
  "error": "需要管理员权限"
}
```

---

## 安全建议

### 开发环境
- ✅ 文件大小限制（图片5MB，文件10MB）
- ✅ 文件类型白名单
- ✅ 自动生成唯一文件名

### 生产环境（建议增强）

1. **文件类型验证**：验证文件真实类型（不只是扩展名）
2. **病毒扫描**：集成病毒扫描服务
3. **图片压缩**：自动压缩大图片
4. **CDN集成**：将文件上传到CDN
5. **访问控制**：限制文件访问权限
6. **存储配额**：限制用户存储空间
7. **文件清理**：定期清理未使用的文件

---

## 配置说明

### 上传目录

默认配置在 `service/upload_service.go` 中：

```go
const (
    UploadDir      = "./uploads"
    ImageUploadDir = "./uploads/images"
    FileUploadDir  = "./uploads/files"
    PublicURL      = "/uploads"
)
```

### 文件大小限制

- 单个文件：10MB
- 图片：5MB
- 批量上传总大小：50MB

如需修改，可在控制器中调整 `maxSize` 常量。

---

## 使用流程

### 文章封面图上传流程

1. **上传图片**：
   ```javascript
   POST /api/admin/upload/image
   ```

2. **获取返回的URL**：
   ```json
   {
     "url": "/uploads/images/20241102-abc123-image.jpg"
   }
   ```

3. **更新文章**：
   ```javascript
   PUT /api/admin/posts/:id
   {
     "cover_image": "/uploads/images/20241102-abc123-image.jpg"
   }
   ```

### 文章内容图片上传流程（富文本编辑器）

1. **在编辑器中插入图片**
2. **编辑器自动调用上传接口**
3. **获取图片URL并插入到内容中**
4. **保存文章内容（包含图片URL）**

---

## 总结

✅ **功能完整**：单文件、批量、图片专用上传
✅ **安全可靠**：文件类型验证、大小限制
✅ **易于使用**：返回完整URL，直接可用
✅ **自动管理**：唯一文件名、目录自动创建
✅ **静态服务**：自动提供文件访问服务

所有接口已实现并测试通过，可以直接使用！🎉

