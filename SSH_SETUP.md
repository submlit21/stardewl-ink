# SSH Key 配置指南

## 你的SSH Key信息
- **公钥指纹**: `SHA256:U+p3AtWKTpyJ/IwpZi72RE4KewTmRdsr6blydWpyc0I`
- **GitHub用户名**: `submlit21`
- **仓库地址**: `git@github.com:submlit21/stardewl-ink.git`

## 配置步骤

### 1. 检查SSH Key是否已添加到GitHub
```bash
# 测试SSH连接
ssh -T git@github.com

# 如果看到类似以下信息，说明已配置成功：
# Hi submlit21! You've successfully authenticated, but GitHub does not provide shell access.
```

### 2. 如果未配置，添加SSH Key到GitHub
1. 访问 https://github.com/settings/keys
2. 点击"New SSH key"
3. 标题: 任意名称（如"My Laptop"）
4. Key类型: Authentication Key
5. 粘贴你的SSH公钥内容

### 3. 使用SSH推送代码
```bash
# 进入项目目录
cd stardewl-ink

# 设置SSH remote
git remote remove origin
git remote add origin git@github.com:submlit21/stardewl-ink.git

# 推送代码
git push -u origin main
```

## 故障排除

### SSH连接失败
```bash
# 1. 检查SSH agent是否运行
eval "$(ssh-agent -s)"

# 2. 添加私钥到agent
ssh-add ~/.ssh/id_rsa

# 3. 测试连接
ssh -T git@github.com
```

### 权限被拒绝
```bash
# 检查公钥是否已添加到GitHub
curl -s https://github.com/submlit21.keys

# 如果看到你的公钥，说明已添加
# 如果没有，需要添加公钥到GitHub
```

### 主机密钥验证失败
```bash
# 添加GitHub到known_hosts
ssh-keyscan github.com >> ~/.ssh/known_hosts
```

## 备选方案：使用HTTPS + Token

如果SSH配置有问题，可以使用GitHub Token：

### 1. 创建GitHub Token
1. 访问 https://github.com/settings/tokens
2. 点击"Generate new token"
3. 选择"repo"权限
4. 生成并复制token

### 2. 使用Token推送
```bash
# 使用HTTPS URL + token
git remote add origin https://<YOUR_TOKEN>@github.com/submlit21/stardewl-ink.git

# 或使用credential helper
git config --global credential.helper store
# 第一次推送时会提示输入用户名和token
```

## 验证配置

### 测试Git配置
```bash
# 检查remote配置
git remote -v

# 检查用户配置
git config --list | grep user

# 测试推送权限
git push --dry-run origin main
```

### 检查SSH配置
```bash
# 查看SSH配置
cat ~/.ssh/config

# 测试GitHub连接
ssh -vT git@github.com
```

## 多设备同步

如果你需要在多台设备上工作：

### 1. 生成新的SSH Key（每台设备）
```bash
ssh-keygen -t ed25519 -C "your-email@example.com"
```

### 2. 添加到GitHub
- 每台设备的公钥都需要添加到GitHub
- 可以使用相同的邮箱或设备名作为标识

### 3. 配置SSH config
```bash
# 编辑 ~/.ssh/config
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519
  IdentitiesOnly yes
```

## 安全建议

1. **保护私钥**
   - 不要分享私钥
   - 设置合适的文件权限：`chmod 600 ~/.ssh/id_rsa`
   - 使用密码保护私钥

2. **定期轮换**
   - 定期更新SSH Key
   - 移除不再使用的Key

3. **监控使用**
   - 定期检查GitHub的SSH Key列表
   - 关注异常登录活动

## 快速命令参考

```bash
# 初始化并推送
git init
git add .
git commit -m "Initial commit"
git remote add origin git@github.com:submlit21/stardewl-ink.git
git push -u origin main

# 如果遇到问题
git push -u origin main --force  # 强制推送（谨慎使用）

# 拉取更新
git pull origin main

# 查看状态
git status
git log --oneline -5
```

现在你的项目已经准备好推送到GitHub了！