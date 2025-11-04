# JSON格式Base64图片上传说明

## 概述

已优化**创建文章**和**更新文章**接口，支持在JSON请求中直接发送base64编码的图片数据，字段名为`image`。

## 使用方式

### Content-Type

所有请求都使用：`application/json; charset=utf-8`

### 创建文章（POST）

**接口**：`POST /api/admin/posts`

**请求格式**：
```json
{
  "title": "文章标题",
  "content": "文章内容",
  "image": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ...", // base64编码的图片数据
  "excerpt": "摘要",
  "status": "published",
  "categories": [1, 2],
  "tags": [1, 2]
}
```

**前端示例**：
```javascript
// 方式1：从文件读取为base64
const fileInput = document.querySelector('input[type="file"]');
const file = fileInput.files[0];

const reader = new FileReader();
reader.onload = function(e) {
  const base64Data = e.target.result; // data:image/jpeg;base64,...
  
  fetch('/api/admin/posts', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json; charset=utf-8'
    },
    body: JSON.stringify({
      title: '文章标题',
      content: '文章内容',
      image: base64Data, // 直接发送base64数据
      status: 'published',
      categories: [1, 2],
      tags: [1, 2]
    })
  })
  .then(res => res.json())
  .then(data => {
    console.log('创建成功:', data.post);
    console.log('封面图URL:', data.post.cover_image); // 自动获取的URL
  });
};
reader.readAsDataURL(file);
```

### 更新文章（PUT）

**接口**：`PUT /api/admin/posts/:id`

**请求格式**：
```json
{
  "title": "更新的标题",
  "content": "更新的内容",
  "image": "data:image/png;base64,iVBORw0KGgoAAAANS...", // base64编码的图片数据
  "categories": [1, 3],
  "tags": [2, 4]
}
```

**前端示例**：
```javascript
// 更新文章时上传新图片
const fileInput = document.querySelector('input[type="file"]');
const file = fileInput.files[0];

const reader = new FileReader();
reader.onload = function(e) {
  const base64Data = e.target.result;
  
  fetch(`/api/admin/posts/${postId}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json; charset=utf-8'
    },
    body: JSON.stringify({
      title: '更新的标题',
      content: '更新的内容',
      image: base64Data, // 新图片
      categories: [1, 3],
      tags: [2, 4]
    })
  })
  .then(res => res.json())
  .then(data => {
    console.log('更新成功:', data.post);
    console.log('新封面图URL:', data.post.cover_image);
  });
};
reader.readAsDataURL(file);
```

## 字段说明

### image字段

- **类型**：`string`
- **格式**：base64编码的图片数据
- **支持格式**：
  - 纯base64字符串：`iVBORw0KGgoAAAANS...`
  - Data URL格式：`data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ...`
  - Data URL格式（PNG）：`data:image/png;base64,iVBORw0KGgoAAAANS...`

### cover_image字段（可选）

- **类型**：`string`
- **说明**：如果提供了`image`字段，`cover_image`会被忽略；如果没有`image`字段，可以使用`cover_image`传入已有的图片URL

### 优先级

1. 如果提供了`image`字段（base64数据）→ 自动解码保存，生成URL
2. 如果没有`image`字段，但有`cover_image`字段 → 直接使用URL
3. 如果都没有 → `cover_image`为空

## 支持的图片格式

- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- WEBP (.webp)
- SVG (.svg)

## 自动处理

后端会自动：
1. 检测base64数据格式（是否包含data URL前缀）
2. 解码base64数据
3. 根据文件头自动识别图片格式
4. 生成唯一文件名
5. 保存到`uploads/images/`目录
6. 生成访问URL并填充到`cover_image`字段

## 完整示例

### React组件示例

```javascript
import { useState } from 'react';

const PostEditor = () => {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [imageFile, setImageFile] = useState(null);
  const [imagePreview, setImagePreview] = useState('');
  const [saving, setSaving] = useState(false);

  const handleImageChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      setImageFile(file);
      
      // 预览图片
      const reader = new FileReader();
      reader.onload = (e) => {
        setImagePreview(e.target.result);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleSave = async (postId = null) => {
    setSaving(true);
    
    try {
      const requestData = {
        title,
        content,
        status: 'published',
        categories: [1],
        tags: [1]
      };
      
        // 如果有图片，转换为base64
        if (imageFile) {
          const base64Data = await new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = (e) => resolve(e.target.result);
            reader.onerror = reject;
            reader.readAsDataURL(imageFile);
          });
          requestData.image = base64Data;
        }
      
      const url = postId 
        ? `/api/admin/posts/${postId}` 
        : '/api/admin/posts';
      const method = postId ? 'PUT' : 'POST';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json; charset=utf-8'
        },
        body: JSON.stringify(requestData)
      });
      
      const data = await response.json();
      console.log('保存成功:', data.post);
      console.log('封面图URL:', data.post.cover_image);
      
    } catch (error) {
      console.error('保存失败:', error);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <input
        type="text"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="文章标题"
      />
      
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="文章内容"
      />
      
      <input
        type="file"
        accept="image/*"
        onChange={handleImageChange}
      />
      
      {imagePreview && (
        <img src={imagePreview} alt="预览" style={{maxWidth: '200px'}} />
      )}
      
      <button onClick={() => handleSave()} disabled={saving}>
        {saving ? '保存中...' : '保存'}
      </button>
    </div>
  );
};
```

### Vue组件示例

```vue
<template>
  <div>
    <input v-model="title" placeholder="文章标题" />
    <textarea v-model="content" placeholder="文章内容" />
    <input type="file" @change="handleImageChange" accept="image/*" />
    <img v-if="imagePreview" :src="imagePreview" alt="预览" style="max-width: 200px" />
    <button @click="handleSave" :disabled="saving">
      {{ saving ? '保存中...' : '保存' }}
    </button>
  </div>
</template>

<script>
export default {
  data() {
    return {
      title: '',
      content: '',
      imageFile: null,
      imagePreview: '',
      saving: false
    };
  },
  methods: {
    handleImageChange(e) {
      const file = e.target.files[0];
      if (file) {
        this.imageFile = file;
        const reader = new FileReader();
        reader.onload = (e) => {
          this.imagePreview = e.target.result;
        };
        reader.readAsDataURL(file);
      }
    },
    async handleSave() {
      this.saving = true;
      
      try {
        const requestData = {
          title: this.title,
          content: this.content,
          status: 'published',
          categories: [1],
          tags: [1]
        };
        
        // 如果有图片，转换为base64
        if (this.imageFile) {
          requestData.image = await new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = (e) => resolve(e.target.result);
            reader.onerror = reject;
            reader.readAsDataURL(this.imageFile);
          });
        }
        
        const response = await fetch('/api/admin/posts', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${this.token}`,
            'Content-Type': 'application/json; charset=utf-8'
          },
          body: JSON.stringify(requestData)
        });
        
        const data = await response.json();
        console.log('保存成功:', data.post);
        console.log('封面图URL:', data.post.cover_image);
        
      } catch (error) {
        console.error('保存失败:', error);
      } finally {
        this.saving = false;
      }
    }
  }
};
</script>
```

## 响应格式

### 创建文章响应

```json
{
  "post": {
    "id": 1,
    "title": "文章标题",
    "content": "文章内容",
    "cover_image": "/uploads/images/20241103-abc12345-image.jpg", // 自动生成的URL
    "status": "published",
    ...
  }
}
```

### 更新文章响应

```json
{
  "post": {
    "id": 1,
    "title": "更新的标题",
    "content": "更新的内容",
    "cover_image": "/uploads/images/20241103-def45678-new-image.jpg", // 新图片URL
    ...
  },
  "categories": [...],
  "tags": [...]
}
```

## 注意事项

### 1. 文件大小限制

- Base64编码会使文件大小增加约33%
- 建议图片大小不超过5MB（原始文件）
- 大图片建议先压缩再转换为base64

### 2. 性能考虑

- Base64编码会增加请求体大小
- 对于大文件，建议使用FormData方式（multipart/form-data）
- 小图片（< 1MB）使用base64更方便

### 3. 数据格式

- 支持纯base64字符串：`iVBORw0KGgoAAAANS...`
- 支持Data URL格式：`data:image/jpeg;base64,/9j/4AAQ...`
- 后端会自动识别并处理

### 4. 图片格式检测

后端会根据文件头自动识别图片格式：
- JPEG: `FF D8`
- PNG: `89 50 4E 47`
- GIF: `47 49 46`
- WEBP: `RIFF ... WEBP`

## 错误处理

### 常见错误

**Base64解码失败**：
```json
{
  "error": "图片保存失败: base64解码失败: ..."
}
```

**无效的图片格式**：
后端会自动检测图片格式，如果不是支持的格式会返回错误。

**文件保存失败**：
```json
{
  "error": "图片保存失败: 创建文件失败: ..."
}
```

## 总结

✅ **支持JSON格式**：`application/json; charset=utf-8`  
✅ **支持base64图片**：`image`字段接收base64编码的图片数据  
✅ **自动处理**：自动解码、保存、生成URL  
✅ **向后兼容**：仍然支持`cover_image`字段（URL字符串）  
✅ **灵活使用**：可以只传`image`（base64），也可以只传`cover_image`（URL）  

现在前端可以在JSON请求中直接发送base64编码的图片数据（字段名为`image`），后端会自动处理！🎉

